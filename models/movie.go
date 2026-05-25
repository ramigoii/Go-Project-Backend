package models

import "gorm.io/gorm"

type Movie struct {
	gorm.Model  `json:"-"`
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null;size:255" json:"title" binding:"required"`
	Description string    `gorm:"type:text" json:"description"`
	Bookings    []Booking `json:"bookings,omitempty"`
	Reviews     []Review  `json:"reviews,omitempty"`
}
