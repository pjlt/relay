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

package mgr

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"relay/internal/common"
	"relay/internal/conf"
	"relay/internal/db"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type responseStruct struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type sessionInfo struct {
	ReqIP     string `json:"req_ip"`
	ReqPort   uint16 `json:"req_port"`
	RespIP    string `json:"resp_ip"`
	RespPort  uint16 `json:"resp_port"`
	ReqToResp uint32 `json:"req_to_resp"`
	RespToReq uint32 `json:"resp_to_req"`
	StartTime int64  `json:"start"`
}

type statSessionData struct {
	Sessions []sessionInfo `json:"sessions"`
}

type userInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userListData struct {
	Users []userInfo `json:"users"`
}

type Server struct {
	router     *gin.Engine
	stopedChan chan struct{}
	httpSvr    *http.Server
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func toGinMode(mode string) string {
	mode = strings.ToLower(mode)
	if mode == gin.ReleaseMode || mode == gin.DebugMode || mode == gin.TestMode {
		return mode
	} else {
		logrus.Warnf("Unknown gin mode(%s), default to release mode", mode)
		return gin.ReleaseMode
	}
}

func New() *Server {
	gin.SetMode(toGinMode(conf.Xml.Mgr.Mode))
	return &Server{
		router:     gin.Default(),
		stopedChan: make(chan struct{}, 2),
	}
}

func (svr *Server) Start() {
	svr.router.GET("/user", svr.users)
	svr.router.GET("/stat", svr.stats)
	svr.router.POST("/user/add", svr.userAdd)
	svr.router.POST("/user/list", svr.userList)
	svr.router.POST("/user/del", svr.userDel)
	// svr.router.POST("/stat/total", svr.statTotal)
	// svr.router.POST("/stat/conns", svr.statSessions)
	svr.httpSvr = &http.Server{
		Addr:    conf.Xml.Mgr.ListenIP + ":" + fmt.Sprint(conf.Xml.Mgr.ListenPort),
		Handler: svr.router,
	}
	go func() {
		if err := svr.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("HTTP listen: %s", err)
		}
		svr.stopedChan <- struct{}{}
	}()
}

func (svr *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := svr.httpSvr.Shutdown(ctx); err != nil {
		logrus.Errorf("Shutdown http server: %s", err)
	}
	<-ctx.Done()
	logrus.Info("Mgr http server stoped.")
	svr.stopedChan <- struct{}{}
}

func (svr *Server) StopedChan() chan struct{} {
	return svr.stopedChan
}

func (svr *Server) users(ctx *gin.Context) {
	ctx.String(200, "users")
}

func (svr *Server) stats(ctx *gin.Context) {
	ctx.String(200, "stats")
}

func (svr *Server) userAdd(ctx *gin.Context) {
	username := ctx.PostForm("username")
	if username == "" {
		ctx.JSON(http.StatusOK, responseStruct{
			Status:  2,
			Message: "Invalid parameter",
		})
		return
	}
	password := common.RandStr(8)
	if len(username) > common.Fixed16 {
		username = username[:common.Fixed16]
	}
	err := db.AddUser(username, password)
	if err != nil {
		ctx.JSON(http.StatusOK, responseStruct{
			Status:  1,
			Message: "Insert database failed",
		})
		return
	}
	logrus.Infof("Add user(%s) success", username)
	userData := userInfo{
		Username: username,
		Password: password,
	}
	ctx.JSON(http.StatusOK, responseStruct{
		Status: 0,
		Data:   userData,
	})
}

func (svr *Server) userList(ctx *gin.Context) {
	index, err := strconv.Atoi(ctx.PostForm("index"))
	if err != nil {
		logrus.Info("userList parse index failed")
		return
	}
	users, err := db.QueryUserList(index)
	if err != nil {
		ctx.JSON(http.StatusOK, responseStruct{
			Status:  1,
			Message: "Query database failed",
		})
		return
	}
	var userData userListData
	for i := 0; i < len(users); i++ {
		userData.Users = append(userData.Users, userInfo{
			Username: users[i].Username,
			Password: users[i].Password,
		})
	}
	ctx.JSON(http.StatusOK, responseStruct{
		Status: 0,
		Data:   userData,
	})
}

func (svr *Server) userDel(ctx *gin.Context) {
	username := ctx.PostForm("username")
	if username == "" {
		ctx.JSON(http.StatusOK, responseStruct{
			Status:  2,
			Message: "Invalid parameter",
		})
		return
	}
	err := db.DelUser(username)
	if err != nil {
		ctx.JSON(http.StatusOK, responseStruct{
			Status:  1,
			Message: "Operate database failed",
		})
		return
	}
	ctx.JSON(http.StatusOK, responseStruct{
		Status: 0,
	})
}

// func (svr *Server) statTotal(ctx *gin.Context) {

// }

// func (svr *Server) statSessions(ctx *gin.Context) {

// }
