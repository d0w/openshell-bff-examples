package services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Sandbox is a minimal representation of a sandbox resource.
type Sandbox struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// SandboxService defines the business logic for sandbox CRUD operations.
// Downstream consumers can decorate this interface (embed it in their own
// struct and override individual methods) without upstream ever knowing
// about the decoration -- see downstream-reuse/pkg/sandbox for an example.
type SandboxService interface {
	CreateSandbox(ctx context.Context, name string) (*Sandbox, error)
	GetSandbox(ctx context.Context, id string) (*Sandbox, error)
	ListSandboxes(ctx context.Context) ([]*Sandbox, error)
	DeleteSandbox(ctx context.Context, id string) error
}

// DefaultSandboxService is the default in-memory implementation of
// SandboxService. It's returned as a concrete type (accept interfaces,
// return structs) -- callers that only need the interface (e.g.
// server.Services, downstream decorators) get it for free via Go's
// implicit interface satisfaction, but callers who want the concrete type
// (e.g. to reach for a future exported method not on the interface) can
// have it too.
type DefaultSandboxService struct {
	mu        sync.Mutex
	sandboxes map[string]*Sandbox
	nextID    int
}

// NewDefaultSandboxService returns the default in-memory implementation of
// SandboxService.
func NewDefaultSandboxService() *DefaultSandboxService {
	return &DefaultSandboxService{
		sandboxes: make(map[string]*Sandbox),
	}
}

func (s *DefaultSandboxService) CreateSandbox(ctx context.Context, name string) (*Sandbox, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	sb := &Sandbox{
		ID:        fmt.Sprintf("sandbox-%d", s.nextID),
		Name:      name,
		Status:    "running",
		CreatedAt: time.Now(),
	}
	s.sandboxes[sb.ID] = sb

	return sb, nil
}

func (s *DefaultSandboxService) GetSandbox(ctx context.Context, id string) (*Sandbox, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sb, ok := s.sandboxes[id]
	if !ok {
		return nil, fmt.Errorf("sandbox %q not found", id)
	}
	return sb, nil
}

func (s *DefaultSandboxService) ListSandboxes(ctx context.Context) ([]*Sandbox, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]*Sandbox, 0, len(s.sandboxes))
	for _, sb := range s.sandboxes {
		out = append(out, sb)
	}
	return out, nil
}

func (s *DefaultSandboxService) DeleteSandbox(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sandboxes[id]; !ok {
		return fmt.Errorf("sandbox %q not found", id)
	}
	delete(s.sandboxes, id)
	return nil
}
