package database

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Deleted   bool      `gorm:"not null;default:false" json:"deleted"`
	Name      string    `gorm:"type:varchar(250);not null" json:"name"`
	Brand     string    `gorm:"type:varchar(250);not null" json:"brand"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	State     string    `gorm:"type:device_state;not null;default:'Available'" json:"state"`
}
