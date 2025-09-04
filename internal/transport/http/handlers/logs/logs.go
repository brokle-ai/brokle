package logs

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

func (h *Handler) ListRequests(c *gin.Context) { response.Success(c, gin.H{"message": "List requests - TODO"}) }
func (h *Handler) GetRequest(c *gin.Context)   { response.Success(c, gin.H{"message": "Get request - TODO"}) }
func (h *Handler) Export(c *gin.Context)       { response.Success(c, gin.H{"message": "Export logs - TODO"}) }