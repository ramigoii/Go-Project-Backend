package handlers

import (
	"MovieDatabase/config"
	"MovieDatabase/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateReview(c *gin.Context) {
	movieID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	// Check movie exists
	var movie models.Movie
	if err := config.DB.First(&movie, uint(movieID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	// Check user hasn't reviewed this movie already
	var existing models.Review
	if err := config.DB.Where("user_id = ? AND movie_id = ?", userID, movieID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already reviewed this movie"})
		return
	}

	var input struct {
		Rating  int    `json:"rating" binding:"required,min=1,max=10"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review := models.Review{
		UserID:  userID.(uint),
		MovieID: uint(movieID),
		Rating:  input.Rating,
		Comment: input.Comment,
	}

	if err := config.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	review.Username = username.(string)

	c.JSON(http.StatusCreated, gin.H{"data": review})
}

func GetMovieReviews(c *gin.Context) {
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

	var reviews []models.Review
	if err := config.DB.Where("movie_id = ?", movieID).
		Order("created_at desc").
		Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate average rating
	var avgRating float64
	if len(reviews) > 0 {
		total := 0
		for _, r := range reviews {
			total += r.Rating
		}
		avgRating = float64(total) / float64(len(reviews))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       reviews,
		"count":      len(reviews),
		"avg_rating": avgRating,
	})
}

func GetReviewByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var review models.Review
	if err := config.DB.Preload("Movie").First(&review, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": review})
}

func UpdateReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var review models.Review
	if err := config.DB.First(&review, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	if review.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this review"})
		return
	}

	var input struct {
		Rating  int    `json:"rating" binding:"min=1,max=10"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Rating != 0 {
		review.Rating = input.Rating
	}
	if input.Comment != "" {
		review.Comment = input.Comment
	}

	if err := config.DB.Save(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": review})
}

func DeleteReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var review models.Review
	if err := config.DB.First(&review, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	if review.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this review"})
		return
	}

	if err := config.DB.Delete(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}
