package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"soc-mvp-distribuido/events/internal/model"
)

type GormEventRepository struct {
	db *gorm.DB
}

func NewGormEventRepository(db *gorm.DB) *GormEventRepository {
	return &GormEventRepository{db: db}
}

var _ EventRepository = (*GormEventRepository)(nil)

func (r *GormEventRepository) Create(ctx context.Context, eventType, severity string) (model.Event, error) {
	event := model.Event{
		Type:     eventType,
		Severity: severity,
	}

	if err := r.db.WithContext(ctx).Create(&event).Error; err != nil {
		return model.Event{}, err
	}

	return event, nil
}

func (r *GormEventRepository) List(ctx context.Context) ([]model.Event, error) {
	var events []model.Event
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *GormEventRepository) Get(ctx context.Context, id uint) (model.Event, error) {
	var event model.Event

	err := r.db.WithContext(ctx).First(&event, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Event{}, ErrNotFound
	}
	if err != nil {
		return model.Event{}, err
	}

	return event, nil
}
