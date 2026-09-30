package handler

import (
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/server"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/service"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
	}
}
