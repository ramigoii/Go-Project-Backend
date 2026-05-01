package main

import (
	"MovieDatabase/config"
	"MovieDatabase/handlers"
	"MovieDatabase/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDatabase()

	handlers.InitAuthHandler()
	middleware.InitAuthMiddleware()

	r := gin.Default()

	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Защищенные маршруты
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/bookings", handlers.CreateBooking)
		protected.GET("/bookings", handlers.GetUserBookings)
		protected.GET("/bookings/:id", handlers.GetBookingByID)
		protected.DELETE("/bookings/:id", handlers.CancelBooking)
		protected.PUT("/bookings/:id", handlers.UpdateBooking)

		protected.GET("/movies", handlers.GetAllMovies)
		protected.GET("/movies/:id", handlers.GetMovieByID)
		protected.POST("/movies", handlers.CreateMovie)
		protected.PUT("/movies/:id", handlers.UpdateMovie)
		protected.DELETE("/movies/:id", handlers.DeleteMovie)
		protected.GET("/movies/search", handlers.SearchMovies)
	}

	log.Println("Movie API running on :8080")
	r.Run(":8080")
}
