package main

import (
	"relay/cmd"
	"relay/internal/server"

	"github.com/sirupsen/logrus"
)

var svr *server.Server

func initFunc() {
	svr = server.New()
	svr.Start()
}

func uninitFunc() {
	if svr != nil {
		svr.Stop()
		svr = nil
	}
}

func dumpFunc() {
	if svr != nil {
		svr.PrintStats()
	}
}

func initLogger() {
	//logrus.SetFormatter(&logrus.TextFormatter{})
	//logrus.SetOutput()
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	initLogger()
	cmd.Run(initFunc, uninitFunc, dumpFunc)
}
