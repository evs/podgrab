package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

//Base is
type Base struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

//BeforeCreate
func (base *Base) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("ID", uuid.New().String())
	return nil
}
