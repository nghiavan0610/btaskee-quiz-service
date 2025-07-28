package utils

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/constants"
)

func GetCIDFromCtx(ctx context.Context) string {
	cid := ctx.Value(constants.KEY_CORRELATION_ID)

	if cid == nil {
		return ""
	}

	return cid.(string)
}

func GetCIDFromFiberCtx(c *fiber.Ctx) string {
	cid := c.Locals(constants.KEY_CORRELATION_ID)

	if cid == nil {
		return ""
	}

	return cid.(string)
}

func ApplyCidToFiberCtx(c *fiber.Ctx) *fiber.Ctx {
	c.Locals(constants.KEY_CORRELATION_ID, uuid.New().String())

	return c
}
