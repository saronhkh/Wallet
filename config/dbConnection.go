package config

import (
	models "Wallet/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("faild to open file", err)
	}
	dbhost := os.Getenv("MYSQL_HOST")
	dbuser := os.Getenv("MYSQL_USER")
	dbpassword := os.Getenv("MYSQL_PASSWORD")
	dbname := os.Getenv("MYSQL_DBNAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbuser, dbpassword, dbhost, dbname)
	var db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("faild to open file", err)
	}

	DB = db
	fmt.Println("db connect successfuly")

	AutoMigrate(db)
}

func AutoMigrate(connection *gorm.DB) {
	connection.Debug().AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Transaction{},
	)
}
