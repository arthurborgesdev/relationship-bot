package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
		return nil, err
	}

	err = db.AutoMigrate(&Message{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
		return nil, err
	}

	err = db.AutoMigrate(&Products{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
		return nil, err
	}

	return db, nil
}
