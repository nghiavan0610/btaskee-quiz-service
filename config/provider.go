package config

import (
	"github.com/google/wire"
)

var ConfigProviderSet = wire.NewSet(
	ProvideConfig,
)
