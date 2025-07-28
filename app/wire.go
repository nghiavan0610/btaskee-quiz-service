//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/nghiavan0610/btaskee-quiz-service/config"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/database"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/events"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/guards"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/handlers"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/repositories"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/services"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/cache"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/fiber"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/websocket"
)

var AppProviderSet = wire.NewSet(
	config.ConfigProviderSet,
	logger.ProvideLogger,
	database.DatabaseProviderSet,
	cache.ProvideCache,
	cache.ProvideRedisClient,
	fiber.NewFiber,
	guards.GuardProviderSet,
	websocket.WebSocketProviderSet,
	repositories.RepositoryProviderSet,
	services.ServiceProviderSet,
	events.EventHandlerProviderSet,
	handlers.HandlerProviderSet,
)

func AppFactory() (*ServiceApp, error) {
	wire.Build(
		AppProviderSet,
		NewServiceApp,
	)
	return &ServiceApp{}, nil
}
