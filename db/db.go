package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/onionwyl/smart-light/model"
	"log"
)

var db *gorm.DB

func Init() bool {
	var err error
	db, err = gorm.Open("mysql", "root:wangyulin@/light?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatalf("connect db error: %v", err)
		return false
	}
	db.SingularTable(true)
	if !db.HasTable(&model.Actions{}) {
		db.CreateTable(&model.Actions{})
	}
	if !db.HasTable(&model.EnvData{}) {
		db.CreateTable(&model.EnvData{})
	}
	return true
}

func GetDBConn() *gorm.DB {
	return db
}
