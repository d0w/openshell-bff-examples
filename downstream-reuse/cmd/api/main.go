// Package main is a downstream BFF that reuses 100% of upstream's code.
//
// It adds nothing, overrides nothing, and decorates nothing. Every symbol
// below — server.ServerConfig, server.Services, server.NewServer,
// services.NewDefaultGatewayService, services.NewDefaultSandboxService —
// is imported directly from the upstream module. This file is the "does
// upstream's boilerplate work as-is for a consumer that wants zero
// changes?" baseline. See downstream-partialreuse for what interface
// decoration looks like once downstream needs to diverge.
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
	"github.com/d0w/openshell-bff-examples/upstream/pkg/services"
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
		Gateway: services.NewDefaultGatewayService(),
		Sandbox: services.NewDefaultSandboxService(),
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
