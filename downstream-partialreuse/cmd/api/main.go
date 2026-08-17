// Package main is a downstream BFF that reuses most of upstream's code
// unmodified (server.NewServer, Services, SandboxHandler,
// NewDefaultGatewayService) and demonstrates two extension mechanisms on
// top of it:
//
//   - Interface decoration: pkg/sandbox.Service embeds upstream's
//     SandboxService and overrides CreateSandbox; Get/List/Delete are
//     inherited untouched. Upstream's SandboxHandler is reused as-is since
//     the interface's shape didn't change.
//   - A brand new handler with its own middleware stack: pkg/audit.Handler
//     exposes a capability (sandbox stats) that has no equivalent in
//     upstream at all. It's attached via server.WithRoutes, upstream's
//     extensibility hook -- Option gets the raw chi.Router, so this file
//     decides the middleware, not upstream. Below, /api/audit reuses
//     upstream's exported middleware.RequireAuth; /debug/ping uses a
//     completely separate, downstream-only auth scheme, to show both are
//     equally possible with no special-casing on upstream's side.
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

	downstreammiddleware "github.com/d0w/openshell-bff-examples/downstream-partialreuse/pkg/middleware"
	upstreammiddleware "github.com/d0w/openshell-bff-examples/upstream/pkg/middleware"
	"github.com/d0w/openshell-bff-examples/upstream/pkg/server"
	upstreamservices "github.com/d0w/openshell-bff-examples/upstream/pkg/services"

	"github.com/d0w/openshell-bff-examples/downstream-partialreuse/pkg/audit"
	downstreamservices "github.com/d0w/openshell-bff-examples/downstream-partialreuse/pkg/services"

	"github.com/go-chi/chi/v5"
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

	// Base sandbox service is upstream's own default implementation --
	// downstream doesn't reimplement storage, it just wraps it.
	baseSandboxSvc := upstreamservices.NewDefaultSandboxService()
	decoratedSandboxSvc := downstreamservices.NewService(baseSandboxSvc)

	svcs := server.Services{
		Gateway: upstreamservices.NewDefaultGatewayService(),
		Sandbox: decoratedSandboxSvc,
	}

	// auditHandler is a downstream-only route that doesn't exist in
	// upstream at all. It reads through the same decorated sandbox service
	// everything else uses, so its stats reflect downstream's
	// CreateSandbox behavior too.
	auditHandler := audit.NewHandler(decoratedSandboxSvc)

	srv := server.NewServer(
		cfg, svcs,
		server.WithRoutes("/api/audit", auditHandler, upstreammiddleware.RequireAuth),

		// Mount at /debug/ping with a completely separate, downstream-only
		// auth scheme. server.Option is handed the raw router, so this
		// isn't a special case upstream had to account for -- it's just a
		// route group with its own middleware, like any chi router.
		func(r chi.Router) {
			r.Route("/debug", func(r chi.Router) {
				r.Use(downstreammiddleware.RequireDownstreamKey)
				r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("pong"))
				})
			})
		},
	)

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
