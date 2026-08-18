package services

import (
	"context"
	"fmt"
	"time"

	upstreamservices "github.com/d0w/openshell-bff-examples/upstream/pkg/services"
)

// SandboxUptimeService is downstream's OWN interface for the capability
// ExtendedService adds. Upstream has no notion of uptime, so there's no
// upstream interface to reuse here
type SandboxUptimeService interface {
	upstreamservices.SandboxService
	SandboxUptime(ctx context.Context, id string) (time.Duration, error)
}

// ExtendedService is the concrete implementation of SandboxUptimeService.
// It demonstrates interface embedding used purely to EXTEND upstream, as
// opposed to Service (sandbox.go) which OVERRIDES it.
//
// ExtendedService implements none of CreateSandbox/GetSandbox/
// ListSandboxes/DeleteSandbox itself -- all four fall straight through to
// the embedded upstreamservices.SandboxService. SandboxUptime is the only
// method it adds, and it has no upstream equivalent at all.
type ExtendedService struct {
	upstreamservices.SandboxService // embedded -- Create/Get/List/Delete inherited untouched
}

func NewExtendedService(base upstreamservices.SandboxService) *ExtendedService {
	return &ExtendedService{SandboxService: base}
}

// SandboxUptime is new functionality with no upstream equivalent
func (s *ExtendedService) SandboxUptime(ctx context.Context, id string) (time.Duration, error) {
	sb, err := s.GetSandbox(ctx, id) // inherited call, not overridden
	if err != nil {
		return 0, fmt.Errorf("sandbox uptime: %w", err)
	}
	return time.Since(sb.CreatedAt), nil
}
