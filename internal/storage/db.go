package storage

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"

	"github.com/yourusername/golang-job-scheduler/internal/config"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) {
	// Just in case you want to automatically migrate
	// We'll call db.AutoMigrate on the Job model
	db.AutoMigrate(&JobModel{})
}

// We'll define a minimal struct for migration, or you can do direct migrations
type JobModel struct {
	ID         string `gorm:"primary_key"`
	Payload    string
	Status     string
	Result     string
	Retries    int
	MaxRetries int
	CreatedAt  string
	UpdatedAt  string
}
