package http

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"health-check/prober/probe"
	"net"
	"net/http"
	"time"
)


type httpProbe struct {
	userAuth string
	passAuth string
}

func New() probe.Probe {
	return httpProbe{}
}

func (p httpProbe) Ping(url string, timeout time.Duration) error {
	response, err := g.Client().BasicAuth(p.userAuth, p.passAuth).Timeout(timeout).Get(url)

	if err != nil || response.StatusCode != http.StatusOK {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				glog.Warningf("%s client timeout,error:%s", url, netErr.Error())
				return nil
			}
		}

		return err
	}

	return nil

}
