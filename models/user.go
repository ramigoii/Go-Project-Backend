package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model `json:"-"`
	ID         uint      `gorm:"primaryKey" json:"id"`
	Username   string    `gorm:"unique;not null;size:100" json:"username" binding:"required"`
	Email      string    `gorm:"unique;not null;size:255" json:"email" binding:"required,email"`
	Password   string    `gorm:"not null" json:"password" binding:"required"` // "-" hides password in JSON
	Bookings   []Booking `json:"bookings,omitempty"`
}
