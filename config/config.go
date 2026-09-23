package config

import (
	"strings"

	"github.com/spf13/viper"
)

var envBindings = []string{
	"app.name",
	"app.env",
	"app.port",
	"postgres.host",
	"postgres.port",
	"postgres.user",
	"postgres.password",
	"postgres.db_name",
	"postgres.ssl_mode",
	"redis.host",
	"redis.port",
	"redis.password",
	"redis.db",
	"jwt.secret",
	"jwt.access_ttl_min",
	"owner.email",
	"owner.password",
	"email.resend_api_key",
	"email.from_address",
	"email.from_name",
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	for _, key := range envBindings {
		if err := v.BindEnv(key); err != nil {
			return nil, err
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
