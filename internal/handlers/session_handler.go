package handlers

import (
	"strconv"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
	ws "github.com/nghiavan0610/btaskee-quiz-service/pkg/websocket"
)

type (
	SessionHandler interface {
		HandleWebSocket(conn *websocket.Conn)
	}

	sessionHandler struct {
		hub    ws.Hub
		logger *logger.Logger
	}
)

var (
	sessionHandlerOnce     sync.Once
	sessionHandlerInstance SessionHandler
)

func ProvideSessionHandler(
	hub ws.Hub,
	logger *logger.Logger,
) SessionHandler {
	sessionHandlerOnce.Do(func() {
		sessionHandlerInstance = &sessionHandler{
			hub:    hub,
			logger: logger,
		}
	})
	return sessionHandlerInstance
}

func (h *sessionHandler) HandleWebSocket(conn *websocket.Conn) {
	if conn == nil {
		h.logger.Error("WebSocket connection is nil", nil)
		return
	}

	sessionIDStr := conn.Params("session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid session ID in WebSocket connection", map[string]interface{}{
			"session_id_str": sessionIDStr,
			"error":          err.Error(),
		})
		conn.Close()
		return
	}

	client := ws.NewClient(conn, h.hub, sessionID)

	h.hub.RegisterClient(client)

	go client.WritePump() // Sends messages from hub to client

	// Run ReadPump in the main thread to keep the handler alive
	client.ReadPump()

}
