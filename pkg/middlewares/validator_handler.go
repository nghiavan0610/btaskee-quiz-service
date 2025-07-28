package middlewares

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/constants"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
	"github.com/nghiavan0610/btaskee-quiz-service/utils"
)

func GetRequest[T any](c *fiber.Ctx, key constants.ReqSourceType) *T {
	return c.Locals(string(key)).(*T)
}

func PathParamsValidator[T interface{}]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data T

		// fmt.Printf("[PathParamsValidator]:%T\n", data)

		err := c.ParamsParser(&data)
		if err != nil {
			logError(c, err)
			return exception.BadRequest(errors.CodeValidation, errors.ErrInvalidPathParams).WithDetails(err.Error())

		}

		err = utils.ValidateStruct(&data)
		if err != nil {
			logError(c, err)
			return exception.BadRequest(errors.CodeValidation, errors.ErrInvalidPathParams).WithMetadata("validation_errors", parseValidationError(err))
		}

		c.Locals(string(constants.KEY_REQ_PATH_PARAMS), &data)

		return c.Next()
	}
}

func QueryStringValidator[T interface{}]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data T

		// fmt.Printf("[QueryStringValidator]:%T\n", data)

		err := c.QueryParser(&data)
		if err != nil {
			logError(c, err)
			return exception.BadRequest(errors.CodeValidation, errors.ErrInvalidQueryParams).WithDetails(err.Error())
		}

		err = utils.ValidateStruct(&data)
		if err != nil {
			logError(c, err)
			return exception.BadRequest(errors.CodeValidation, errors.ErrInvalidQueryParams).WithMetadata("validation_errors", parseValidationError(err))
		}

		c.Locals(string(constants.KEY_REQ_QUERY_PARAMS), &data)

		return c.Next()
	}
}

func BodyValidator[T interface{}]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data T

		// fmt.Printf("[BodyValidator]:%T\n", data)

		err := c.BodyParser(&data)
		if err != nil {
			logError(c, err)
			return exception.BadRequest(errors.CodeValidation, errors.ErrInvalidBodyParams).WithDetails(err.Error())
		}

		err = utils.ValidateStruct(&data)
		if err != nil {
			logError(c, err)
			return exception.BadRequest(errors.CodeValidation, errors.ErrInvalidBodyParams).WithMetadata("validation_errors", parseValidationError(err))
		}

		c.Locals(string(constants.KEY_REQ_BODY_PARAMS), &data)

		return c.Next()
	}
}

func PayloadValidator[T interface{}]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data T

		_ = c.ParamsParser(&data)
		_ = c.QueryParser(&data)
		_ = c.BodyParser(&data)

		err := utils.ValidateStruct(&data)
		if err != nil {
			logError(c, err)
			return exception.BadRequest(errors.CodeValidation, errors.ErrInvalidPayloadParams).WithMetadata("validation_errors", parseValidationError(err))
		}

		c.Locals(string(constants.KEY_REQ_PAYLOAD_PARAMS), &data)

		return c.Next()
	}
}

func logError(c *fiber.Ctx, err error) {
	slog.Error(c.Method(), c.OriginalURL(), slog.Any("err", err))

	// Avoid log with sensitive data
	if !strings.Contains(c.Path(), "oidc") {
		fmt.Printf("Request Body %s\n", c.Body())
	}
}

func parseValidationError(err error) map[string]string {
	if err == nil {
		return nil
	}

	if errs, ok := err.(validator.ValidationErrors); ok {
		messages := make(map[string]string)
		for _, e := range errs {
			field := e.Field()
			tag := e.Tag()

			if fn, found := utils.ValidationMessages[tag]; found {
				messages[field] = fn(e)
			} else {
				messages[field] = e.Error()
			}
		}
		return messages
	}
	return map[string]string{"error": err.Error()}
}
