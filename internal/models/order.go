package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CustomerID      uuid.UUID `gorm:"not null"`
	Status          string    `gorm:"type:varchar(50);not null;check:status IN ('pending', 'approved', 'paid', 'shipped', 'completed', 'cancelled')"`
	PrescriptionURL *string   `gorm:"type:text"`
	ShippingCost    float64   `gorm:"type:numeric(10,2);default:0.00"`
	Subtotal        float64   `gorm:"type:numeric(10,2);default:0.00"`
	CreatedAt       time.Time `gorm:"type:timestamptz;default:now()"`
	UpdatedAt       time.Time `gorm:"type:timestamptz;default:now()"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = uuid.New()
	return
}
