package services

import (
	"context"
	"fmt"
	"log/slog"

	upstreamservices "github.com/d0w/openshell-bff-examples/upstream/pkg/services"
)

// Service decorates an upstream SandboxService.
type Service struct {
	upstreamservices.SandboxService // embedded -- inherits Get/List/Delete as-is
	logger                          *slog.Logger
}

func NewService(base upstreamservices.SandboxService) *Service {
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
func (s *Service) CreateSandbox(ctx context.Context, name string) (*upstreamservices.Sandbox, error) {
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
