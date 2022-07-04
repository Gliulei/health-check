package main

import (
	"context"
	"flag"
	"fmt"
	"health-check/config"
	"health-check/http_server"
	"health-check/library/global"
	"health-check/log"
	"health-check/log/xlog"
	craft "health-check/raft"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var configPath = flag.String("config", "etc/health-check.ini", "path to health-check config file")

func main() {
	flag.Parse()

	var err error
	config.Cfg, err = config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("load config error: %v\n", err.Error())
		return
	}

	if err := initXLog(&config.Cfg.Log); err != nil {
		fmt.Printf("init xlog error: %v\n", err.Error())
		return
	}

	r := config.Cfg.Raft
	myRaft, fm, err := craft.NewRaft(r.Addr, r.Id, r.GetPathDir())
	if err != nil {
		log.Warn("%s", err.Error())
		return
	}
	craft.Bootstrap(myRaft, r.Id, r.Addr, r.Nodes)

	// 监听leader变化（使用此方法无法保证强一致性读，仅做leader变化过程观察）
	httpServer := http_server.New(myRaft, fm)
	httpSrv := &http.Server{
		Addr:    config.Cfg.HttpAddr,
		Handler: httpServer.Engine,
	}
	go func() {
		for leader := range myRaft.LeaderCh() {
			if leader {
				atomic.StoreInt32(&global.IsLeader, 1)
			} else {
				atomic.StoreInt32(&global.IsLeader, 0)
			}
		}
	}()

	go func() {
		// 启动http server
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			// 关闭raft
			shutdownFuture := myRaft.Shutdown()
			if err := shutdownFuture.Error(); err != nil {
				log.Warn("%s", err.Error())
			}
			if err := httpSrv.Shutdown(ctx); err != nil {
				log.Warn("%s", err.Error())
			}
			cancel()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}

}

func initXLog(logConfig *config.Log) error {
	cfg := make(map[string]string)
	cfg["path"] = logConfig.Path
	cfg["filename"] = logConfig.Filename
	cfg["level"] = logConfig.Level
	cfg["service"] = logConfig.Service
	cfg["format"] = logConfig.Format
	cfg["skip"] = "5" // 设置xlog打印方法堆栈需要跳过的层数, 5目前为调用log.Debug()等方法的方法名, 比xlog默认值多一层.

	logger, err := xlog.CreateLogManager(logConfig.Output, cfg)
	if err != nil {
		return err
	}

	log.SetGlobalLogger(logger)
	return nil
}
