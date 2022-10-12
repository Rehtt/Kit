package service

import (
	"flag"
	"{{.Mod}}/internal/conf"
	"{{.Mod}}/internal/database"
)

var (
	confPath = flag.String("conf","conf.ini","配置文件地址")
)

func init() {
	flag.Parse()
	conf.Init(*confPath)
	database.Init()

}

func Run() {

}
