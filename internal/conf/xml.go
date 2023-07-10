package conf

import (
	"encoding/xml"
	"flag"
	"io/ioutil"
)

const defaultXmlPath = "/path/to/relay.xml"

var Xml relayConf

type relayConf struct {
	Log logConf `xml:"log"`
	Net netConf `xml:"net"`
	DB  dbConf  `xml:"db"`
}

type logConf struct {
	Path   string `xml:"path"`
	Prefix string `xml:"prefix"`
	Level  string `xml:"level"`
}

type netConf struct {
	ListenPort uint16 `xml:"port"`
	ListenIP   string `xml:"ip"`
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
	content, err := ioutil.ReadFile(xmlPath)
	if err != nil {
		return err
	}
	cfg := relayConf{}
	err = xml.Unmarshal(content, &cfg)
	if err != nil {
		return err
	}
	Xml = cfg
	return nil
}
