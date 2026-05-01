package services

import (
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

type AuthClient struct {
	client      *resty.Client
	baseURL     string
	internalKey string
}

type User struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenValidationResponse struct {
	Valid    bool   `json:"valid"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

func NewAuthClient() *AuthClient {
	baseURL := os.Getenv("AUTH_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	internalKey := os.Getenv("INTERNAL_API_KEY")
	if internalKey == "" {
		internalKey = "your-internal-secret-key"
	}

	client := resty.New()
	client.SetTimeout(5 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)

	return &AuthClient{
		client:      client,
		baseURL:     baseURL,
		internalKey: internalKey,
	}
}

func (c *AuthClient) RegisterUser(username, email, password string) (uint, error) {
	var result struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"username": username,
			"email":    email,
			"password": password,
		}).
		SetResult(&result).
		Post(c.baseURL + "/auth/register")

	if err != nil {
		return 0, fmt.Errorf("failed to register user: %w", err)
	}

	if resp.StatusCode() != 201 {
		return 0, fmt.Errorf("registration failed: %s", resp.String())
	}

	return result.ID, nil
}

func (c *AuthClient) LoginUser(username, password string) (string, uint, error) {
	var result LoginResponse

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"username": username,
			"password": password,
		}).
		SetResult(&result).
		Post(c.baseURL + "/auth/login")

	if err != nil {
		return "", 0, fmt.Errorf("failed to login: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", 0, fmt.Errorf("login failed: %s", resp.String())
	}

	return result.Token, result.UserID, nil
}

func (c *AuthClient) ValidateToken(token string) (*TokenValidationResponse, error) {
	var result TokenValidationResponse

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"token": token,
		}).
		SetResult(&result).
		Post(c.baseURL + "/auth/validate")

	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("validation failed: %s", resp.String())
	}

	return &result, nil
}

func (c *AuthClient) GetUserByID(userID uint) (*User, error) {
	var result struct {
		User User `json:"user"`
	}

	resp, err := c.client.R().
		SetHeader("X-Internal-API-Key", c.internalKey).
		SetResult(&result).
		Get(fmt.Sprintf("%s/internal/users/%d", c.baseURL, userID))

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("get user failed: %s", resp.String())
	}

	return &result.User, nil
}

func (c *AuthClient) GetUserByUsername(username string) (*User, error) {
	var result struct {
		User User `json:"user"`
	}

	resp, err := c.client.R().
		SetHeader("X-Internal-API-Key", c.internalKey).
		SetResult(&result).
		Get(fmt.Sprintf("%s/internal/users/by-username/%s", c.baseURL, username))

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("get user failed: %s", resp.String())
	}

	return &result.User, nil
}
