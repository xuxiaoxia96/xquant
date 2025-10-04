package config

type XquantConfig struct {
	RedisConfig `mapstructure:",squash"`
	MySQLConfig `mapstructure:",squash"`
	EnvConfig   `mapstructure:",squash"`
}

type EnvConfig struct {
	IsProduct bool   `mapstructure:"IS_PRODUCT"`
	LogLevel  string `mapstructure:"LOG_LEVEL"`
}

type RedisConfig struct {
	RedisAddr string `mapstructure:"REDIS_ADDR"`
	RedisUser string `mapstructure:"REDIS_USER"`
	RedisPass string `mapstructure:"REDIS_PASS"`
}

type MySQLConfig struct {
	MySQLAddr string `mapstructure:"MySQL_ADDR"`
	MySQLUser string `mapstructure:"MySQL_USER"`
	MySQLPass string `mapstructure:"MySQL_PASS"`
}
