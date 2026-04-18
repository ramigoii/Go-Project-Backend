package config

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "host=postgres user=postgres password=postgres dbname=movie_db port=5432 sslmode=disable"
	log.Println("DSN:", dsn)

	var database *gorm.DB
	var err error

	for i := 0; i < 5; i++ {
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("✅ Connected to database")
			break
		}

		log.Println("⏳ Waiting for DB...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("❌ Failed to connect:", err)
	}

	DB = database
}
