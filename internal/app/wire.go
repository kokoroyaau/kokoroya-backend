package app

import "github.com/google/wire"

// ProviderSet exposes app providers to wire.
var ProviderSet = wire.NewSet(NewApp)
