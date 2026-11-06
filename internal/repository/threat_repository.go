package repository

import (
	"context"
	"errors"

	"github.com/zverobot1k/threat-service-blokhin/internal/models"
	"gorm.io/gorm"
)

var ErrNotFound = gorm.ErrRecordNotFound

type ThreatRepository interface {
	Create(ctx context.Context, threat *models.Threat) error
	FindByMD5(ctx context.Context, md5 string) (*models.Threat, error)
	Update(ctx context.Context, threat *models.Threat) error
}

type GORMThreatRepository struct {
	db *gorm.DB
}

func NewGORMThreatRepository(db *gorm.DB) *GORMThreatRepository {
	return &GORMThreatRepository{db: db}
}

func (r *GORMThreatRepository) Create(ctx context.Context, threat *models.Threat) error {
	return r.db.WithContext(ctx).Create(threat).Error
}

func (r *GORMThreatRepository) FindByMD5(ctx context.Context, md5 string) (*models.Threat, error) {
	var threat models.Threat
	if err := r.db.WithContext(ctx).First(&threat, "md5 = ?", md5).Error; err != nil {
		return nil, err
	}
	return &threat, nil
}

func (r *GORMThreatRepository) Update(ctx context.Context, threat *models.Threat) error {
	result := r.db.WithContext(ctx).Save(threat)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsConflict(err error) bool {
	var stateErr interface{ SQLState() string }
	return errors.As(err, &stateErr) && stateErr.SQLState() == "23505"
}
