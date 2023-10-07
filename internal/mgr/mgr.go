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
	"fmt"
	"math/rand"
	"net/http"
	"relay/internal/conf"
	"relay/internal/db"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

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
	router *gin.Engine
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func New() *Server {
	return &Server{
		router: gin.Default(),
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
	go svr.router.Run(conf.Xml.Mgr.ListenIP + ":" + fmt.Sprint(conf.Xml.Mgr.ListenPort))
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
	key := randSeq(8)
	err := db.AddUser(username, key)
	if err != nil {
		ctx.JSON(http.StatusOK, responseStruct{
			Status:  1,
			Message: "Insert database failed",
		})
		return
	}
	logrus.Infof("Add user{username:%s, key:%s} success", username, key)
	ctx.JSON(http.StatusOK, responseStruct{
		Status: 0,
	})
}

func (svr *Server) userList(ctx *gin.Context) {
	logrus.Info("userList")
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
