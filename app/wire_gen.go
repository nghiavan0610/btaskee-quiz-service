// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

// Injectors from wire.go:

func AppFactory() (*ServiceApp, error) {
	configConfig := config.ProvideConfig()
	loggerLogger := logger.ProvideLogger(configConfig)
	app := fiber.NewFiber(loggerLogger, configConfig)
	databaseConnection, err := database.ProvideDatabase(configConfig, loggerLogger)
	if err != nil {
		return nil, err
	}
	cacheCache, err := cache.ProvideCache(configConfig, loggerLogger)
	if err != nil {
		return nil, err
	}
	client, err := cache.ProvideRedisClient(configConfig, loggerLogger)
	if err != nil {
		return nil, err
	}
	hub := websocket.ProvideHub(loggerLogger, client)
	healthHandler := handlers.ProvideHealthHandler(configConfig)
	queries := database.ProvideDatabaseQueries(databaseConnection)
	authRepository := repositories.ProvideAuthRepository(queries)
	tokenService := services.ProvideTokenService(cacheCache, configConfig)
	authService := services.ProvideAuthService(authRepository, tokenService, loggerLogger)
	authGuard := guards.ProvideAuthGuard(tokenService, cacheCache)
	authHandler := handlers.ProvideAuthHandler(authService, authGuard)
	userRepository := repositories.ProvideUserRepository(queries)
	userService := services.ProvideUserService(userRepository, tokenService, loggerLogger)
	userHandler := handlers.ProvideUserHandler(userService, authGuard)
	quizRepository := repositories.ProvideQuizRepository(queries)
	questionRepository := repositories.ProvideQuestionRepository(queries)
	validationService := services.ProvideValidationService(quizRepository, questionRepository)
	quizService := services.ProvideQuizService(loggerLogger, quizRepository, questionRepository, validationService)
	quizHandler := handlers.ProvideQuizHandler(quizService, authGuard)
	questionService := services.ProvideQuestionService(loggerLogger, questionRepository, quizRepository, validationService)
	questionHandler := handlers.ProvideQuestionHandler(configConfig, questionService, authGuard)
	sessionRepository := repositories.ProvideSessionRepository(queries)
	sessionService := services.ProvideSessionService(sessionRepository, quizRepository, questionRepository, loggerLogger)
	gameHandler := handlers.ProvideGameHandler(sessionService, authGuard)
	sessionHandler := handlers.ProvideSessionHandler(hub, loggerLogger)
	webSocketHandler := handlers.ProvideWebSocketHandler(sessionHandler)
	v := handlers.ProvideAppHandlers(healthHandler, authHandler, userHandler, quizHandler, questionHandler, gameHandler, webSocketHandler)
	gameEventHandler := events.ProvideGameEventHandler(sessionRepository, questionRepository, hub, loggerLogger)
	serviceApp := NewServiceApp(configConfig, loggerLogger, app, databaseConnection, cacheCache, hub, v, gameEventHandler)
	return serviceApp, nil
}

// wire.go:

var AppProviderSet = wire.NewSet(config.ConfigProviderSet, logger.ProvideLogger, database.DatabaseProviderSet, cache.ProvideCache, cache.ProvideRedisClient, fiber.NewFiber, guards.GuardProviderSet, websocket.WebSocketProviderSet, repositories.RepositoryProviderSet, services.ServiceProviderSet, events.EventHandlerProviderSet, handlers.HandlerProviderSet)
