package tcp

import (
	"github.com/gogf/gf/os/glog"
	"health-check/prober/probe"
	"net"
	"time"
)

type tcpProbe struct{}

func New() probe.Probe {
	return tcpProbe{}
}

func (p tcpProbe) Ping(addr string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				glog.Warningf("%s client timeout,error:%s", addr, netErr.Error())
				return nil
			}
		}
		return err
	}
	if err = conn.Close(); err != nil {
		glog.Warningf("%s close fail,error:%s", addr, err.Error())
	}
	return nil

}
