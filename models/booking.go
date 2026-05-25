package models

import (
	"time"

	"gorm.io/gorm"
)

type Booking struct {
	gorm.Model `json:"-"`
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	MovieID    uint      `gorm:"not null" json:"movie_id"`
	SeatNumber string    `gorm:"not null;size:10" json:"seat_number" binding:"required"`
	ShowTime   time.Time `gorm:"not null" json:"show_time" binding:"required"`
	TotalPrice float64   `gorm:"default:0" json:"total_price"`
	Movie      Movie     `gorm:"foreignKey:MovieID" json:"movie,omitempty"`
}
