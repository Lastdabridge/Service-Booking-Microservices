package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetUpDatabaseConnection() *gorm.DB {

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" || appEnv == "local" {
		if err := godotenv.Load(".env"); err != nil {
			log.Println(".env file not found")
		}
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Fatal("dbUser пуст")
	}

	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		log.Fatal("dbPass пуст")
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		log.Fatal("dbHost пуст")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("dbName пуст")
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		log.Fatal("dbPort пуст")
	}

	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v", dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	return db
}
