package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"soc-mvp-distribuido/events/internal/model"
)

func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	if err := db.AutoMigrate(&model.Event{}); err != nil {
		return nil, fmt.Errorf("running auto-migration: %w", err)
	}

	return db, nil
}
