package constants

// Score Calculation Constants
const (
	MaxQuestionScore  = 1000
	MinQuestionScore  = 100
	QuestionTimeLimit = 30
)

// WebSocket Connection Constants
const (
	WebSocketReadLimit    = 512 // Max message size in bytes
	WebSocketWriteTimeout = 10  // Write timeout in seconds
	WebSocketPingInterval = 54  // Ping interval in seconds
	WebSocketPongTimeout  = 60  // Pong timeout in seconds
	ClientSendBufferSize  = 256 // Client send channel buffer size
)

// Redis Constants
const (
	RedisPresenceExpiration = 30 // Server presence expiration in minutes
)
