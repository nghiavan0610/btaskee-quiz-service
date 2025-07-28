package response

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
)

type OffsetPagination struct {
	Page       int32 `json:"page"`
	Limit      int32 `json:"limit"`
	TotalItems int64 `json:"total_items"`
}

type CursorPagination struct {
	Next int64 `json:"next"`
}

type Response struct {
	Success bool        `json:"success"`
	Code    int         `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(http.StatusOK).JSON(Response{
		Success: true,
		Data:    data,
	})
}

func Error(c *fiber.Ctx, err *exception.AppError) error {
	errorData := map[string]interface{}{
		"code":    err.Code,
		"message": err.Message,
	}

	if err.Details != "" {
		errorData["details"] = err.Details
	}

	if len(err.Metadata) > 0 {
		errorData["metadata"] = err.Metadata
	}

	return c.Status(err.Status).JSON(Response{
		Success: false,
		Error:   errorData,
	})
}
