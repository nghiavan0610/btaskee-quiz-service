package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/nghiavan0610/btaskee-quiz-service/config"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/handlers"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/middlewares"
	"github.com/nghiavan0610/btaskee-quiz-service/utils"
)

func NewFiber(log *logger.Logger, config *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middlewares.ErrorHandler(log),
	})

	app.Use(helmet.New(helmet.Config{
		CrossOriginEmbedderPolicy: "false",
		ContentSecurityPolicy:     "false",
		CrossOriginOpenerPolicy:   "cross-origin",
		CrossOriginResourcePolicy: "cross-origin",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.CORS.GetAllowOrigins(),
		AllowMethods:     config.CORS.AllowMethods,
		AllowHeaders:     config.CORS.AllowHeaders,
		AllowCredentials: config.CORS.AllowCredentials,
	}))

	app.Use(favicon.New())
	app.Use(fiberLogger.New())

	app.Use(middlewares.RecoveryHandler(log))

	app.Use(middlewares.RateLimit(middlewares.RateLimitConfig{
		RequestsPerSecond: config.RateLimit.RPS,
		BurstSize:         config.RateLimit.Burst,
		KeyGenerator:      middlewares.DefaultKeyGenerator("general"),
	}))

	// Serve static files for demo client
	// app.Static("/", "./static")

	app.Use(func(c *fiber.Ctx) error {
		c = utils.ApplyCidToFiberCtx(c)
		return c.Next()
	})

	return app
}

func SetupRoutes(app *fiber.App, handlers []handlers.AppHandler, config *config.Config) {
	apiV1 := app.Group("/api/" + config.Server.ServiceVersion)

	for _, handler := range handlers {
		if handler != nil {
			handler.RegisterRoutes(apiV1)
		}
	}

	app.Use(middlewares.NotFoundHandler())
}
