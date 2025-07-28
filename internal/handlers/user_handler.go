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
	UserHandler interface {
		RegisterRoutes(r fiber.Router)
	}

	userHandler struct {
		userService services.UserService
		authGuard   guards.AuthGuard
	}
)

var (
	userHandlerOnce     sync.Once
	userHandlerInstance UserHandler
)

func ProvideUserHandler(userService services.UserService, authGuard guards.AuthGuard) UserHandler {
	userHandlerOnce.Do(func() {
		userHandlerInstance = &userHandler{userService, authGuard}
	})
	return userHandlerInstance
}

func (h *userHandler) RegisterRoutes(r fiber.Router) {
	userGroup := r.Group("/users")

	protectedGroup := userGroup.Group("", h.authGuard.AccessTokenGuard())
	protectedGroup.Get("/mine", h.getMine)
	protectedGroup.Put("/mine",
		middlewares.BodyValidator[dtos.UpdateMineRequest](),
		h.updateMine,
	)
}

func (h *userHandler) getMine(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	res, appErr := h.userService.GetMine(c.Context(), authUser)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *userHandler) updateMine(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	req := middlewares.GetRequest[dtos.UpdateMineRequest](c, constants.KEY_REQ_BODY_PARAMS)

	res, appErr := h.userService.UpdateMine(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}
