package http_server

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"health-check/config"
	"health-check/library"
	"health-check/library/global"
	"health-check/library/utils"
	"health-check/prober"
	craft "health-check/raft"
	"time"
)

type HttpServer struct {
	*gin.Engine
	raft *raft.Raft
	fsm  *craft.Fsm
}

func New(raft *raft.Raft, fsm *craft.Fsm) *HttpServer {
	r := gin.Default()
	s := &HttpServer{
		Engine: r,
		raft:   raft,
		fsm:    fsm,
	}

	s.registerURL()
	s.Engine = r
	return s
}

func (s *HttpServer) registerURL() {
	group := s.Engine.Group("api/instance", gin.BasicAuth(gin.Accounts{config.Cfg.AuthUser: config.Cfg.AuthPassword}))
	group.POST("add", s.Add)
	group.DELETE("remove", s.Remove)
	group.GET("lists", s.GetList)
}

func (s *HttpServer) Add(c *gin.Context) {
	if s.raft.State() != raft.Leader {
		addr, _ := s.raft.LeaderWithID()
		utils.NewHttpProxy(string(addr)).Request(c.Writer, c.Request)
		return
	}
	var g library.GroupForm
	if err := c.Bind(&g); err != nil {
		global.FailWithMessage(c, err.Error())
		return
	}
	data, err := json.Marshal(&craft.StoreCommand{Op: "add", Value: g})
	if err != nil {
		global.FailWithMessage(c, err.Error())
		return
	}

	future := s.raft.Apply(data, 5*time.Second)
	if err := future.Error(); err != nil {
		global.FailWithMessage(c, err.Error())
		return
	}
	global.Ok(c)
}

func (s *HttpServer) Remove(c *gin.Context) {
	if s.raft.State() != raft.Leader {
		addr, _ := s.raft.LeaderWithID()
		utils.NewHttpProxy(string(addr)).Request(c.Writer, c.Request)
		return
	}
	var g library.GroupForm
	if err := c.Bind(&g); err != nil {
		global.FailWithMessage(c, err.Error())
		return
	}

	data, err := json.Marshal(&craft.StoreCommand{Op: "remove", Value: g})
	if err != nil {
		global.FailWithMessage(c, err.Error())
		return
	}

	future := s.raft.Apply(data, 5*time.Second)
	if err := future.Error(); err != nil {
		global.FailWithMessage(c, err.Error())
		return
	}

	global.Ok(c)
}

func (s *HttpServer) GetList(c *gin.Context) {
	var data interface{}
	var err error
	groupName := c.Query("groupName")
	if len(groupName) > 0 {
		//data, err = prober.Manager.GetInstance(groupName)
		data, err = prober.Manager.GetInstanceOne(groupName)
		if err != nil {
			global.OkWithMessage(c, err.Error())
			return
		}
	}
	global.OkWithData(c, data)
}
