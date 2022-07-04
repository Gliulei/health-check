package prober

import (
	"context"
	"fmt"
	"health-check/config"
	"health-check/library/global"
	"health-check/library/utils"
	"health-check/log"
	"health-check/prober/probe"
	httpProbe "health-check/prober/probe/http"
	tcpProbe "health-check/prober/probe/tcp"
	"sync/atomic"
	"time"
)

const (
	TcpProbe  = "tcp"
	HttpProbe = "http"
)

type Group struct {
	name          string
	instances     map[string]instance
	failInstances map[string]struct{}

	cancel context.CancelFunc

	probeType string

	probe probe.Probe

	timeout time.Duration //超时时间

	done chan struct{}

	url string //http 接口

	userAuth string //认证用户
	passAuth string //认证密码

}

type instance struct {
	ip               string
	port             int
	status           probe.Result
	url              string
	onRecoverHttpUrl string
	onFailHttpUrl    string
}

func NewGroup(name string, probeType string, url string, cancel context.CancelFunc) Group {
	g := Group{
		name:          name,
		url:           url,
		cancel:        cancel,
		instances:     make(map[string]instance, 0),
		failInstances: make(map[string]struct{}, 0),
		done:          make(chan struct{}, 1),
	}
	if probeType == HttpProbe {
		g.probe = httpProbe.New()
	} else {
		g.probe = tcpProbe.New()
	}

	return g
}

func (g *Group) AddInstance(ip string, port int) error {
	var url string
	addr := fmt.Sprintf("%s:%d", ip, port)
	if g.probeType == HttpProbe {
		url = g.url
	} else {
		url = addr
	}
	g.instances[addr] = instance{ip: ip, port: port, url: url, status: probe.Success}

	return nil
}

func (g *Group) RemoveInstance(ip string, port int) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	delete(g.instances, addr)

	return nil
}

func (g *Group) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Warn("group %s is ended", g.name)
			g.done <- struct{}{}
			return
		case <-ticker.C:
			//not leader skip
			if atomic.LoadInt32(&global.IsLeader) == 0 {
				continue
			}
			for addr, instance := range g.instances {
				log.Notice("leader addr is %s start probe %s", config.Cfg.Raft.Addr, instance.url)
				err := g.probe.Ping(instance.url, g.timeout)

				repeat := 3
				for repeat > 0 && err != nil {
					err = g.probe.Ping(instance.url, g.timeout)
					repeat--
				}

				if err != nil {
					log.Warn("leader addr is %s probe fail %s", config.Cfg.Raft.Addr, instance.url)
					if instance.status == probe.Success {
						instance.status = probe.Failure
						g.failInstances[addr] = struct{}{}
						//计算此时失败的比列
						ratio := float64(len(g.failInstances)) / float64(len(g.instances))

						if instance.onFailHttpUrl != "" {
							go func() {
								err = utils.HttpClient(g.timeout).Post(instance.onFailHttpUrl, map[string]interface{}{"addr": addr, "ratio": ratio})
								if err != nil {
									log.Warn("%s", err.Error())
								}
							}()
						}
						//instance.Notice(err)
					}
				} else {
					log.Notice("leader addr is %s probe success %s", config.Cfg.Raft.Addr, instance.url)
					if instance.status == probe.Failure {
						instance.status = probe.Success
						//实例正常之后要从failInstances里删除
						delete(g.failInstances, addr)

						if instance.onRecoverHttpUrl != "" {
							go func() {
								err = utils.HttpClient(g.timeout).Post(instance.onRecoverHttpUrl, map[string]string{"addr": addr})
								if err != nil {
									log.Warn("%s", err.Error())
								}
							}()
						}

						//instance.Notice(nil)
					}
				}
			}
		}
	}
}

func (ins *instance) Notice(err error) {
	var msg string
	addr := fmt.Sprintf("%s:%d", ins.ip, ins.port)
	switch ins.status {
	case probe.Success:
		msg = fmt.Sprintf("%s 实例恢复正常", addr)
		break
	case probe.Failure:
		msg = fmt.Sprintf("%s 实例异常, error:%s", addr, err.Error())
		break
	}
	log.Warn(msg)
}
