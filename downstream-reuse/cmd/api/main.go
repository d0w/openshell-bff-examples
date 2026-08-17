package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/d0w/openshell-bff-examples/upstream/pkg/server"
	gateway "github.com/d0w/openshell-bff-examples/upstream/pkg/services/gateway"
	sandbox "github.com/d0w/openshell-bff-examples/upstream/pkg/services/sandbox"
)

func main() {
	logger := slog.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := server.ServerConfig{
		Port:         ":" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	svcs := server.Services{
		Gateway: gateway.NewDefaultGatewayService(),
		Sandbox: sandbox.NewDefaultSandboxService(),
	}

	srv := server.NewServer(cfg, svcs)

	go func() {
		logger.Info("starting server", "addr", cfg.Port)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
