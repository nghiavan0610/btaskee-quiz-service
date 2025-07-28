package handlers

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/dtos"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/guards"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/services"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/constants"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/middlewares"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/response"
)

type (
	GameHandler interface {
		RegisterRoutes(r fiber.Router)
	}

	gameHandler struct {
		sessionService services.SessionService
		authGuard      guards.AuthGuard
	}
)

var (
	gameHandlerOnce     sync.Once
	gameHandlerInstance GameHandler
)

func ProvideGameHandler(
	sessionService services.SessionService,
	authGuard guards.AuthGuard,
) GameHandler {
	gameHandlerOnce.Do(func() {
		gameHandlerInstance = &gameHandler{
			sessionService: sessionService,
			authGuard:      authGuard,
		}
	})
	return gameHandlerInstance
}

func (h *gameHandler) RegisterRoutes(r fiber.Router) {
	gameGroup := r.Group("/games")

	gameGroup.Post("/",
		middlewares.BodyValidator[dtos.CreateSessionRequest](),
		h.createSession,
	)

	gameGroup.Get("/join/:join_code",
		middlewares.PathParamsValidator[dtos.JoinSessionRequest](),
		h.joinSession,
	)
}

func (h *gameHandler) createSession(c *fiber.Ctx) error {
	// Get authenticated user (optional for guest hosts)
	authUser, _ := h.authGuard.GetAuthUser(c)

	req := middlewares.GetRequest[dtos.CreateSessionRequest](c, constants.KEY_REQ_BODY_PARAMS)

	res, appErr := h.sessionService.CreateSession(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *gameHandler) joinSession(c *fiber.Ctx) error {
	// Get authenticated user (optional for guest participants)
	authUser, _ := h.authGuard.GetAuthUser(c)

	req := middlewares.GetRequest[dtos.JoinSessionRequest](c, constants.KEY_REQ_PATH_PARAMS)

	res, appErr := h.sessionService.JoinSession(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}
