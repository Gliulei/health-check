package config

import (
	"github.com/go-ini/ini"
	"health-check/library/errors"
)

var Cfg *Config

type Config struct {
	OnRecoverHttpUrl string `ini:"on_recover_http_url"`
	OnFailHttpUrl    string `ini:"on_fail_http_url"`
	HttpAddr         string `ini:"http_addr"`
	Raft             `ini:"raft"`
	Log              `ini:"log"`
}

type Raft struct {
	Id      string `ini:"id"`
	Dir     string `ini:"dir"`
	Nodes string `ini:"nodes"`
	Addr    string `ini:"addr"`
}

type Log struct {
	Output   string `ini:"output"`
	Path     string `ini:"path"`
	Filename string `ini:"filename"`
	Level    string `ini:"level"`
	Service  string `ini:"service"`
	Format   string `ini:"format"`
}

func LoadConfig(configPath string) (*Config, error) {
	cfg, err := ini.ShadowLoad(configPath)
	if err != nil {
		return nil, err
	}
	rc := Config{}
	err = cfg.MapTo(&rc)
	if err != nil {
		return nil, err
	}
	err = checkRaftConfig(rc.Raft)

	if err != nil {
		return nil, err
	}

	return &rc, nil
}

func checkRaftConfig(r Raft) error {
	if r.Id == "" || r.Nodes == "" || r.Addr == "" || r.Dir == "" {
		return errors.ErrRaftConfigIsEmpty
	}

	return nil
}

func (r Raft) GetPathDir() string {
	return  r.Dir + r.Id
}
