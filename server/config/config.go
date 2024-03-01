package config

import (
	"github.com/manx98/wss_port_forwarding/logger"
	"github.com/manx98/wss_port_forwarding/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/ini.v1"
)

type ServerConfig struct {
	Bind     string `json:"bind"`
	Path     string `json:"path"`
	Password []byte
}

func LoadServerConfig(section *ini.Section) *ServerConfig {
	cfg := &ServerConfig{
		Bind:     section.Key("bind").String(),
		Path:     section.Key("path").String(),
		Password: utils.DeriveKey(section.Key("password").String()),
	}
	if cfg.Bind == "" {
		logger.Fatal("server config bind can't be empty!")
	}
	levelStr := section.Key("log_level").MustString("debug")
	var level zapcore.Level
	err := level.UnmarshalText([]byte(levelStr))
	if err != nil {
		logger.Fatal("server config log_level is invalid!", zap.String("log_level", levelStr))
	} else {
		logger.SetLogLevel(level)
	}
	return cfg
}

func LoadConfig(path string) *ServerConfig {
	load, err := ini.Load(path)
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}
	return LoadServerConfig(load.Section("server"))
}
