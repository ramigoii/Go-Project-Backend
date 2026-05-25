package models

import "gorm.io/gorm"

type Review struct {
	gorm.Model `json:"-"`
	ID         uint   `gorm:"primaryKey" json:"id"`
	UserID     uint   `gorm:"not null" json:"user_id"`
	MovieID    uint   `gorm:"not null" json:"movie_id"`
	Rating     int    `gorm:"not null;check:rating >= 1 AND rating <= 10" json:"rating"`
	Comment    string `gorm:"type:text" json:"comment"`
	Username   string `gorm:"-" json:"username,omitempty"` // filled from auth-service
	Movie      Movie  `gorm:"foreignKey:MovieID" json:"movie,omitempty"`
}
