package config

import "github.com/spf13/viper"

type Config struct {
	DBUrl      string `mapstructure:"DB_URL"`
	ServerPort string `mapstructure:"SERVER_PORT"`
	Env        string `mapstructure:"ENV"`
}

func Load() (Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("ENV", "development")

	if err := viper.ReadInConfig(); err != nil {
	}

	var cfg Config
	err := viper.Unmarshal(&cfg)
	return cfg, err
}
