/*
 * BSD 3-Clause License
 *
 * Copyright (c) 2023 Zhennan Tu <zhennan.tu@gmail.com>
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice, this
 *    list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * 3. Neither the name of the copyright holder nor the names of its
 *    contributors may be used to endorse or promote products derived from
 *    this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 * OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"bytes"
	"fmt"
	"path"
	"relay/internal/app"
	"relay/internal/conf"
	"relay/internal/mgr"
	"relay/internal/server"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var relaySvr *server.Server
var mgrSvr *mgr.Server

func initFunc() {
	relaySvr = server.New(conf.Xml.Net.ListenIP, conf.Xml.Net.ListenPort)
	relaySvr.Start()
	mgrSvr = mgr.New()
	mgrSvr.Start()
}

func uninitFunc() {
	if relaySvr != nil {
		relaySvr.Stop()
		timer := time.NewTimer(relaySvr.ReadTimeout() + time.Millisecond*10)
		select {
		case <-timer.C:
		case <-relaySvr.StopedChan():
		}
		relaySvr = nil
	}
}

func dumpFunc() {
	if relaySvr != nil {
		relaySvr.PrintStats()
	}
}

var levelList = []string{
	"PANIC",
	"FATAL",
	"ERROR",
	"WARN",
	"INFO",
	"DEBUG",
	"TRACE",
}

type theLogFormater struct{}

func (formater *theLogFormater) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	level := levelList[int(entry.Level)]
	strList := strings.Split(entry.Caller.File, "/")
	fileName := strList[len(strList)-1]
	b.WriteString(fmt.Sprintf("[%s][%s][%s:%d] %s\n",
		entry.Time.Format("2006/01/02 15:04:05.678"), level, fileName,
		entry.Caller.Line, entry.Message))
	return b.Bytes(), nil
}

func initLogger() {
	logger := &lumberjack.Logger{
		Filename: path.Join(conf.Xml.Log.Path, conf.Xml.Log.Prefix+".log"),
		MaxSize:  conf.Xml.Log.MaxSize,
		MaxAge:   conf.Xml.Log.MaxAge,
	}
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&theLogFormater{})
	logrus.SetOutput(logger)
	logrus.SetLevel(convertLogLevel(conf.Xml.Log.Level))
	logrus.Info("Log system initialized")
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
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		panic(fmt.Sprintf("Unknown log level %s", level))
	}
}

func main() {
	initLogger()
	app.Run(initFunc, uninitFunc, dumpFunc)
}
