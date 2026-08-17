// Package sandbox demonstrates interface decoration (see the architecture
// doc's "Service / Business Logic" section): downstream embeds upstream's
// SandboxService interface in its own struct and overrides only the one
// method it needs to change. GetSandbox, ListSandboxes, and DeleteSandbox
// are never redefined here -- calls to them fall straight through to the
// embedded upstream implementation. CreateSandbox is shadowed to layer on
// downstream-only policy before returning.
//
// This is the "behavior change on an existing method, same signature" case:
// no new interface is needed, and upstream's SandboxHandler can be reused
// completely unmodified since the interface's shape never changed.
package sandbox

import (
	"context"
	"fmt"
	"log/slog"

	upstreamsandbox "github.com/d0w/openshell-bff-examples/upstream/pkg/services/sandbox"
)

// Service decorates an upstream SandboxService.
type Service struct {
	upstreamsandbox.SandboxService // embedded -- inherits Get/List/Delete as-is
	logger *slog.Logger
}

// NewService wraps base (typically upstream's default implementation) with
// downstream-specific CreateSandbox behavior. The returned value still
// satisfies upstreamsandbox.SandboxService, so it can be handed to
// upstream's own NewSandboxHandler / server.Services without either of
// those knowing decoration happened.
func NewService(base upstreamsandbox.SandboxService) upstreamsandbox.SandboxService {
	return &Service{
		SandboxService: base,
		logger:         slog.Default(),
	}
}

// CreateSandbox shadows the embedded interface method. It delegates to
// upstream to actually create the sandbox, then applies a downstream-only
// policy (tagging + an extra "provisioning" gate upstream has no concept
// of) before returning -- proof the response was touched by downstream
// code, not just relayed from upstream untouched.
func (s *Service) CreateSandbox(ctx context.Context, name string) (*upstreamsandbox.Sandbox, error) {
	sb, err := s.SandboxService.CreateSandbox(ctx, name)
	if err != nil {
		return nil, err
	}

	// downstream-only policy: tag every sandbox it creates so it's
	// possible to tell, purely from the response body, which BFF handled
	// the request; also hold it in "provisioning" instead of upstream's
	// "running" to model an extra approval step that only exists here.
	sb.Name = fmt.Sprintf("%s (downstream-managed)", sb.Name)
	sb.Status = "provisioning"

	s.logger.Info("downstream: decorated CreateSandbox", "id", sb.ID, "name", sb.Name)
	return sb, nil
}
