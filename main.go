package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Reply is a simple struct wrapping the status and response text
type Reply struct {
	status int
	resp   string
}

const (
	ipifyHost = "https://api.ipify.org"
)

const (
	robot     string = "your call is"
	emptyCall string = robot + " [empty]"
	stdRet    string = "return status will be"
	resp200   string = stdRet + " 200"
	resp400   string = stdRet + " 400"
	resp500   string = stdRet + " 500"
)

func statusSwitch(statusDigit string) (reply *Reply) {
	switch {
	case len(statusDigit) == 0:
		return &Reply{http.StatusOK, emptyCall}
	case len(statusDigit) != 3:
		return &Reply{http.StatusOK, robot + " " + statusDigit}
	case statusDigit == "500":
		return &Reply{http.StatusInternalServerError, resp500}
	case statusDigit == "400":
		return &Reply{http.StatusBadRequest, resp400}
	default:
		return &Reply{http.StatusOK, resp200}
	}
}

type ResourceGetter interface {
	Get() *Reply
	API() string
}

type Resource struct {
	addr string
}

func (r Resource) API() string { return r.addr }
func (r Resource) Get() *Reply {
	resp, err := http.Get(r.addr)
	if err != nil {
		return &Reply{http.StatusBadGateway, err.Error()}
	}
	defer resp.Body.Close()

	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &Reply{http.StatusBadGateway, err.Error()}
	}

	return &Reply{http.StatusOK, string(text)}
}

func GetResource(getter ResourceGetter) *Reply {
	start := time.Now()
	reply := getter.Get()
	diff := time.Since(start)
	embedReply := Reply{reply.status, fmt.Sprintf(
		"<- (delay %v from %v) %v",
		diff.Round(time.Microsecond).String(),
		getter.API(),
		reply.resp,
	)}
	return &embedReply
}

func setupDefaults() *gin.Engine {
	r := gin.Default()

	r.NoRoute(func(c *gin.Context) {
		anything := c.Request.URL.Path
		reply := statusSwitch(anything[1:])
		c.String(reply.status, reply.resp)
	})
	return r
}

type Forwarder struct {
	route string
	ResourceGetter
}

func makeForwarder(route string, res ResourceGetter) *Forwarder {
	return &Forwarder{route, res}
}

func setupForwardings(r *gin.Engine, fwds []Forwarder) *gin.Engine {
	for _, fwd := range fwds {
		r.GET(fwd.route, func(c *gin.Context) {
			reply := GetResource(fwd)
			c.String(reply.status, reply.resp)
		})
	}
	return r
}

type Forwarders []Forwarder

func (f *Forwarders) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *Forwarders) Set(value string) error {
	tokens := strings.SplitN(value, ":", 2)
	if len(tokens) != 2 {
		return fmt.Errorf("wrong format, please check the usage")
	}
	if !strings.HasPrefix(tokens[0], "/") {
		return fmt.Errorf("route part should starts with '/'")
	}
	if !strings.HasPrefix(tokens[1], "http") {
		tokens[1] = "http://" + tokens[1]
	}
	_, err := url.ParseRequestURI(tokens[1])
	if err != nil {
		return fmt.Errorf("url parsing error (%v)", err)
	}
	u, err := url.Parse(tokens[1])
	if err != nil {
		return fmt.Errorf("url parsing error (%v)", err)
	}
	if len(u.Host) == 0 {
		return fmt.Errorf("not found Host part")
	}
	*f = append(*f, *makeForwarder(tokens[0], Resource{tokens[1]}))
	return nil
}

var (
	serverHost string
	serverPort int
	forwards   Forwarders
)

func getBuiltinForwarders() *Forwarders {
	var forwards Forwarders
	forwards = append(forwards, *makeForwarder("/getip", Resource{ipifyHost}))
	return &forwards
}

func init() {
	defaultHost := "" // listen and serve on 0.0.0.0:serverPort (for windows "localhost:serverPort")
	switch runtime.GOOS {
	case "windows":
		defaultHost = "localhost"
	default:
		defaultHost = "0.0.0.0"
	}

	builtinFwds := *getBuiltinForwarders()

	flag.IntVar(&serverPort, "p", 54321, "specify the port of echoes listen and serve")
	flag.StringVar(&serverHost, "H", defaultHost, "bind the host of echoes with address")
	flag.Var(&forwards, "fwd", "Forwardings pairs. format: [-fwd <route>:<URL> [-fwd <route>:<URL> ...]]\nExample: /foo:1.2.3.4:8080/bar")

	forwards = append(forwards, builtinFwds...)
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: go run main.go [options] [root]\n")
	flag.PrintDefaults()
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // https://stackoverflow.com/a/12122718

	flag.Parse()

	r := setupDefaults()
	r = setupForwardings(r, forwards)
	addr := fmt.Sprintf("%v:%v", serverHost, serverPort)
	r.Run(addr)
}
