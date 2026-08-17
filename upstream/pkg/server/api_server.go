package server

import (
	"context"
	"net/http"
	"time"

	"github.com/d0w/openshell-bff-examples/pkg/handlers"
	"github.com/d0w/openshell-bff-examples/pkg/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	gateway "github.com/d0w/openshell-bff-examples/pkg/services/gateway"
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
	Gateway gateway.GatewayService
}

func NewServer(cfg ServerConfig, svcs Services) *Server {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	gatewayHandler := handlers.NewGatewayHandler(svcs.Gateway)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.RequireAuth)

		r.Route("/gateway", gatewayHandler.RegisterRoutes)
	})

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
