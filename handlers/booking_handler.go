package handlers

import (
	"MovieDatabase/config"
	"MovieDatabase/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateBooking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var input struct {
		MovieID    uint      `json:"movie_id" binding:"required"`
		SeatNumber string    `json:"seat_number" binding:"required"`
		ShowTime   time.Time `json:"show_time" binding:"required"`
		TotalPrice float64   `json:"total_price"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var movie models.Movie
	if err := config.DB.First(&movie, input.MovieID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	// Check seat not already taken for this showtime
	var existingBooking models.Booking
	if err := config.DB.Where(
		"movie_id = ? AND seat_number = ? AND show_time = ? AND deleted_at IS NULL",
		input.MovieID, input.SeatNumber, input.ShowTime,
	).First(&existingBooking).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Seat already booked for this showtime"})
		return
	}

	booking := models.Booking{
		UserID:     userID.(uint),
		MovieID:    input.MovieID,
		SeatNumber: input.SeatNumber,
		ShowTime:   input.ShowTime,
		TotalPrice: input.TotalPrice,
	}

	if err := config.DB.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config.DB.Preload("Movie").First(&booking, booking.ID)
	c.JSON(http.StatusCreated, gin.H{"data": booking})
}

func GetUserBookings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var bookings []models.Booking
	if err := config.DB.
		Where("user_id = ?", userID).
		Preload("Movie").
		Order("show_time desc").
		Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": bookings})
}

func GetBookingByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var booking models.Booking
	if err := config.DB.Preload("Movie").First(&booking, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	if booking.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view this booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": booking})
}

func CancelBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var booking models.Booking
	if err := config.DB.First(&booking, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	if booking.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to cancel this booking"})
		return
	}

	if booking.ShowTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel past bookings"})
		return
	}

	if err := config.DB.Delete(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
}

func UpdateBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var booking models.Booking
	if err := config.DB.First(&booking, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	if booking.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this booking"})
		return
	}

	var input struct {
		SeatNumber string    `json:"seat_number"`
		ShowTime   time.Time `json:"show_time"`
		TotalPrice float64   `json:"total_price"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.SeatNumber != "" {
		// Check new seat not taken
		var existing models.Booking
		if err := config.DB.Where(
			"movie_id = ? AND seat_number = ? AND show_time = ? AND id != ? AND deleted_at IS NULL",
			booking.MovieID, input.SeatNumber, booking.ShowTime, booking.ID,
		).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Seat already booked for this showtime"})
			return
		}
		booking.SeatNumber = input.SeatNumber
	}

	if !input.ShowTime.IsZero() {
		booking.ShowTime = input.ShowTime
	}

	if input.TotalPrice > 0 {
		booking.TotalPrice = input.TotalPrice
	}

	if err := config.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config.DB.Preload("Movie").First(&booking, booking.ID)
	c.JSON(http.StatusOK, gin.H{"data": booking})
}

func GetMovieBookings(c *gin.Context) {
	movieID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	var movie models.Movie
	if err := config.DB.First(&movie, uint(movieID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	var bookings []models.Booking
	if err := config.DB.
		Where("movie_id = ?", movieID).
		Order("show_time asc").
		Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": bookings})
}

func GetAvailableSeats(c *gin.Context) {
	movieID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	showTimeStr := c.Query("show_time")
	if showTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "show_time query parameter is required"})
		return
	}

	showTime, err := time.Parse(time.RFC3339, showTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid show_time format, use RFC3339 e.g. 2025-12-01T18:00:00Z"})
		return
	}

	var movie models.Movie
	if err := config.DB.First(&movie, uint(movieID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	var bookedSeats []string
	config.DB.Model(&models.Booking{}).
		Where("movie_id = ? AND show_time = ? AND deleted_at IS NULL", movieID, showTime).
		Pluck("seat_number", &bookedSeats)

	// Generate all seats A1-E10
	rows := []string{"A", "B", "C", "D", "E"}
	allSeats := make([]string, 0, 50)
	for _, row := range rows {
		for i := 1; i <= 10; i++ {
			allSeats = append(allSeats, row+strconv.Itoa(i))
		}
	}

	bookedMap := make(map[string]bool)
	for _, seat := range bookedSeats {
		bookedMap[seat] = true
	}

	availableSeats := make([]string, 0)
	for _, seat := range allSeats {
		if !bookedMap[seat] {
			availableSeats = append(availableSeats, seat)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"movie_id":        movieID,
		"movie_title":     movie.Title,
		"show_time":       showTime,
		"booked_seats":    bookedSeats,
		"available_seats": availableSeats,
		"total_seats":     len(allSeats),
		"available_count": len(availableSeats),
	})
}
