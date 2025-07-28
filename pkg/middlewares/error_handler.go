package middlewares

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	customErrors "github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/response"
)

func RecoveryHandler(log *logger.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch x := r.(type) {
				case string:
					err = errors.New(x)
				case error:
					err = x
				default:
					err = fmt.Errorf("unknown panic: %v", r)
				}

				log.ErrorFields("Panic recovered", map[string]interface{}{
					"error":  err.Error(),
					"path":   ctx.Path(),
					"method": ctx.Method(),
					"ip":     ctx.IP(),
				})

				appErr := exception.InternalError(customErrors.CodeInternal, customErrors.ErrInternal)

				errorHandler := ErrorHandler(log)
				errorHandler(ctx, appErr)
			}
		}()

		return ctx.Next()
	}
}

func NotFoundHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		appErr := exception.NotFound(customErrors.CodeNotFound, customErrors.ErrRouteNotFound).
			WithMetadata("path", ctx.Path()).
			WithMetadata("method", ctx.Method())

		return response.Error(ctx, appErr)
	}
}

func ErrorHandler(log *logger.Logger) fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		appErr := exception.InternalError(customErrors.CodeInternal, customErrors.ErrInternal)

		switch e := err.(type) {
		case *exception.AppError:
			appErr = e
			logAppError(log, e, ctx)

		case *fiber.Error:
			// Handle Fiber framework errors
			appErr = exception.NewAppError(e.Code, customErrors.ErrHTTP, e.Message)

			fiberLogData := map[string]interface{}{
				"status":  e.Code,
				"message": e.Message,
				"path":    ctx.Path(),
				"method":  ctx.Method(),
				"ip":      ctx.IP(),
			}

			if e.Code >= 400 && e.Code < 500 {
				log.WarnFields("Fiber client error occurred", fiberLogData)
			} else {
				log.ErrorFields("Fiber server error occurred", fiberLogData)
			}

		default:
			log.Error(customErrors.ErrUnexpected, err)
		}

		appErr.WithMetadata("path", ctx.Path()).
			WithMetadata("method", ctx.Method()).
			WithMetadata("timestamp", time.Now().Unix())

		return response.Error(ctx, appErr)
	}
}

func logAppError(log *logger.Logger, appErr *exception.AppError, ctx *fiber.Ctx) {
	logData := map[string]interface{}{
		"code":    appErr.Code,
		"message": appErr.Message,
		"status":  appErr.Status,
		"path":    ctx.Path(),
		"method":  ctx.Method(),
		"ip":      ctx.IP(),
	}

	for key, value := range appErr.Metadata {
		logData[key] = value
	}

	switch {
	case appErr.Status >= 400 && appErr.Status < 500:
		// Client errors - log as warnings (no stack trace)
		log.WarnFields("Client error occurred", logData)

	case appErr.Status == 429:
		// Rate limiting - log as info
		log.Info("Rate limit exceeded", logData)

	case appErr.Status >= 500:
		// Server errors - log as errors (with stack trace)
		log.ErrorFields("Server error occurred", logData)

	default:
		// Unknown status - log as error
		log.ErrorFields("Unknown error occurred", logData)
	}
}
