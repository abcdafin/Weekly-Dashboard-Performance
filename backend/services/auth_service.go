package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"weekly-dashboard/config"
	"weekly-dashboard/database"
	"weekly-dashboard/middleware"
	"weekly-dashboard/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUserInfo represents user info from Google API
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// AuthService handles authentication operations
type AuthService struct {
	oauthConfig *oauth2.Config
}

// NewAuthService creates a new AuthService instance
func NewAuthService() *AuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     config.AppConfig.GoogleClientID,
		ClientSecret: config.AppConfig.GoogleClientSecret,
		RedirectURL:  config.AppConfig.GoogleRedirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/spreadsheets.readonly",
		},
		Endpoint: google.Endpoint,
	}

	return &AuthService{
		oauthConfig: oauthConfig,
	}
}

// GetAuthURL generates the Google OAuth authorization URL
func (s *AuthService) GetAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCode exchanges authorization code for tokens
func (s *AuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.oauthConfig.Exchange(ctx, code)
}

// GetUserInfo fetches user info from Google API
func (s *AuthService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := s.oauthConfig.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: status %d, body: %s", resp.StatusCode, string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// CreateOrUpdateUser creates or updates a user in the database
func (s *AuthService) CreateOrUpdateUser(userInfo *GoogleUserInfo, token *oauth2.Token) (*models.User, error) {
	var user models.User

	result := database.DB.Where("email = ?", userInfo.Email).First(&user)

	if result.RowsAffected == 0 {
		// Create new user
		user = models.User{
			Email:        userInfo.Email,
			Name:         userInfo.Name,
			Picture:      userInfo.Picture,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			TokenExpiry:  token.Expiry,
			LastLogin:    time.Now(),
		}
		if err := database.DB.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		log.Printf("Created new user: %s", user.Email)
	} else {
		// Update existing user
		user.Name = userInfo.Name
		user.Picture = userInfo.Picture
		user.AccessToken = token.AccessToken
		if token.RefreshToken != "" {
			user.RefreshToken = token.RefreshToken
		}
		user.TokenExpiry = token.Expiry
		user.LastLogin = time.Now()

		if err := database.DB.Save(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		log.Printf("Updated user: %s", user.Email)
	}

	return &user, nil
}

// GenerateJWT generates a JWT token for the user
func (s *AuthService) GenerateJWT(user *models.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(config.AppConfig.JWTExpiration) * time.Hour)

	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "weekly-dashboard",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GetOAuthConfig returns the OAuth config for creating clients
func (s *AuthService) GetOAuthConfig() *oauth2.Config {
	return s.oauthConfig
}

// RefreshToken refreshes the user's OAuth token if expired
func (s *AuthService) RefreshToken(ctx context.Context, user *models.User) (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Expiry:       user.TokenExpiry,
	}

	// Check if token is expired
	if token.Expiry.After(time.Now()) {
		return token, nil
	}

	// Refresh the token
	tokenSource := s.oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update user with new token
	user.AccessToken = newToken.AccessToken
	if newToken.RefreshToken != "" {
		user.RefreshToken = newToken.RefreshToken
	}
	user.TokenExpiry = newToken.Expiry

	if err := database.DB.Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return newToken, nil
}
