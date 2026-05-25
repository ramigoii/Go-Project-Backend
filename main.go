package main

import (
	"MovieDatabase/config"
	"MovieDatabase/handlers"
	"MovieDatabase/middleware"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDatabase()

	middleware.InitAuthMiddleware()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))
	// ─── PUBLIC ROUTES ───────────────────────────────────────────
	// Movies — read-only public
	r.GET("/movies", handlers.GetAllMovies)
	r.GET("/movies/search", handlers.SearchMovies) // MUST be before /movies/:id
	r.GET("/movies/:id", handlers.GetMovieByID)

	// Reviews — read-only public
	r.GET("/movies/:id/reviews", handlers.GetMovieReviews)
	r.GET("/reviews/:id", handlers.GetReviewByID)

	// Seats — public info
	r.GET("/movies/:id/seats", handlers.GetAvailableSeats)

	// ─── PROTECTED ROUTES ────────────────────────────────────────
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		// Movies — write operations
		auth.POST("/movies", handlers.CreateMovie)
		auth.PUT("/movies/:id", handlers.UpdateMovie)
		auth.DELETE("/movies/:id", handlers.DeleteMovie)

		// Movie bookings (nested, MUST be registered before /movies/:id above — already handled since public group is separate)
		auth.GET("/movies/:id/bookings", handlers.GetMovieBookings)

		// Reviews — write operations
		auth.POST("/movies/:id/reviews", handlers.CreateReview)
		auth.PUT("/reviews/:id", handlers.UpdateReview)
		auth.DELETE("/reviews/:id", handlers.DeleteReview)

		// Bookings — fully private
		auth.POST("/bookings", handlers.CreateBooking)
		auth.GET("/bookings", handlers.GetUserBookings)
		auth.GET("/bookings/:id", handlers.GetBookingByID)
		auth.PUT("/bookings/:id", handlers.UpdateBooking)
		auth.DELETE("/bookings/:id", handlers.CancelBooking)
	}

	log.Println("Movie API running on :8080")
	r.Run(":8080")
}
