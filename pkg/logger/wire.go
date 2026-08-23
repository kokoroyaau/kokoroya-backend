package logger

import "github.com/google/wire"

// ProviderSet exposes logger providers to wire.
var ProviderSet = wire.NewSet(New)
