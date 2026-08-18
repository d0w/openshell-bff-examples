package services

import "context"

// GatewayInfo is a minimal representation of gateway metadata.
type GatewayInfo struct {
	Name string `json:"name"`
}

// GatewayService defines the business logic for gateway operations.
type GatewayService interface {
	GetGatewayInfo(ctx context.Context) (*GatewayInfo, error)
	ProcessGateway(ctx context.Context, payload map[string]any) (map[string]any, error)
	GetGatewayByID(ctx context.Context, id string) (*GatewayInfo, error)
}

// DefaultGatewayService is the default implementation of GatewayService.
// Returned as a concrete type (accept interfaces, return structs); assigns
// implicitly to any GatewayService-typed field or param.
type DefaultGatewayService struct {
	// fields...
}

// NewDefaultGatewayService returns the default implementation of GatewayService.
func NewDefaultGatewayService() *DefaultGatewayService {
	return &DefaultGatewayService{}
}

func (g *DefaultGatewayService) GetGatewayInfo(ctx context.Context) (*GatewayInfo, error) {
	return &GatewayInfo{Name: "default-gateway"}, nil
}

func (g *DefaultGatewayService) ProcessGateway(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return payload, nil
}

func (g *DefaultGatewayService) GetGatewayByID(ctx context.Context, id string) (*GatewayInfo, error) {
	return &GatewayInfo{Name: id}, nil
}
