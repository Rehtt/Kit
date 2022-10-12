package conf

import (
	"gopkg.in/ini.v1"
	"log"
)

var (
	Cfg *ini.File
)

func Init(filePath string) {
	var err error
	Cfg, err = ini.Load(filePath)
	if err != nil {
		log.Fatalln("读取配置文件失败：", err.Error())
	}
}