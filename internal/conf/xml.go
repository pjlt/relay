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

package conf

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
)

const defaultXmlPath = "/path/to/relay.xml"
const defaultXmlConfig = `
<?xml version="1.0" encoding="UTF-8" ?>
<relay>

    <log>
        <path>log</path>
        <prefix>relay</prefix>
        <level>info</level>
		<maxsize>10</maxsize>
        <maxage>30</maxage>
    </log>

    <net>
        <ip>0.0.0.0</ip>
        <port>19000</port>
    </net>

    <mgr>
        <ip>0.0.0.0</ip>
        <port>19001</port>
		<mode>release</mode>
    </mgr>

    <db>
        <path>user.db</path>
    </db>

</relay>
`

var Xml relayConf

type relayConf struct {
	Log logConf `xml:"log"`
	Net netConf `xml:"net"`
	Mgr mgrConf `xml:"mgr"`
	DB  dbConf  `xml:"db"`
}

type logConf struct {
	Path    string `xml:"path"`
	Prefix  string `xml:"prefix"`
	Level   string `xml:"level"`
	MaxSize int    `xml:"maxsize"`
	MaxAge  int    `xml:"maxage"`
}

type netConf struct {
	ListenPort uint16 `xml:"port"`
	ListenIP   string `xml:"ip"`
}

type mgrConf struct {
	ListenPort uint16 `xml:"port"`
	ListenIP   string `xml:"ip"`
	Mode       string `xml:"mode"`
}

type dbConf struct {
	Path string `xml:"path"`
}

func init() {
	xmlPath := flag.String("c", defaultXmlPath, "配置文件路径")
	flag.Parse()
	if err := loadConfig(*xmlPath); err != nil {
		panic(err)
	}
}

func loadConfig(xmlPath string) error {
	content, err := os.ReadFile(xmlPath)
	if err != nil {
		fmt.Printf("Read config from '%s' failed, using default config.\n\n", xmlPath)
		content = []byte(defaultXmlConfig)
	}
	cfg := relayConf{}
	err = xml.Unmarshal(content, &cfg)
	if err != nil {
		return err
	}
	Xml = cfg
	return nil
}
