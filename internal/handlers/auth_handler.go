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
	AuthHandler interface {
		RegisterRoutes(r fiber.Router)
	}

	authHandler struct {
		authService services.AuthService
		authGuard   guards.AuthGuard
	}
)

var (
	authHandlerOnce     sync.Once
	authHandlerInstance AuthHandler
)

func ProvideAuthHandler(authService services.AuthService, authGuard guards.AuthGuard) AuthHandler {
	authHandlerOnce.Do(func() {
		authHandlerInstance = &authHandler{authService, authGuard}
	})
	return authHandlerInstance
}

func (h *authHandler) RegisterRoutes(r fiber.Router) {
	authGroup := r.Group("/auth")

	authGroup.Post("/signup",
		middlewares.BodyValidator[dtos.SignUpRequest](),
		h.signUp,
	)
	authGroup.Post("/signin",
		middlewares.BodyValidator[dtos.SignInRequest](),
		h.signIn,
	)
	authGroup.Get("/refresh",
		h.authGuard.RefreshTokenGuard(),
		h.refreshToken,
	)
	authGroup.Get("/signout",
		h.authGuard.AccessTokenGuard(),
		h.signOut,
	)
}

func (h *authHandler) signUp(c *fiber.Ctx) error {
	req := middlewares.GetRequest[dtos.SignUpRequest](c, constants.KEY_REQ_BODY_PARAMS)

	res, appErr := h.authService.SignUp(c.Context(), req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *authHandler) signIn(c *fiber.Ctx) error {
	req := middlewares.GetRequest[dtos.SignInRequest](c, constants.KEY_REQ_BODY_PARAMS)

	res, appErr := h.authService.SignIn(c.Context(), req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *authHandler) refreshToken(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	res, appErr := h.authService.RefreshToken(c.Context(), authUser)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *authHandler) signOut(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	h.authService.SignOut(c.Context(), authUser)

	return response.Success(c, true)
}
