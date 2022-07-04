package utils

import (
	"github.com/gogf/gf/frame/g"
	"health-check/log"
	"net"
	"net/http"
	"time"
)

type Client struct {
	timeout time.Duration
}

func HttpClient(timeout time.Duration) *Client {
	return &Client{timeout: timeout}
}

func (c *Client) Post(url string, data interface{}) error {
	response, err := g.Client().Timeout(c.timeout).Post(url, data)

	if err != nil || response.StatusCode != http.StatusOK {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				log.Trace("%s client timeout,error:%s", url, netErr.Error())
				return nil
			}
		}

		return err
	}

	return nil
}
