package grpc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
)

const (
	serveStartupWait = 100 * time.Millisecond
	errorPropagation = 1 * time.Second
	shutdownTimeout  = 5 * time.Second
)

func TestServerShutdown_FiltersExpectedError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	grpcServer := grpc.NewServer()
	server := &Server{
		grpcServer: grpcServer,
		logger:     logger,
		port:       0,
	}

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(serveStartupWait)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	shutdownErr := server.Shutdown(ctx)
	if shutdownErr != nil {
		t.Errorf("Shutdown returned error: %v", shutdownErr)
	}

	select {
	case err := <-server.ServeErr():
		t.Errorf("Expected no error in serveErr channel during graceful shutdown, got: %v", err)
	case <-time.After(errorPropagation):
	}
}

func TestServerShutdown_SendsUnexpectedErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	grpcServer := grpc.NewServer()
	server := &Server{
		grpcServer: grpcServer,
		logger:     logger,
		port:       0,
	}

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if server.listener != nil {
		_ = server.listener.Close()
	}

	select {
	case err := <-server.ServeErr():
		if err == nil {
			t.Error("Expected non-nil error in serveErr channel")
		}
		if errors.Is(err, grpc.ErrServerStopped) {
			t.Error("Expected non-shutdown error, got ErrServerStopped")
		}
	case <-time.After(errorPropagation):
		t.Error("Expected error in serveErr channel for unexpected failure")
	}

	server.grpcServer.Stop()
}

func TestServerStart_BindError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	grpcServer1 := grpc.NewServer()
	server1 := &Server{
		grpcServer: grpcServer1,
		logger:     logger,
		port:       0,
	}

	err := server1.Start()
	if err != nil {
		t.Fatalf("First server failed to start: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = server1.Shutdown(ctx)
	}()

	actualPort := server1.listener.Addr().(*net.TCPAddr).Port

	grpcServer2 := grpc.NewServer()
	server2 := &Server{
		grpcServer: grpcServer2,
		logger:     logger,
		port:       actualPort,
	}

	err = server2.Start()
	if err == nil {
		t.Fatal("Expected second server to fail binding to same port")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "bind") && !strings.Contains(errMsg, "address already in use") {
		t.Errorf("Expected bind-related error, got: %v", err)
	}
}
