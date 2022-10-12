
package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"{{.Mod}}/internal/conf"
)

var (
	RDB *gorm.DB
	WDB *gorm.DB
)

func Init() {
	log.Println("开始连接数据库")
	rdb, err := conf.Cfg.Section("").GetKey("rdb")
	if err != nil {
		log.Fatalln("配置文件缺少'rdb'字段")
	}
	RDB, err = gorm.Open(mysql.Open(rdb.String()), &gorm.Config{})
	if err != nil {
		log.Fatalln("rdb数据库连接失败：", err.Error())
	}

	wdb, err := conf.Cfg.Section("").GetKey("db")
	if err != nil {
		log.Fatalln("配置文件缺少'db'字段")
	}
	WDB, err = gorm.Open(mysql.Open(wdb.String()), &gorm.Config{})
	if err != nil {
		log.Fatalln("db数据库连接失败：", err.Error())
	}
	log.Println("数据库连接成功")
}
