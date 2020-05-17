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

type Resource struct {
	addr string
}

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

// SetupEndpoints is setting all needed endpoints for our gin sever
func SetupEndpoints() *gin.Engine {
	r := gin.Default()
	r.GET("/getip", func(c *gin.Context) {
		server := Resource{ipifyHost}
		reply := server.Get()
		c.String(reply.status, reply.resp)
	})

	r.NoRoute(func(c *gin.Context) {
		anything := c.Request.URL.Path
		reply := statusSwitch(anything[1:])
		c.String(reply.status, reply.resp)
	})
	return r
}

type Forwarder struct {
	route      string
	forwardAPI string
}

func SetupCustomForwardings(r *gin.Engine, fwds []Forwarder) *gin.Engine {
	for _, fwd := range fwds {
		r.GET(fwd.route, func(c *gin.Context) {
			server := Resource{fwd.forwardAPI}
			reply := server.Get()
			fwdReply := Reply{reply.status, fmt.Sprintf("<- (from %v) %v", fwd.forwardAPI, reply.resp)}
			c.String(fwdReply.status, fwdReply.resp)
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
	*f = append(*f, Forwarder{tokens[0], tokens[1]})
	return nil
}

var (
	serverHost string
	serverPort int
	forwards   Forwarders
)

func init() {
	defaultHost := "" // listen and serve on 0.0.0.0:serverPort (for windows "localhost:serverPort")
	switch runtime.GOOS {
	case "windows":
		defaultHost = "localhost"
	default:
		defaultHost = "0.0.0.0"
	}

	flag.IntVar(&serverPort, "p", 54321, "specify the port of echoes listen and serve")
	flag.StringVar(&serverHost, "H", defaultHost, "bind the host of echoes with address")
	flag.Var(&forwards, "fwd", "Forwardings pairs. format: [-fwd <route>:<URL> [-fwd <route>:<URL> ...]]\nExample: /foo:1.2.3.4:8080/bar")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: go run main.go [options] [root]\n")
	flag.PrintDefaults()
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // https://stackoverflow.com/a/12122718

	flag.Parse()

	r := SetupEndpoints()
	r = SetupCustomForwardings(r, forwards)
	addr := fmt.Sprintf("%v:%v", serverHost, serverPort)
	r.Run(addr)
}
