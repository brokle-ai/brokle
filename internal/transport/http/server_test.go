package http

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	serveStartupWait = 100 * time.Millisecond
	errorPropagation = 1 * time.Second
	shutdownTimeout  = 5 * time.Second
)

func TestServerShutdown_FiltersExpectedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	httpServer := &http.Server{
		Addr:    ":0",
		Handler: engine,
	}

	serveErr := make(chan error, 1)

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to bind: %v", err)
	}

	go func() {
		if err := httpServer.Serve(lis); err != nil && err != http.ErrServerClosed {
			serveErr <- err
		}
	}()

	time.Sleep(serveStartupWait)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	shutdownErr := httpServer.Shutdown(ctx)
	if shutdownErr != nil {
		t.Errorf("Shutdown returned error: %v", shutdownErr)
	}

	select {
	case err := <-serveErr:
		t.Errorf("Expected no error in serveErr channel during graceful shutdown, got: %v", err)
	case <-time.After(errorPropagation):
	}
}
