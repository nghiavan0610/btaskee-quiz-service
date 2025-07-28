package websocket

import (
	"github.com/google/wire"
)

var WebSocketProviderSet = wire.NewSet(
	ProvideHub,
)
