package apikey

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"brokle/internal/config"
	"brokle/pkg/response"
)

type Handler struct {
	config *config.Config
	logger *logrus.Logger
}

func NewHandler(config *config.Config, logger *logrus.Logger) *Handler {
	return &Handler{config: config, logger: logger}
}

func (h *Handler) List(c *gin.Context)   { response.Success(c, gin.H{"message": "List API keys - TODO"}) }
func (h *Handler) Create(c *gin.Context) { response.Success(c, gin.H{"message": "Create API key - TODO"}) }
func (h *Handler) Delete(c *gin.Context) { response.Success(c, gin.H{"message": "Delete API key - TODO"}) }