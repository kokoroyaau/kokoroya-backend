package config

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Owner    OwnerConfig    `mapstructure:"owner"`
	Email    EmailConfig    `mapstructure:"email"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Port string `mapstructure:"port"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret       string `mapstructure:"secret"`
	AccessTTLMin int    `mapstructure:"access_ttl_min"`
}

type OwnerConfig struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

type EmailConfig struct {
	ResendAPIKey string `mapstructure:"resend_api_key"`
	FromAddress  string `mapstructure:"from_address"`
	FromName     string `mapstructure:"from_name"`
}
