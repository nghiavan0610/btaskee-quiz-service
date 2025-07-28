package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nghiavan0610/btaskee-quiz-service/config"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/database"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/events"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/handlers"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/cache"
	fiberPkg "github.com/nghiavan0610/btaskee-quiz-service/pkg/fiber"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/websocket"
)

type ServiceApp struct {
	config      *config.Config
	logger      *logger.Logger
	server      *fiber.App
	database    *database.DatabaseConnection
	cache       cache.Cache
	hub         websocket.Hub
	handlers    []handlers.AppHandler
	gameHandler events.GameEventHandler // Add this to ensure it gets instantiated
}

func NewServiceApp(
	config *config.Config,
	logger *logger.Logger,
	server *fiber.App,
	database *database.DatabaseConnection,
	cache cache.Cache,
	hub websocket.Hub,
	handlers []handlers.AppHandler,
	gameHandler events.GameEventHandler, // Add this parameter
) *ServiceApp {
	return &ServiceApp{
		config:      config,
		logger:      logger,
		server:      server,
		database:    database,
		cache:       cache,
		hub:         hub,
		handlers:    handlers,
		gameHandler: gameHandler,
	}
}

func (app *ServiceApp) Start() {
	defer func() {
		if r := recover(); r != nil {
			app.logger.Error("Recovered from panic", r)
			app.shutdown()
		}
	}()

	// Start WebSocket hub
	hubCtx, hubCancel := context.WithCancel(context.Background())
	defer hubCancel()

	go app.hub.Run(hubCtx)

	fiberPkg.SetupRoutes(app.server, app.handlers, app.config)

	serverErr := make(chan error, 1)
	go func() {
		var err error
		addr := ":" + app.config.Server.Port

		if app.config.Server.GoEnv == "production" {
			err = app.server.ListenTLS(addr, "cert.pem", "key.pem")
		} else {
			err = app.server.Listen(addr)
		}

		serverErr <- err
	}()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		app.logger.Info("Received signal:", sig)
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			app.logger.Error("Server error:", err)
		}
	}

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.server.Shutdown(); err != nil {
		app.logger.Error("Fiber shutdown error:", err)
	}

	app.shutdown()
}

func (app *ServiceApp) shutdown() {
	app.logger.Info("Performing cleanup before shutdown...")

	if app.database != nil {
		app.logger.Info("Closing database connection...")
		app.database.Close()
	}

	if app.cache != nil {
		app.logger.Info("Closing cache connection...")
		app.cache.Close()
	}

	app.logger.Info("Shutdown complete")
}
