package dao

import (
	"fmt"

	"github.com/Misaka13906/FantasyGoWorld/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	connectMysql()
	migrateMysql()
}

func connectMysql() {
	var err error
	dbConfig := config.Configs.MysqlDb
	DB, err = gorm.Open(mysql.Open(fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?%s",
		dbConfig.User,
		dbConfig.Pswd,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
		"charset=utf8mb4&parseTime=True&loc=Local",
	)), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to MySQL: %v", err))
	}
}

func migrateMysql() {
	err := DB.AutoMigrate(User{})
	if err != nil {
		panic(fmt.Sprintf("Failed to run MySQL migrations: %v", err))
	}
	fmt.Println("MySQL migrations completed successfully.")
}
