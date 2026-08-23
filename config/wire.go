package config

import "github.com/google/wire"

// ProviderSet exposes config providers to wire.
var ProviderSet = wire.NewSet(Load)
