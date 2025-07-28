package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/constants"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
	"github.com/nghiavan0610/btaskee-quiz-service/utils"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	ID            string
	SessionID     int64
	ParticipantID int64
	UserID        *int64
	Nickname      string
	IsHost        bool
	Conn          *websocket.Conn
	Hub           Hub
	Send          chan []byte
	mu            sync.RWMutex
}

type RoomMessage struct {
	RoomID        int64
	Message       []byte
	ExcludeClient *Client
}

type CrossServerMessage struct {
	ServerID      string    `json:"server_id"`
	RoomID        int64     `json:"room_id"`
	Message       []byte    `json:"message"`
	ExcludeClient string    `json:"exclude_client,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

type (
	Hub interface {
		RegisterClient(client *Client)
		UnregisterClient(client *Client)
		BroadcastToAll(message []byte)
		BroadcastToRoom(roomID int64, message []byte, excludeClient *Client)
		GetRoomClients(roomID int64) []*Client
		GetRoomClientCount(roomID int64) int
		SendToClient(client *Client, message []byte)
		GetLogger() *logger.Logger
		Run(ctx context.Context)

		// WebSocket message handler registration
		SetMessageHandler(handler WebSocketMessageHandler)
	}

	// WebSocketMessageHandler processes business logic for WebSocket messages
	WebSocketMessageHandler interface {
		HandleMessage(client *Client, wsMsg *models.WSMessage)
	}

	hub struct {
		// Client management
		roomClients map[int64]map[*Client]bool
		register    chan *Client
		unregister  chan *Client
		broadcast   chan RoomMessage
		mu          sync.RWMutex

		// Redis for pub/sub
		redisClient *redis.Client
		serverID    string
		ctx         context.Context
		cancel      context.CancelFunc

		// Message handling
		messageHandler WebSocketMessageHandler

		logger *logger.Logger
	}
)

var (
	hubOnce     sync.Once
	hubInstance Hub
)

func ProvideHub(logger *logger.Logger, redisClient *redis.Client) Hub {
	hubOnce.Do(func() {
		hubInstance = newHub(logger, redisClient)
	})
	return hubInstance
}

func newHub(logger *logger.Logger, redisClient *redis.Client) Hub {
	ctx, cancel := context.WithCancel(context.Background())

	hubServerID := fmt.Sprintf("server-%d-%s", time.Now().Unix(), utils.GenerateRandomString(8))

	hub := &hub{
		roomClients: make(map[int64]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan RoomMessage),
		redisClient: redisClient,
		serverID:    hubServerID,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
	}

	// Start Redis subscription
	go hub.subscribeToRedis()

	logger.Info("[HUB] Initialized WebSocket hub",
		"server_id", hub.serverID,
		"redis_host", redisClient.Options().Addr,
	)

	return hub
}

func (h *hub) Run(ctx context.Context) {
	defer h.logger.Info("WebSocket hub stopped")

	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case roomMsg := <-h.broadcast:
			h.handleRoomBroadcast(roomMsg)

		case <-ctx.Done():
			return
		}
	}
}

func (h *hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *hub) BroadcastToAll(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for roomID := range h.roomClients {
		h.BroadcastToRoom(roomID, message, nil)
	}
}

func (h *hub) BroadcastToRoom(roomID int64, message []byte, excludeClient *Client) {
	// 1. Broadcast to local clients
	h.broadcastToLocalRoom(roomID, message, excludeClient)

	// 2. Publish to Redis for other servers
	excludeClientID := ""
	if excludeClient != nil {
		excludeClientID = excludeClient.ID
	}

	crossServerMsg := &CrossServerMessage{
		ServerID:      h.serverID,
		RoomID:        roomID,
		Message:       message,
		ExcludeClient: excludeClientID,
		Timestamp:     time.Now(),
	}

	data, err := json.Marshal(crossServerMsg)
	if err != nil {
		h.logger.Error("Failed to marshal cross-server message", err)
		return
	}

	// Publish to Redis channel for this room
	channel := fmt.Sprintf("quiz:room:%d", roomID)
	if err := h.redisClient.Publish(h.ctx, channel, data).Err(); err != nil {
		h.logger.Error("Failed to publish to Redis", map[string]interface{}{
			"error":   err,
			"channel": channel,
			"room_id": roomID,
		})
	}
}

func (h *hub) GetRoomClients(roomID int64) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if roomClients, exists := h.roomClients[roomID]; exists {
		for client := range roomClients {
			clients = append(clients, client)
		}
	}
	return clients
}

func (h *hub) GetRoomClientCount(roomID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomClients, exists := h.roomClients[roomID]; exists {
		return len(roomClients)
	}
	return 0
}

func (h *hub) SendToClient(client *Client, message []byte) {
	if client == nil {
		return
	}

	select {
	case client.Send <- message:
	default:
		// Channel full, drop message silently
	}
}

func (h *hub) GetLogger() *logger.Logger {
	return h.logger
}

func (h *hub) SetMessageHandler(handler WebSocketMessageHandler) {
	h.messageHandler = handler
}

func (h *hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Initialize room if it doesn't exist
	if h.roomClients[client.SessionID] == nil {
		h.roomClients[client.SessionID] = make(map[*Client]bool)
	}

	// Add client to room
	h.roomClients[client.SessionID][client] = true

	// Track server presence for this room in Redis
	h.trackServerPresence(client.SessionID, true)
}

func (h *hub) handleUnregister(client *Client) {
	roomID := client.SessionID

	h.logger.Info("Hub handleUnregister called", map[string]interface{}{
		"client_id":      client.ID,
		"session_id":     roomID,
		"participant_id": client.ParticipantID,
		"nickname":       client.Nickname,
	})

	if h.messageHandler != nil {
		if gameHandler, ok := h.messageHandler.(interface{ HandleClientDisconnect(*Client) }); ok {
			h.logger.Info("Calling HandleClientDisconnect", map[string]interface{}{
				"client_id": client.ID,
			})
			gameHandler.HandleClientDisconnect(client)
		} else {
			h.logger.Warn("MessageHandler does not implement HandleClientDisconnect", map[string]interface{}{
				"handler_type": fmt.Sprintf("%T", h.messageHandler),
			})
		}
	} else {
		h.logger.Warn("No messageHandler registered")
	}

	// Remove client from room management
	h.mu.Lock()
	defer h.mu.Unlock()

	if roomClients, exists := h.roomClients[roomID]; exists {
		if _, exists := roomClients[client]; exists {
			delete(roomClients, client)
			close(client.Send)

			h.logger.Info("Client removed from room", map[string]interface{}{
				"client_id":         client.ID,
				"session_id":        roomID,
				"remaining_clients": len(roomClients),
			})

			// Clean up empty room
			if len(roomClients) == 0 {
				delete(h.roomClients, roomID)
				// Remove server presence if no more clients in this room
				h.trackServerPresence(roomID, false)
				h.logger.Info("Empty room cleaned up", map[string]interface{}{
					"session_id": roomID,
				})
			}
		} else {
			h.logger.Warn("Client not found in room", map[string]interface{}{
				"client_id":  client.ID,
				"session_id": roomID,
			})
		}
	} else {
		h.logger.Warn("Room not found", map[string]interface{}{
			"session_id": roomID,
		})
	}
}

func (h *hub) handleRoomBroadcast(roomMsg RoomMessage) {
	h.broadcastToLocalRoom(roomMsg.RoomID, roomMsg.Message, roomMsg.ExcludeClient)
}

func (h *hub) broadcastToLocalRoom(roomID int64, message []byte, excludeClient *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomClients, exists := h.roomClients[roomID]; exists {
		for client := range roomClients {
			if excludeClient != nil && client.ID == excludeClient.ID {
				continue
			}
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(roomClients, client)
			}
		}
	}
}

func (h *hub) subscribeToRedis() {
	pubsub := h.redisClient.PSubscribe(h.ctx, "quiz:room:*")
	defer pubsub.Close()

	for {
		select {
		case msg := <-pubsub.Channel():
			h.handleRedisMessage(msg)
		case <-h.ctx.Done():
			return
		}
	}
}

func (h *hub) handleRedisMessage(msg *redis.Message) {
	var crossServerMsg CrossServerMessage
	if err := json.Unmarshal([]byte(msg.Payload), &crossServerMsg); err != nil {
		h.logger.Error("Failed to unmarshal Redis message", err)
		return
	}

	// Skip if message came from this server
	if crossServerMsg.ServerID == h.serverID {
		return
	}

	// Find excluded client (if any) by ID
	var excludeClient *Client
	if crossServerMsg.ExcludeClient != "" {
		for _, client := range h.GetRoomClients(crossServerMsg.RoomID) {
			if client.ID == crossServerMsg.ExcludeClient {
				excludeClient = client
				break
			}
		}
	}

	// Broadcast to local clients only (don't re-publish to Redis)
	h.broadcastToLocalRoom(crossServerMsg.RoomID, crossServerMsg.Message, excludeClient)
}

// trackServerPresence tracks which servers have clients for which rooms
func (h *hub) trackServerPresence(roomID int64, isJoining bool) {
	key := fmt.Sprintf("quiz:room:%d:servers", roomID)

	if isJoining {
		// Add this server to the room's server set with expiration
		h.redisClient.SAdd(h.ctx, key, h.serverID)
		h.redisClient.Expire(h.ctx, key, constants.RedisPresenceExpiration*time.Minute)
	} else {
		// Remove this server from the room's server set
		h.redisClient.SRem(h.ctx, key, h.serverID)
	}
}

func NewClient(conn *websocket.Conn, hub Hub, sessionID int64) *Client {
	return &Client{
		ID:        utils.GenerateRandomHex(16),
		SessionID: sessionID,
		Conn:      conn,
		Hub:       hub,
		Send:      make(chan []byte, constants.ClientSendBufferSize),
	}
}

// SetParticipantInfo sets participant information for the client
func (c *Client) SetParticipantInfo(participantID int64, userID *int64, nickname string, isHost bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ParticipantID = participantID
	c.UserID = userID
	c.Nickname = nickname
	c.IsHost = isHost
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(constants.WebSocketPingInterval * time.Second)
	defer func() {
		if r := recover(); r != nil {
			c.Hub.GetLogger().Error("WritePump panic recovered", map[string]interface{}{
				"client_id": c.ID,
				"panic":     r,
			})
		}
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(constants.WebSocketWriteTimeout * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(constants.WebSocketWriteTimeout * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		if r := recover(); r != nil {
			c.Hub.GetLogger().Error("ReadPump panic recovered", map[string]interface{}{
				"client_id": c.ID,
				"panic":     r,
			})
		}
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(constants.WebSocketReadLimit)
	c.Conn.SetReadDeadline(time.Now().Add(constants.WebSocketPongTimeout * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(constants.WebSocketPongTimeout * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.GetLogger().Error("WebSocket unexpected close", map[string]interface{}{
					"client_id": c.ID,
					"error":     err.Error(),
				})
			}
			break
		}

		// Parse the message
		var wsMsg models.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			c.Hub.GetLogger().Error("Failed to parse WebSocket message", map[string]interface{}{
				"client_id": c.ID,
				"error":     err.Error(),
			})
			continue
		}

		// Delegate message handling to the registered handler (service layer)
		if hub, ok := c.Hub.(*hub); ok && hub.messageHandler != nil {
			hub.messageHandler.HandleMessage(c, &wsMsg)
		}
	}
}
