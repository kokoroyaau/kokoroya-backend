package database

import "github.com/google/wire"

// ProviderSet exposes database providers to wire.
var ProviderSet = wire.NewSet(NewPostgresConnection, NewRedisClient)
