package handlers

import (
	"MovieDatabase/config"
	"MovieDatabase/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllMovies(c *gin.Context) {
	var movies []models.Movie

	title := c.Query("title")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	offset := (page - 1) * limit

	query := config.DB

	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}

	if err := query.
		Limit(limit).
		Offset(offset).
		Find(&movies).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":  page,
		"limit": limit,
		"data":  movies,
	})
}

func GetMovieByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	var movie models.Movie
	if err := config.DB.
		Preload("Bookings").
		First(&movie, uint(id)).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movie})
}

func CreateMovie(c *gin.Context) {
	var input models.Movie

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": input})
}

func UpdateMovie(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	var movie models.Movie
	if err := config.DB.First(&movie, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	var input models.Movie
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Model(&movie).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movie})
}

func DeleteMovie(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	var bookingCount int64
	config.DB.Model(&models.Booking{}).
		Where("movie_id = ?", id).
		Count(&bookingCount)

	if bookingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Cannot delete movie with existing bookings",
		})
		return
	}

	if err := config.DB.Delete(&models.Movie{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie deleted successfully"})
}

func SearchMovies(c *gin.Context) {
	title := c.Query("title")

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title query is required"})
		return
	}

	var movies []models.Movie
	if err := config.DB.
		Where("title ILIKE ?", "%"+title+"%").
		Find(&movies).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movies})
}
