package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"soc-mvp-distribuido/backend/internal/model"
)

func Connect(dsn string) (*gorm.DB, error) {
	// sin TranslateError gorm no convierte los errores de postgres a los suyos
	// y un username repetido nunca llegaba como gorm.ErrDuplicatedKey
	// salia como 500 en vez de 409
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	if err := db.AutoMigrate(&model.Account{}); err != nil {
		return nil, fmt.Errorf("running auto-migration: %w", err)
	}

	return db, nil
}
