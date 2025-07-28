package events

import "github.com/google/wire"

var EventHandlerProviderSet = wire.NewSet(
	ProvideGameEventHandler,
)
