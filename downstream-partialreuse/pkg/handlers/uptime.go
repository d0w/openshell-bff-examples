// UptimeHandler is a downstream-only capability with no upstream
// equivalent: reporting how long a sandbox has been running. It's the
// consumer of downstreamservices.SandboxUptimeService -- proof that
// interface is a real, usable contract and not just a marker on
// ExtendedService.
//
// Depending on SandboxUptimeService instead of the concrete
// *services.ExtendedService means this handler works against anything
// satisfying that interface -- ExtendedService today, a future
// alternative implementation, or a test double -- without ever importing
// a concrete type.
package handlers

import (
	"encoding/json"
	"net/http"

	downstreamservices "github.com/d0w/openshell-bff-examples/downstream-partialreuse/pkg/services"

	"github.com/go-chi/chi/v5"
)

type UptimeHandler struct {
	svc downstreamservices.SandboxUptimeService
}

func NewUptimeHandler(svc downstreamservices.SandboxUptimeService) *UptimeHandler {
	return &UptimeHandler{svc: svc}
}

func (h *UptimeHandler) RegisterRoutes(r chi.Router) {
	r.Get("/{id}", h.SandboxUptime)
}

func (h *UptimeHandler) SandboxUptime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	d, err := h.svc.SandboxUptime(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     id,
		"uptime": d.String(),
	})
}
