package database

import (
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(user, password, dbname, port string) (db *gorm.DB, close func() error) {
	return NewWithHost(user, password, dbname, port, "localhost")
}

func NewWithHost(user, password, dbname, port, host string) (db *gorm.DB, close func() error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %s", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %s", err)
	}

	return db, sqlDB.Close
}
