package data

import (
	"SecKill/conf"
	"SecKill/model"
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var Db *gorm.DB

func initMysql(config conf.AppConfig) {
	log.Println("load dbService config...")

	dbType := config.App.Database.Type
	username := config.App.Database.User
	password := config.App.Database.Password
	address := config.App.Database.Address
	dbName := config.App.Database.DbName
	dbLink := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		username, password, address, dbName)

	log.Println("Init dbService connections...")
	var err error
	for Db, err = gorm.Open(dbType, dbLink); err != nil; Db, err = gorm.Open(dbType, dbLink) {
		log.Println("Failed to connect database: ", err.Error())
		log.Println("Reconnecting in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
	// set open connections number
	Db.DB().SetMaxOpenConns(config.App.Database.MaxOpen)
	Db.DB().SetMaxIdleConns(config.App.Database.MaxIdle)

	user := &model.User{}
	coupon := &model.Coupon{}
	tables := []interface{}{user, coupon}

	if config.App.FlushAllForTest {
		log.Println("FlushAllForTest is true. Delete records of all tables.")
		for _, table := range tables {
			if err := Db.DropTableIfExists(table).Error; err != nil {
				log.Fatal("Error dropping table:", err)
			}
		}
	}

	for _, table := range tables {
		if !Db.HasTable(table) {
			Db.AutoMigrate(table)
		}
	}

	// create unique index
	Db.Model(user).AddUniqueIndex("username_index", "username")
	Db.Model(coupon).AddUniqueIndex("coupon_index", "username", "coupon_name")

	log.Println("---Mysql connection is initialized.---")

	// Db.Model(credit_card).
	//	 AddForeignKey("owner_id", "users(id)", "RESTRICT", "RESTRICT").
	//	 AddUniqueIndex("unique_owner", "owner_id")
}
