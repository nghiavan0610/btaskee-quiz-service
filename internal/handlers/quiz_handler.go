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
	QuizHandler interface {
		RegisterRoutes(r fiber.Router)
	}

	quizHandler struct {
		quizService services.QuizService
		authGuard   guards.AuthGuard
	}
)

var (
	quizHandlerOnce     sync.Once
	quizHandlerInstance QuizHandler
)

func ProvideQuizHandler(
	quizService services.QuizService,
	authGuard guards.AuthGuard,
) QuizHandler {
	quizHandlerOnce.Do(func() {
		quizHandlerInstance = &quizHandler{
			quizService: quizService,
			authGuard:   authGuard,
		}
	})
	return quizHandlerInstance
}

func (h *quizHandler) RegisterRoutes(r fiber.Router) {
	quizGroup := r.Group("/quizzes")

	protectedGroup := quizGroup.Group("/mine", h.authGuard.AccessTokenGuard())

	protectedGroup.Post("/",
		middlewares.RateLimit(middlewares.RateLimitConfig{
			RequestsPerSecond: 0.05,
			BurstSize:         3,
			KeyGenerator:      middlewares.DefaultKeyGenerator("quiz_create"),
		}),
		middlewares.BodyValidator[dtos.CreateQuizRequest](),
		h.createQuiz,
	)
	protectedGroup.Put("/:quiz_id",
		middlewares.RateLimit(middlewares.RateLimitConfig{
			RequestsPerSecond: 0.05,
			BurstSize:         3,
			KeyGenerator:      middlewares.DefaultKeyGenerator("quiz_update"),
		}),
		middlewares.PayloadValidator[dtos.UpdateQuizRequest](),
		h.updateQuiz,
	)
	protectedGroup.Get("/:quiz_id",
		middlewares.PathParamsValidator[dtos.GetMyQuizDetailRequest](),
		h.getMyQuizDetail,
	)
	protectedGroup.Get("/",
		middlewares.QueryStringValidator[dtos.GetMyQuizListRequest](),
		h.getMyQuizList,
	)
	protectedGroup.Delete("/:quiz_id",
		middlewares.PathParamsValidator[dtos.DeleteQuizRequest](),
		h.deleteQuiz,
	)

	// Public routes (no auth required)
	quizGroup.Get("/",
		middlewares.QueryStringValidator[dtos.GetQuizListRequest](),
		h.getQuizList,
	)
	quizGroup.Get("/:quiz_id",
		middlewares.PathParamsValidator[dtos.GetQuizDetailRequest](),
		h.getQuizDetail,
	)

}

func (h *quizHandler) createQuiz(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	req := middlewares.GetRequest[dtos.CreateQuizRequest](c, constants.KEY_REQ_BODY_PARAMS)

	res, appErr := h.quizService.CreateQuiz(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *quizHandler) updateQuiz(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	req := middlewares.GetRequest[dtos.UpdateQuizRequest](c, constants.KEY_REQ_PAYLOAD_PARAMS)

	res, appErr := h.quizService.UpdateQuiz(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *quizHandler) getMyQuizDetail(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	req := middlewares.GetRequest[dtos.GetMyQuizDetailRequest](c, constants.KEY_REQ_PATH_PARAMS)

	res, appErr := h.quizService.GetMyQuizDetail(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *quizHandler) getMyQuizList(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	req := middlewares.GetRequest[dtos.GetMyQuizListRequest](c, constants.KEY_REQ_QUERY_PARAMS)

	res, appErr := h.quizService.GetMyQuizList(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *quizHandler) deleteQuiz(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	req := middlewares.GetRequest[dtos.DeleteQuizRequest](c, constants.KEY_REQ_PATH_PARAMS)

	appErr = h.quizService.DeleteQuiz(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, true)
}

func (h *quizHandler) getQuizList(c *fiber.Ctx) error {
	req := middlewares.GetRequest[dtos.GetQuizListRequest](c, constants.KEY_REQ_QUERY_PARAMS)

	res, appErr := h.quizService.GetQuizList(c.Context(), req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *quizHandler) getQuizDetail(c *fiber.Ctx) error {
	authUser, _ := h.authGuard.GetAuthUser(c)

	req := middlewares.GetRequest[dtos.GetQuizDetailRequest](c, constants.KEY_REQ_PATH_PARAMS)

	res, appErr := h.quizService.GetQuizDetail(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}
