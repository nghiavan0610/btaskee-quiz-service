package guards

import (
	"github.com/google/wire"
)

var GuardProviderSet = wire.NewSet(
	ProvideAuthGuard,
)
