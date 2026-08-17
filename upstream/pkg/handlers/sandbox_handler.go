package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	sandbox "github.com/d0w/openshell-bff-examples/upstream/pkg/services/sandbox"

	"github.com/go-chi/chi/v5"
)

// SandboxHandler exposes HTTP endpoints for sandbox CRUD operations.
type SandboxHandler struct {
	*Handler                              // gives access to h.logger and other shared fields
	sandboxService sandbox.SandboxService // service interface
}

// NewSandboxHandler constructs a SandboxHandler backed by the given service.
// Because it depends only on the SandboxService interface, any downstream
// decorator that satisfies the interface can be passed in here unmodified
// -- this handler never needs to be rewritten to pick up decorated behavior.
func NewSandboxHandler(svc sandbox.SandboxService) *SandboxHandler {
	return &SandboxHandler{
		Handler:        &Handler{logger: slog.Default()},
		sandboxService: svc,
	}
}

// RegisterRoutes wires up the sandbox routes onto the given router.
func (h *SandboxHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.CreateSandbox)
	r.Get("/", h.ListSandboxes)
	r.Get("/{id}", h.GetSandbox)
	r.Delete("/{id}", h.DeleteSandbox)
}

// CreateSandbox creates a new sandbox from a JSON payload.
func (h *SandboxHandler) CreateSandbox(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sb, err := h.sandboxService.CreateSandbox(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sb)
}

// GetSandbox returns a single sandbox by ID.
func (h *SandboxHandler) GetSandbox(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sb, err := h.sandboxService.GetSandbox(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sb)
}

// ListSandboxes returns all known sandboxes.
func (h *SandboxHandler) ListSandboxes(w http.ResponseWriter, r *http.Request) {
	list, err := h.sandboxService.ListSandboxes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// DeleteSandbox deletes a sandbox by ID.
func (h *SandboxHandler) DeleteSandbox(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.sandboxService.DeleteSandbox(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
