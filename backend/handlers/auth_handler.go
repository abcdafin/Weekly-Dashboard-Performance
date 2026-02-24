package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"weekly-dashboard/config"
	"weekly-dashboard/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// generateState generates a random state string for OAuth
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GoogleLogin initiates Google OAuth flow
// @Summary Initiate Google OAuth login
// @Description Redirects user to Google OAuth consent screen
// @Tags auth
// @Produce json
// @Success 302 {string} string "Redirect to Google OAuth"
// @Router /api/v1/auth/google [get]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := generateState()

	// Store state in cookie for validation
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	url := h.authService.GetAuthURL(state)
	log.Printf("Redirecting to Google OAuth: %s", url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles Google OAuth callback
// @Summary Handle Google OAuth callback
// @Description Exchanges authorization code for tokens and creates user session
// @Tags auth
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Param state query string true "OAuth state parameter"
// @Success 302 {string} string "Redirect to frontend with token"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Validate state
	state := c.Query("state")
	savedState, err := c.Cookie("oauth_state")
	if err != nil || state != savedState {
		log.Printf("State mismatch: received=%s, saved=%s", state, savedState)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid OAuth state",
		})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Authorization code not provided",
		})
		return
	}

	// Exchange code for tokens
	token, err := h.authService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to exchange authorization code",
		})
		return
	}

	// Get user info from Google
	userInfo, err := h.authService.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get user information",
		})
		return
	}

	// Create or update user in database
	user, err := h.authService.CreateOrUpdateUser(userInfo, token)
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create user account",
		})
		return
	}

	// Generate JWT token
	jwtToken, err := h.authService.GenerateJWT(user)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate authentication token",
		})
		return
	}

	// Redirect to frontend with token
	frontendURL := config.AppConfig.FrontendURL
	redirectURL := frontendURL + "/auth/callback?token=" + jwtToken
	log.Printf("Authentication successful for user: %s, redirecting to: %s", user.Email, redirectURL)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// Logout handles user logout
// @Summary Logout user
// @Description Clears user session
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear cookies if any
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetCurrentUser returns the current authenticated user
// @Summary Get current user
// @Description Returns current authenticated user information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User information"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}
