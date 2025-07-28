package handlers

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nghiavan0610/btaskee-quiz-service/config"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/response"
)

type (
	HealthHandler interface {
		RegisterRoutes(r fiber.Router)
	}

	healthHandler struct {
		config *config.Config
	}
)

var (
	healthHandlerOnce     sync.Once
	healthHandlerInstance HealthHandler
)

func ProvideHealthHandler(config *config.Config) HealthHandler {
	healthHandlerOnce.Do(func() {
		healthHandlerInstance = &healthHandler{config}
	})
	return healthHandlerInstance
}

func (h *healthHandler) RegisterRoutes(r fiber.Router) {
	r.Get("/health", h.healthCheck)
}

func (h *healthHandler) healthCheck(ctx *fiber.Ctx) error {
	return response.Success(ctx, map[string]interface{}{
		"service": h.config.Server.ServiceName,
		"version": h.config.Server.ServiceVersion,
		"time":    time.Now().Format(time.RFC3339),
	})
}
