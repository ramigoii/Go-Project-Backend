package handlers_test

import (
	"MovieDatabase/config"
	"MovieDatabase/handlers"
	"MovieDatabase/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory SQLite DB for tests
func setupTestDB(t *testing.T) {
	dsn := "host=localhost user=postgres password=postgres dbname=movie_db_test port=5434 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.Migrator().DropTable(
		&models.Review{},
		&models.Booking{},
		&models.Movie{},
	)
	if err != nil {
		t.Log(err)
	}

	err = db.AutoMigrate(
		&models.Movie{},
		&models.Booking{},
		&models.Review{},
	)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	config.DB = db
}

// authMiddleware injects a fake user_id and username into gin context
func fakeAuthMiddleware(userID uint, username string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("username", username)
		c.Next()
	}
}

func setupRouter(userID uint, username string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(fakeAuthMiddleware(userID, username))
	return r
}

// ─── MOVIE TESTS ────────────────────────────────────────────────────────────

// Test 1: Get all movies returns 200 and empty list initially
func TestGetAllMovies_Empty(t *testing.T) {
	setupTestDB(t)
	r := setupRouter(1, "testuser")
	r.GET("/movies", handlers.GetAllMovies)

	req, _ := http.NewRequest(http.MethodGet, "/movies", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Equal(t, 0, len(data))
}

// Test 2: Create a movie returns 201 with correct data
func TestCreateMovie_Success(t *testing.T) {
	setupTestDB(t)
	r := setupRouter(1, "testuser")
	r.POST("/movies", handlers.CreateMovie)

	body := `{"title":"Inception","description":"A mind-bending thriller"}`
	req, _ := http.NewRequest(http.MethodPost, "/movies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Inception", data["title"])
}

// Test 3: Create movie without title returns 400
func TestCreateMovie_MissingTitle(t *testing.T) {
	setupTestDB(t)
	r := setupRouter(1, "testuser")
	r.POST("/movies", handlers.CreateMovie)

	body := `{"description":"No title here"}`
	req, _ := http.NewRequest(http.MethodPost, "/movies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test 4: Get movie by ID returns 200 with correct movie
func TestGetMovieByID_Success(t *testing.T) {
	setupTestDB(t)
	movie := models.Movie{Title: "Interstellar", Description: "Space odyssey"}
	config.DB.Create(&movie)

	r := setupRouter(1, "testuser")
	r.GET("/movies/:id", handlers.GetMovieByID)

	req, _ := http.NewRequest(http.MethodGet, "/movies/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Interstellar", data["title"])
}

// Test 5: Get movie by non-existent ID returns 404
func TestGetMovieByID_NotFound(t *testing.T) {
	setupTestDB(t)
	r := setupRouter(1, "testuser")
	r.GET("/movies/:id", handlers.GetMovieByID)

	req, _ := http.NewRequest(http.MethodGet, "/movies/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Test 6: Delete movie without bookings returns 200
func TestDeleteMovie_Success(t *testing.T) {
	setupTestDB(t)
	movie := models.Movie{Title: "To Delete"}
	config.DB.Create(&movie)

	r := setupRouter(1, "testuser")
	r.DELETE("/movies/:id", handlers.DeleteMovie)

	req, _ := http.NewRequest(http.MethodDelete, "/movies/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test 7: Search movies by title returns matching results
func TestSearchMovies_Found(t *testing.T) {
	setupTestDB(t)
	config.DB.Create(&models.Movie{Title: "The Matrix"})
	config.DB.Create(&models.Movie{Title: "John Wick"})

	r := setupRouter(1, "testuser")
	r.GET("/movies/search", handlers.SearchMovies)

	req, _ := http.NewRequest(http.MethodGet, "/movies/search?title=matrix", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Equal(t, 1, len(data))
}

// ─── REVIEW TESTS ────────────────────────────────────────────────────────────

// Test 8: Create review returns 201
func TestCreateReview_Success(t *testing.T) {
	setupTestDB(t)
	config.DB.Create(&models.Movie{Title: "Dune"})

	r := setupRouter(1, "testuser")
	r.POST("/movies/:id/reviews", handlers.CreateReview)

	body := `{"rating":9,"comment":"Epic sci-fi"}`
	req, _ := http.NewRequest(http.MethodPost, "/movies/1/reviews", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(9), data["rating"])
}

// Test 9: Duplicate review returns 409
func TestCreateReview_Duplicate(t *testing.T) {
	setupTestDB(t)
	config.DB.Create(&models.Movie{Title: "Dune"})
	config.DB.Create(&models.Review{UserID: 1, MovieID: 1, Rating: 8, Comment: "Great"})

	r := setupRouter(1, "testuser")
	r.POST("/movies/:id/reviews", handlers.CreateReview)

	body := `{"rating":7,"comment":"Second review attempt"}`
	req, _ := http.NewRequest(http.MethodPost, "/movies/1/reviews", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// Test 10: Delete review by another user returns 403
func TestDeleteReview_Forbidden(t *testing.T) {
	setupTestDB(t)
	config.DB.Create(&models.Movie{Title: "Dune"})
	// Review belongs to user 2
	config.DB.Create(&models.Review{UserID: 2, MovieID: 1, Rating: 8, Comment: "Good"})

	// But request comes from user 1
	r := setupRouter(1, "testuser")
	r.DELETE("/reviews/:id", handlers.DeleteReview)

	req, _ := http.NewRequest(http.MethodDelete, "/reviews/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
