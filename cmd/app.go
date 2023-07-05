package cmd

import (
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

// Run 执行一个非阻塞函数，然后自己进入永久性的wait中，
// 直到捕获到SIGTERM、SIGINT
func Run(init func(), uninit func(), dump func()) {
	defer func() {
		log.Println(string(debug.Stack()))
	}()
	if init != nil {
		init()
	}
	quit := make(chan os.Signal, 5)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			if dump != nil {
				dump()
			}
		case <-quit:
			if uninit != nil {
				uninit()
				return
			}
		}
	}
}
