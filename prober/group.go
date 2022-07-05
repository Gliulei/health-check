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
	Name          string
	Instances     map[string]instance
	FailInstances map[string]struct{}

	Cancel context.CancelFunc

	ProbeType string

	Probe probe.Probe

	Timeout time.Duration //超时时间

	Done chan struct{}

	Url string //http 接口

	UserAuth string //认证用户
	PassAuth string //认证密码

	Running bool

}

type instance struct {
	Ip               string
	Port             int
	Status           probe.Result
	Url              string
	OnRecoverHttpUrl string
	OnFailHttpUrl    string
}

func NewGroup(name string, probeType string, url string, cancel context.CancelFunc) Group {
	g := Group{
		Name:          name,
		Url:           url,
		Cancel:        cancel,
		Instances:     make(map[string]instance, 0),
		FailInstances: make(map[string]struct{}, 0),
		Done:          make(chan struct{}, 1),
	}
	if probeType == HttpProbe {
		g.Probe = httpProbe.New()
	} else {
		g.Probe = tcpProbe.New()
	}

	return g
}

func (g *Group) AddInstance(ip string, port int) error {
	var url string
	addr := fmt.Sprintf("%s:%d", ip, port)
	if g.ProbeType == HttpProbe {
		url = g.Url
	} else {
		url = addr
	}
	g.Instances[addr] = instance{Ip: ip, Port: port, Url: url, Status: probe.Success}

	return nil
}

func (g *Group) RemoveInstance(ip string, port int) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	delete(g.Instances, addr)

	return nil
}

func (g *Group) Run(ctx context.Context) {
	g.Running = true
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Warn("group %s is ended", g.Name)
			g.Running = false
			g.Done <- struct{}{}
			return
		case <-ticker.C:
			//not leader skip
			if atomic.LoadInt32(&global.IsLeader) == 0 {
				continue
			}
			for addr, instance := range g.Instances {
				log.Notice("leader addr is %s start probe %s", config.Cfg.Raft.Addr, instance.Url)
				err := g.Probe.Ping(instance.Url, g.Timeout)

				repeat := 3
				for repeat > 0 && err != nil {
					err = g.Probe.Ping(instance.Url, g.Timeout)
					repeat--
				}

				if err != nil {
					log.Warn("leader addr is %s probe fail %s", config.Cfg.Raft.Addr, instance.Url)
					if instance.Status == probe.Success {
						instance.Status = probe.Failure
						g.FailInstances[addr] = struct{}{}
						//计算此时失败的比列
						ratio := float64(len(g.FailInstances)) / float64(len(g.Instances))

						if instance.OnFailHttpUrl != "" {
							go func() {
								err = utils.HttpClient(g.Timeout).Post(instance.OnFailHttpUrl, map[string]interface{}{"addr": addr, "ratio": ratio})
								if err != nil {
									log.Warn("%s", err.Error())
								}
							}()
						}
						//instance.Notice(err)
					}
				} else {
					log.Notice("leader addr is %s probe success %s", config.Cfg.Raft.Addr, instance.Url)
					if instance.Status == probe.Failure {
						instance.Status = probe.Success
						//实例正常之后要从failInstances里删除
						delete(g.FailInstances, addr)

						if instance.OnRecoverHttpUrl != "" {
							go func() {
								err = utils.HttpClient(g.Timeout).Post(instance.OnRecoverHttpUrl, map[string]string{"addr": addr})
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
	addr := fmt.Sprintf("%s:%d", ins.Ip, ins.Port)
	switch ins.Status {
	case probe.Success:
		msg = fmt.Sprintf("%s 实例恢复正常", addr)
		break
	case probe.Failure:
		msg = fmt.Sprintf("%s 实例异常, error:%s", addr, err.Error())
		break
	}
	log.Warn(msg)
}
