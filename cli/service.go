package main

import (
	"context"
	"github.com/pkg/errors"
	"github.com/ruixiaoedu/ota/core"
	"github.com/ruixiaoedu/ota/unixsocket"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// service 启动服务
func service(c *core.Core) {

	g, ctx := errgroup.WithContext(context.Background())

	// 初始化unix socket
	us := unixsocket.NewService(c)
	g.Go(func() error {
		return us.Server()
	})

	// 监听退出命令
	g.Go(func() error {
		quit := make(chan os.Signal, 0)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-quit:
			return errors.Errorf("get os signal: %v", sig)
		}
	})

	// 报告错误
	log.Fatalf("error exiting: %+v\n", g.Wait())
}
