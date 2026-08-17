// Package audit is a downstream-only capability that doesn't exist in
// upstream at all: aggregate stats over sandboxes. It's composed purely
// from upstream's existing SandboxService.ListSandboxes -- no upstream
// interface change was needed for this capability to exist.
//
// This is the third case from the architecture doc's decision table: "a
// genuinely new method/route upstream doesn't have." Unlike the
// CreateSandbox override in ../sandbox, this isn't a decorated service --
// it's a brand new handler, attached to upstream's server via the
// server.WithHandler extensibility hook rather than by upstream needing to
// know this domain exists.
package audit

import (
	"encoding/json"
	"net/http"

	upstreamsandbox "github.com/d0w/openshell-bff-examples/upstream/pkg/services/sandbox"

	"github.com/go-chi/chi/v5"
)

// Handler exposes audit endpoints over a SandboxService. It depends only
// on upstream's public SandboxService interface, so it works the same
// whether it's handed the plain upstream service or downstream's decorated
// one.
type Handler struct {
	sandboxService upstreamsandbox.SandboxService
}

// NewHandler constructs an audit Handler backed by the given service.
func NewHandler(svc upstreamsandbox.SandboxService) *Handler {
	return &Handler{sandboxService: svc}
}

// RegisterRoutes satisfies server.RouteRegistrar. Upstream's NewServer
// calls this the same way it calls its own handlers' RegisterRoutes --
// it has no idea "audit" exists as a concept.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/sandboxes", h.SandboxStats)
}

type sandboxStats struct {
	Total    int            `json:"total"`
	ByStatus map[string]int `json:"byStatus"`
}

// SandboxStats returns a count of sandboxes, broken down by status.
func (h *Handler) SandboxStats(w http.ResponseWriter, r *http.Request) {
	list, err := h.sandboxService.ListSandboxes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stats := sandboxStats{Total: len(list), ByStatus: map[string]int{}}
	for _, sb := range list {
		stats.ByStatus[sb.Status]++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
