package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/zverobot1k/threat-service-blokhin/internal/models"
)

func OpenDatabase(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err = db.AutoMigrate(&models.Threat{}); err != nil {
		return nil, err
	}
	return db, nil
}
