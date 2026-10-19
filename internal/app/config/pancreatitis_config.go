package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type JWTConfig struct {
	ExpiresIn time.Duration
	Secret    string
}

type RedisConfig struct {
	Host        string
	Port        int
	Password    string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type Config struct {
	ServiceHost string
	ServicePort int

	JWT   JWTConfig
	Redis RedisConfig
}

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	if cfg.JWT.Secret == "" {
		return nil, errors.New("JWT_SECRET не задан в .env")
	}
	if cfg.JWT.ExpiresIn <= 0 {
		return nil, errors.New("[JWT] ExpiresIn в config.toml должен быть больше 0 (например, \"1h\")")
	}

	cfg.Redis.Host = os.Getenv("REDIS_HOST")
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Redis.Port, err = strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		return nil, fmt.Errorf("REDIS_PORT должен быть числом: %w", err)
	}

	log.Info("config parsed")

	return cfg, nil
}
