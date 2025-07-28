package handlers

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type (
	WebSocketHandler interface {
		RegisterRoutes(r fiber.Router)
	}

	webSocketHandler struct {
		sessionHandler SessionHandler
	}
)

var (
	webSocketHandlerOnce     sync.Once
	webSocketHandlerInstance WebSocketHandler
)

func ProvideWebSocketHandler(
	sessionHandler SessionHandler,
) WebSocketHandler {
	webSocketHandlerOnce.Do(func() {
		webSocketHandlerInstance = &webSocketHandler{
			sessionHandler: sessionHandler,
		}
	})
	return webSocketHandlerInstance
}

func (h *webSocketHandler) RegisterRoutes(r fiber.Router) {
	wsGroup := r.Group("/ws")

	wsGroup.Use("/sessions/:session_id", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	wsGroup.Get("/sessions/:session_id", websocket.New(h.sessionHandler.HandleWebSocket))
}
