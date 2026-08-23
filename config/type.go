package config

// Config is the root configuration schema loaded from config.json via viper.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Owner    OwnerConfig    `mapstructure:"owner"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Port string `mapstructure:"port"`
}

// PostgresConfig holds Postgres connection settings.
type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig holds settings for signing and verifying access tokens.
type JWTConfig struct {
	Secret       string `mapstructure:"secret"`
	AccessTTLMin int    `mapstructure:"access_ttl_min"`
}

// OwnerConfig holds the credentials used to seed the owner account.
type OwnerConfig struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}
