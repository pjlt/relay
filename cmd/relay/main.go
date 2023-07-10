package main

import (
	"fmt"
	"relay/internal/app"
	"relay/internal/conf"
	"relay/internal/server"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var svr *server.Server

func initFunc() {
	svr = server.New(conf.Xml.Net.ListenIP, conf.Xml.Net.ListenPort)
	svr.Start()
}

func uninitFunc() {
	if svr != nil {
		svr.Stop()
		timer := time.NewTimer(svr.ReadTimeout() + time.Millisecond*10)
		select {
		case <-timer.C:
		case <-svr.StopedChan():
		}
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
	logrus.SetLevel(convertLogLevel(conf.Xml.Log.Level))
}

func convertLogLevel(level string) logrus.Level {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warning":
		fallthrough
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		panic(fmt.Sprintf("Unknown log level %s", level))
	}
}

func main() {
	initLogger()
	app.Run(initFunc, uninitFunc, dumpFunc)
}
