// Package audit is a downstream-only capability that doesn't exist in
// upstream at all: aggregate stats over sandboxes. It's composed purely
// from upstream's existing SandboxService.ListSandboxes -- no upstream
// interface change was needed for this capability to exist.
//
// This is the third case from the architecture doc's decision table: "a
// genuinely new method/route upstream doesn't have." Unlike the
// CreateSandbox override in ../services, this isn't a decorated service --
// it's a brand new handler, attached to upstream's server via the
// server.WithRoutes extensibility hook rather than by upstream needing to
// know this domain exists.
package audit

import (
	"encoding/json"
	"net/http"

	upstreamservices "github.com/d0w/openshell-bff-examples/upstream/pkg/services"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	sandboxService upstreamservices.SandboxService
}

func NewHandler(svc upstreamservices.SandboxService) *Handler {
	return &Handler{sandboxService: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/sandboxes", h.SandboxStats)
}

type sandboxStats struct {
	Total    int            `json:"total"`
	ByStatus map[string]int `json:"byStatus"`
}

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
