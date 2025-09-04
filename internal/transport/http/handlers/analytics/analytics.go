package analytics

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

func (h *Handler) Overview(c *gin.Context)  { response.Success(c, gin.H{"message": "Analytics overview - TODO"}) }
func (h *Handler) Requests(c *gin.Context)  { response.Success(c, gin.H{"message": "Analytics requests - TODO"}) }
func (h *Handler) Costs(c *gin.Context)     { response.Success(c, gin.H{"message": "Analytics costs - TODO"}) }
func (h *Handler) Providers(c *gin.Context) { response.Success(c, gin.H{"message": "Analytics providers - TODO"}) }
func (h *Handler) Models(c *gin.Context)    { response.Success(c, gin.H{"message": "Analytics models - TODO"}) }