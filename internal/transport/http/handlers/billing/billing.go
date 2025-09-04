package billing

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

func (h *Handler) GetUsage(c *gin.Context)           { response.Success(c, gin.H{"message": "Get usage - TODO"}) }
func (h *Handler) ListInvoices(c *gin.Context)       { response.Success(c, gin.H{"message": "List invoices - TODO"}) }
func (h *Handler) GetSubscription(c *gin.Context)    { response.Success(c, gin.H{"message": "Get subscription - TODO"}) }
func (h *Handler) UpdateSubscription(c *gin.Context) { response.Success(c, gin.H{"message": "Update subscription - TODO"}) }