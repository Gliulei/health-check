package utils

import (
	"fmt"
	"health-check/log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

type HttpProxy struct {
	proxy *httputil.ReverseProxy
}

func NewHttpProxy(targetHost string) *HttpProxy {
	p, err := newProxy(targetHost)
	if err != nil {
		log.Warn(err.Error())
		return nil
	}
	return &HttpProxy{
		proxy: p,
	}
}

func (h *HttpProxy) Request(w http.ResponseWriter, r *http.Request) {
	h.proxy.ServeHTTP(w, r)
}

// NewProxy takes target host and creates a reverse proxy
func newProxy(targetHost string) (*httputil.ReverseProxy, error) {

	addr := strings.Split(targetHost, ":")
	port, _ := strconv.Atoi(addr[1])
	port = port + 1
	targetHost = fmt.Sprintf("%s:%d", addr[0], port)

	URL := &url.URL{}
	URL.Scheme = "http"

	URL.Host = targetHost

	fmt.Println("proxy test url")
	fmt.Println(URL.String())
	proxy := httputil.NewSingleHostReverseProxy(URL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		modifyRequest(req)
	}

	proxy.ModifyResponse = modifyResponse()
	proxy.ErrorHandler = errorHandler()
	return proxy, nil
}

func modifyRequest(req *http.Request) {
	req.Header.Set("X-Proxy", "Simple-Reverse-Proxy")
}

func errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Got error while modifying response: %v \n", err)
		return
	}
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		return nil
	}
}