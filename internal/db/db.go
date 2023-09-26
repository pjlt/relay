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

package db

import (
	"fmt"
	"relay/internal/conf"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var dbConn *gorm.DB

// 结构体'User'默认对应数据库表'users'
type User struct {
	ID       string
	Username string
	Key      string
}

func init() {
	db, err := gorm.Open(sqlite.Open(conf.Xml.DB.Path), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to open sqlite database(%s): %v", conf.Xml.DB.Path, err))
	}
	dbConn = db
}

func QueryByUserName(username string) (*User, error) {
	user := User{Username: username}
	result := dbConn.First(&user)
	if result.Error != nil {
		logrus.Errorf("Select table 'users' with {username:'%s'} failed with: %v", username, result.Error)
		return nil, result.Error
	}
	return &user, nil
}

func QueryUserList(index int) ([]User, error) {
	const kLimit int = 10
	var users []User
	result := dbConn.Limit(kLimit).Offset(index).Find(&users)
	if result.Error != nil {
		logrus.Errorf("Query table 'users' with limit(%d) offset(%d) failed with: %v", kLimit, index, result.Error)
		return nil, result.Error
	}
	return users, nil
}

func AddUser(username string, key string) error {
	user := User{
		Username: username,
		Key:      key,
	}
	result := dbConn.Create(&user)
	if result.Error != nil {
		logrus.Errorf("Insert record to table 'users' with {username:%s, key:%s} failed", username, key)
		return result.Error
	}
	return nil
}

func DelUser(username string) error {
	user := User{
		Username: username,
	}
	result := dbConn.Delete(&user)
	if result.Error != nil {
		logrus.Errorf("Delete record from table 'users' with {username:%s} failed", username)
		return result.Error
	}
	return nil
}
