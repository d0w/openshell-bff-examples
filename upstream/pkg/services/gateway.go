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

type gatewayService struct {
	// fields...
}

// NewDefaultGatewayService returns the default in-memory implementation of GatewayService.
func NewDefaultGatewayService() GatewayService {
	return &gatewayService{}
}

func (g *gatewayService) GetGatewayInfo(ctx context.Context) (*GatewayInfo, error) {
	return &GatewayInfo{Name: "default-gateway"}, nil
}

func (g *gatewayService) ProcessGateway(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return payload, nil
}

func (g *gatewayService) GetGatewayByID(ctx context.Context, id string) (*GatewayInfo, error) {
	return &GatewayInfo{Name: id}, nil
}
