package user

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/user"
	"brokle/pkg/response"
)

// Handler handles user endpoints
type Handler struct {
	config      *config.Config
	logger      *logrus.Logger
	userService user.Service
}

// NewHandler creates a new user handler
func NewHandler(config *config.Config, logger *logrus.Logger, userService user.Service) *Handler {
	return &Handler{
		config:      config,
		logger:      logger,
		userService: userService,
	}
}

// GetProfile handles GET /users/me
func (h *Handler) GetProfile(c *gin.Context) {
	response.Success(c, gin.H{"message": "Get user profile - TODO"})
}

// UpdateProfile handles PUT /users/me
func (h *Handler) UpdateProfile(c *gin.Context) {
	response.Success(c, gin.H{"message": "Update user profile - TODO"})
}

// GetPreferences handles GET /users/me/preferences
func (h *Handler) GetPreferences(c *gin.Context) {
	response.Success(c, gin.H{"message": "Get user preferences - TODO"})
}

// UpdatePreferences handles PUT /users/me/preferences
func (h *Handler) UpdatePreferences(c *gin.Context) {
	response.Success(c, gin.H{"message": "Update user preferences - TODO"})
}