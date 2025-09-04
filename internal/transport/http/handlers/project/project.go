package project

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

func (h *Handler) List(c *gin.Context)   { response.Success(c, gin.H{"message": "List projects - TODO"}) }
func (h *Handler) Create(c *gin.Context) { response.Success(c, gin.H{"message": "Create project - TODO"}) }
func (h *Handler) Get(c *gin.Context)    { response.Success(c, gin.H{"message": "Get project - TODO"}) }
func (h *Handler) Update(c *gin.Context) { response.Success(c, gin.H{"message": "Update project - TODO"}) }
func (h *Handler) Delete(c *gin.Context) { response.Success(c, gin.H{"message": "Delete project - TODO"}) }