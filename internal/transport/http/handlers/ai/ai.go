package ai

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

func (h *Handler) ChatCompletions(c *gin.Context) { response.Success(c, gin.H{"message": "Chat completions - TODO"}) }
func (h *Handler) Completions(c *gin.Context)     { response.Success(c, gin.H{"message": "Completions - TODO"}) }
func (h *Handler) Embeddings(c *gin.Context)      { response.Success(c, gin.H{"message": "Embeddings - TODO"}) }
func (h *Handler) ListModels(c *gin.Context)      { response.Success(c, gin.H{"message": "List models - TODO"}) }
func (h *Handler) GetModel(c *gin.Context)        { response.Success(c, gin.H{"message": "Get model - TODO"}) }