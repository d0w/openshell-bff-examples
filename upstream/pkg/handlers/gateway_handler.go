package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	gateway "github.com/d0w/openshell-bff-examples/upstream/pkg/services/gateway"

	"github.com/go-chi/chi/v5"
)

// GatewayHandler exposes HTTP endpoints for gateway operations.
type GatewayHandler struct {
	*Handler                              // gives access to h.logger and other shared fields
	gatewayService gateway.GatewayService // service interface
}

// NewGatewayHandler constructs a GatewayHandler backed by the given service.
func NewGatewayHandler(svc gateway.GatewayService) *GatewayHandler {
	return &GatewayHandler{
		Handler:        &Handler{logger: slog.Default()},
		gatewayService: svc,
	}
}

// RegisterRoutes wires up the gateway routes onto the given router.
func (h *GatewayHandler) RegisterRoutes(r chi.Router) {
	// r.Use(h.someGatewaySpecificMiddleware)

	r.Get("/", h.GetGatewayInfo)
	r.Post("/process", h.ProcessGateway)
	r.Get("/{id}", h.GetGatewayByID)
}

// GetGatewayInfo returns general information about the gateway.
func (h *GatewayHandler) GetGatewayInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.gatewayService.GetGatewayInfo(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// ProcessGateway accepts a JSON payload and passes it through the gateway service.
func (h *GatewayHandler) ProcessGateway(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Pass r.Context() down into the service for request scoped variables
	result, err := h.gatewayService.ProcessGateway(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetGatewayByID returns gateway info for a specific ID.
func (h *GatewayHandler) GetGatewayByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	info, err := h.gatewayService.GetGatewayByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
