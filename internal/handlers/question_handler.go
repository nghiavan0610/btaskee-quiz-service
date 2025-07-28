package handlers

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/nghiavan0610/btaskee-quiz-service/config"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/dtos"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/guards"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/services"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/constants"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/middlewares"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/response"
)

type (
	QuestionHandler interface {
		RegisterRoutes(r fiber.Router)
	}

	questionHandler struct {
		config          *config.Config
		questionService services.QuestionService
		authGuard       guards.AuthGuard
	}
)

var (
	questionHandlerOnce     sync.Once
	questionHandlerInstance QuestionHandler
)

func ProvideQuestionHandler(
	config *config.Config,
	questionService services.QuestionService,
	authGuard guards.AuthGuard,
) QuestionHandler {
	questionHandlerOnce.Do(func() {
		questionHandlerInstance = &questionHandler{
			config:          config,
			questionService: questionService,
			authGuard:       authGuard,
		}
	})
	return questionHandlerInstance
}

func (h *questionHandler) RegisterRoutes(r fiber.Router) {
	questionGroup := r.Group("/questions", h.authGuard.AccessTokenGuard())

	questionGroup.Post("/",
		middlewares.RateLimit(middlewares.RateLimitConfig{
			RequestsPerSecond: 1,
			BurstSize:         5,
			KeyGenerator:      middlewares.DefaultKeyGenerator("question_create"),
		}),
		middlewares.BodyValidator[dtos.CreateQuestionRequest](),
		h.createQuestion,
	)
	questionGroup.Put("/index",
		middlewares.BodyValidator[dtos.UpdateQuestionIndexRequest](),
		h.updateQuestionIndex,
	)
	questionGroup.Put("/:question_id",
		middlewares.RateLimit(middlewares.RateLimitConfig{
			RequestsPerSecond: 1,
			BurstSize:         5,
			KeyGenerator:      middlewares.DefaultKeyGenerator("question_update"),
		}),
		middlewares.PayloadValidator[dtos.UpdateQuestionRequest](),
		h.updateQuestion,
	)

	questionGroup.Delete("/:question_id",
		middlewares.PathParamsValidator[dtos.DeleteQuestionRequest](),
		h.deleteQuestion,
	)
}

func (h *questionHandler) createQuestion(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	req := middlewares.GetRequest[dtos.CreateQuestionRequest](c, constants.KEY_REQ_BODY_PARAMS)

	res, appErr := h.questionService.CreateQuestion(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *questionHandler) updateQuestion(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	req := middlewares.GetRequest[dtos.UpdateQuestionRequest](c, constants.KEY_REQ_PAYLOAD_PARAMS)

	res, appErr := h.questionService.UpdateQuestion(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, res)
}

func (h *questionHandler) updateQuestionIndex(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	req := middlewares.GetRequest[dtos.UpdateQuestionIndexRequest](c, constants.KEY_REQ_BODY_PARAMS)

	appErr = h.questionService.UpdateQuestionIndex(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, true)
}

func (h *questionHandler) deleteQuestion(c *fiber.Ctx) error {
	authUser, appErr := h.authGuard.GetAuthUser(c)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	req := middlewares.GetRequest[dtos.DeleteQuestionRequest](c, constants.KEY_REQ_PATH_PARAMS)

	appErr = h.questionService.DeleteQuestion(c.Context(), authUser, req)
	if appErr != nil {
		return response.Error(c, appErr)
	}

	return response.Success(c, true)
}
