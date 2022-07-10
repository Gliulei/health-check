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

	cancel context.CancelFunc

	ProbeType string

	probe probe.Probe

	Timeout time.Duration //超时时间

	done chan struct{}

	ProbeUrl string //http 接口

	UserAuth string //认证用户
	PassAuth string //认证密码

	running bool
}

type instance struct {
	Ip               string
	Port             int
	Status           probe.Result
	ProbeAddr        string
	OnRecoverHttpUrl string
	OnFailHttpUrl    string
}

func NewGroup(name string, probeType string, url string, cancel context.CancelFunc) *Group {
	g := Group{
		Name:          name,
		ProbeUrl:      url,
		cancel:        cancel,
		Instances:     make(map[string]instance, 0),
		FailInstances: make(map[string]struct{}, 0),
		done:          make(chan struct{}, 1),
	}
	if probeType == HttpProbe {
		g.probe = httpProbe.New()
	} else {
		g.probe = tcpProbe.New()
	}

	return &g
}

func (g *Group) InitGroup(cancelFunc context.CancelFunc) {
	g.cancel = cancelFunc
	g.done = make(chan struct{}, 1)
	if g.ProbeType == HttpProbe {
		g.probe = httpProbe.New()
	} else {
		g.probe = tcpProbe.New()
	}

}

func (g *Group) AddInstance(ip string, port int) error {
	var probeAddr string
	addr := fmt.Sprintf("%s:%d", ip, port)
	if g.ProbeType == HttpProbe {
		probeAddr = g.ProbeUrl
	} else {
		probeAddr = addr
	}
	g.Instances[addr] = instance{Ip: ip, Port: port, ProbeAddr: probeAddr, Status: probe.Success}

	return nil
}

func (g *Group) RemoveInstance(ip string, port int) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	delete(g.Instances, addr)
	delete(g.FailInstances, addr)

	return nil
}

func (g *Group) Run(ctx context.Context) {
	g.running = true
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Warn("group %s is be cancelled", g.Name)
			g.running = false
			g.done <- struct{}{}
			return
		case <-ticker.C:
			//not leader skip
			if atomic.LoadInt32(&global.IsLeader) == 0 {
				continue
			}
			for addr, instance := range g.Instances {
				log.Notice("leader addr is %s start probe %s", config.Cfg.Raft.Addr, instance.ProbeAddr)
				err := g.probe.Ping(instance.ProbeAddr, g.Timeout)

				repeat := 3
				for repeat > 0 && err != nil {
					err = g.probe.Ping(instance.ProbeAddr, g.Timeout)
					repeat--
				}

				if err != nil {
					log.Warn("leader addr is %s probe fail %s", config.Cfg.Raft.Addr, instance.ProbeAddr)
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
					log.Notice("leader addr is %s probe success %s", config.Cfg.Raft.Addr, instance.ProbeAddr)
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
