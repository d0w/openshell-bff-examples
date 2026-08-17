package server

import (
	"context"
	"net/http"
	"time"

	"github.com/d0w/openshell-bff-examples/upstream/pkg/handlers"
	"github.com/d0w/openshell-bff-examples/upstream/pkg/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/d0w/openshell-bff-examples/upstream/pkg/services"
)

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type Server struct {
	srv *http.Server
}

type Services struct {
	Gateway services.GatewayService
	Sandbox services.SandboxService
}

// Option customizes the router NewServer builds. It's handed the root
// chi.Router directly after upstream's own /gateway and /sandbox routes
// are mounted -- there's no flag or hidden mounting behavior to configure
// around. An Option decides everything about how its routes are grouped
// and what middleware runs on them: reuse upstream's exported
// middleware.RequireAuth, write an entirely different auth scheme, mix
// both across different route groups, or skip auth for a given route.
// Upstream never sees or cares which of these a given Option does.
type Option func(r chi.Router)

// WithRoutes is a small convenience Option for the common case: mount
// registrar's routes at pattern, running mws first (in order) as this
// route group's middleware stack. Pass middleware.RequireAuth to reuse
// upstream's auth, pass a custom middleware.Func to run downstream's own,
// pass both to layer them, or pass none to mount unauthenticated.
// Equivalent to calling r.Route(pattern, ...) directly, which any Option
// is also free to do for full control.
func WithRoutes(pattern string, registrar interface{ RegisterRoutes(r chi.Router) }, mws ...func(http.Handler) http.Handler) Option {
	return func(r chi.Router) {
		r.Route(pattern, func(r chi.Router) {
			r.Use(mws...)
			registrar.RegisterRoutes(r)
		})
	}
}

func NewServer(cfg ServerConfig, svcs Services, opts ...Option) *Server {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	gatewayHandler := handlers.NewGatewayHandler(svcs.Gateway)
	sandboxHandler := handlers.NewSandboxHandler(svcs.Sandbox)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.RequireAuth)

		r.Route("/gateway", gatewayHandler.RegisterRoutes)
		r.Route("/sandbox", sandboxHandler.RegisterRoutes)
	})

	for _, opt := range opts {
		opt(r)
	}

	httpServer := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{srv: httpServer}
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
