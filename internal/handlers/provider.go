package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

type AppHandler interface {
	RegisterRoutes(r fiber.Router)
}

func ProvideAppHandlers(
	healthHandler HealthHandler,
	authHandler AuthHandler,
	userHandler UserHandler,
	quizHandler QuizHandler,
	questionHandler QuestionHandler,
	gameHandler GameHandler,
	webSocketHandler WebSocketHandler,
) []AppHandler {
	return []AppHandler{
		healthHandler,
		authHandler,
		userHandler,
		quizHandler,
		questionHandler,
		gameHandler,
		webSocketHandler,
	}
}

var HandlerProviderSet = wire.NewSet(
	ProvideAppHandlers,
	ProvideHealthHandler,
	ProvideAuthHandler,
	ProvideUserHandler,
	ProvideQuizHandler,
	ProvideQuestionHandler,
	ProvideSessionHandler,
	ProvideGameHandler,
	ProvideWebSocketHandler,
)
