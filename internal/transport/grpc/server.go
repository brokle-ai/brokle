package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// Server wraps gRPC server with lifecycle management
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *slog.Logger
	port       int
}

// NewServer creates a new gRPC server for OTLP ingestion
func NewServer(
	port int,
	otlpHandler *OTLPHandler,
	otlpMetricsHandler *OTLPMetricsHandler,
	otlpLogsHandler *OTLPLogsHandler,
	authInterceptor *AuthInterceptor,
	logger *slog.Logger,
) (*Server, error) {
	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		// Chain interceptors (memory limiter → auth → logging)
		// Memory limiter FIRST to reject requests before auth/processing
		grpc.ChainUnaryInterceptor(
			MemoryLimiterInterceptor(DefaultMemoryLimiterConfig(), logger),
			authInterceptor.Unary(),
			LoggingInterceptor(logger),
		),

		// Size limits (match HTTP 10MB request limit)
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB max request
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB max response

		// Keepalive settings for long-lived connections
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    1 * time.Minute,  // Send keepalive ping every 1 minute
			Timeout: 20 * time.Second, // Wait 20s for ping ack before closing
		}),

		// Enforce keepalive from clients
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second, // Min time between pings
			PermitWithoutStream: true,             // Allow pings without active streams
		}),
	)

	// Register OTLP service handlers (all three signal types)
	RegisterOTLPTraceService(grpcServer, otlpHandler)
	RegisterOTLPMetricsService(grpcServer, otlpMetricsHandler)
	RegisterOTLPLogsService(grpcServer, otlpLogsHandler)

	return &Server{
		grpcServer: grpcServer,
		logger:     logger,
		port:       port,
	}, nil
}

// Start begins listening and serving gRPC requests (blocking)
func (s *Server) Start() error {
	// Create TCP listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}
	s.listener = lis

	s.logger.Info("Starting gRPC OTLP server", "port", s.port)

	// Start serving (blocks until server stops)
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("gRPC server failed: %w", err)
	}

	return nil
}

// Shutdown gracefully stops the gRPC server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Gracefully stopping gRPC server")

	// Graceful stop with timeout
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		// Timeout exceeded - force stop
		s.logger.Warn("Graceful shutdown timeout, forcing stop")
		s.grpcServer.Stop()
		return ctx.Err()
	case <-stopped:
		s.logger.Info("gRPC server stopped gracefully")
		return nil
	}
}
