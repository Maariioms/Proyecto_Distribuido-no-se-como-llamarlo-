package repository

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"soc-mvp-distribuido/backend/internal/model"
)

type GormAccountRepository struct {
	db *gorm.DB
}

func NewGormAccountRepository(db *gorm.DB) *GormAccountRepository {
	return &GormAccountRepository{db: db}
}

var _ AccountRepository = (*GormAccountRepository)(nil)

func (r *GormAccountRepository) Create(ctx context.Context, username, password string) (model.Account, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.Account{}, err
	}

	account := model.Account{
		Username:     username,
		PasswordHash: string(hash),
	}

	if err := r.db.WithContext(ctx).Create(&account).Error; err != nil {
		return model.Account{}, err
	}

	return account, nil
}

func (r *GormAccountRepository) Authenticate(ctx context.Context, username, password string) (model.Account, error) {
	var account model.Account

	err := r.db.WithContext(ctx).Where("username = ?", username).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Account{}, ErrInvalidCredentials
	}
	if err != nil {
		return model.Account{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		return model.Account{}, ErrInvalidCredentials
	}

	return account, nil
}
