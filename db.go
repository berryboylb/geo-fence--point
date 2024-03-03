package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var dbDriver *gorm.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	//get db creds
	db_user := os.Getenv("DB_USER")
	db_password := os.Getenv("DB_PASSWORD")
	db_name := os.Getenv("DB_NAME")
	db_host := os.Getenv("DB_HOST")
	db_port_string := os.Getenv("DB_PORT")

	//check if any is missing
	if db_port_string == "" ||  db_user == "" || db_password == ""|| db_name == "" || db_host == ""{
		log.Fatal("Error loading db port env variables")
	}

	//convert the port string to integer
	db_port, err := strconv.Atoi(db_port_string)
	if err != nil {
		log.Fatal("Error convert db port env variable to integer")
	}

	//create connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", db_host, db_port, db_user, db_password, db_name)
	dbDriver, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Disables implicit prepared statement usage
	}), &gorm.Config{ TranslateError: true })
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
}

func GetDB() *gorm.DB {
	return dbDriver
}
