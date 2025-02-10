package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderItem struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrderID     uuid.UUID `gorm:"not null"`
	ProductID   uuid.UUID `gorm:"not null"`
	ProductName string    `gorm:"not null"`
	Quantity    int       `gorm:"not null;check:quantity > 0"`
	Price       float64   `gorm:"not null"`
	CreatedAt   time.Time `gorm:"default:now()"`
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	oi.ID = uuid.New()
	return
}
