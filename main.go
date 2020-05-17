package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

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

type PublicIPServer struct {
	apihost string
}

func (p PublicIPServer) GetMyPublicIP() (string, error) {
	resp, err := http.Get(p.apihost)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(text), nil
}

// SetupEndpoints is setting all needed endpoints for our gin sever
func SetupEndpoints() *gin.Engine {
	r := gin.Default()
	r.GET("/getip", func(c *gin.Context) {
		server := PublicIPServer{ipifyHost}
		result, err := server.GetMyPublicIP()
		if err != nil {
			c.String(http.StatusBadGateway, err.Error())
		}
		c.String(200, result)
	})

	r.NoRoute(func(c *gin.Context) {
		anything := c.Request.URL.Path
		reply := statusSwitch(anything[1:])
		c.String(reply.status, reply.resp)
	})
	return r
}

var (
	serverHost string
	serverPort int
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
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: go run main.go [options] [root]\n")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	r := SetupEndpoints()
	addr := fmt.Sprintf("%v:%v", serverHost, serverPort)
	r.Run(addr)
}
