package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort  string
	RedisAddr   string
	IpRate      int
	TokenRate   int
	BanDuration time.Duration
}

func LoadConfig() Config {
	ipRate, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_IP"))
	tokenRate, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_TOKEN"))
	banDuration, _ := strconv.Atoi(os.Getenv("BAN_DURATION"))

	return Config{
		ServerPort:  os.Getenv("SERVER_PORT"),
		RedisAddr:   os.Getenv("REDIS_ADDR"),
		IpRate:      ipRate,
		TokenRate:   tokenRate,
		BanDuration: time.Duration(banDuration) * time.Second,
	}
}
