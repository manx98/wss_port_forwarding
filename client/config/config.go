package config

import (
	"github.com/manx98/wss_port_forwarding/logger"
	"github.com/manx98/wss_port_forwarding/server/config"
	"go.uber.org/zap"
	"gopkg.in/ini.v1"
)

type Transport struct {
	Remote string `json:"remote"`
	Local  string `json:"local"`
}

type ClientConfig struct {
	Server    *config.ServerConfig
	Transport map[string]*Transport `json:"transport"`
}

func LoadConfig(path string) *ClientConfig {
	cfg, err := ini.Load(path)
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}
	clientCfg := &ClientConfig{
		Server:    config.LoadServerConfig(cfg.Section("common")),
		Transport: make(map[string]*Transport),
	}
	for _, sec := range cfg.Sections() {
		name := sec.Name()
		if name == "common" || name == ini.DefaultSection {
			continue
		}
		transport := &Transport{
			Remote: sec.Key("remote").String(),
			Local:  sec.Key("local").String(),
		}
		if transport.Remote == "" || transport.Local == "" {
			logger.Fatal("transport config check failed, remote and local can't be empty!", zap.Any(name, transport))
		}
		clientCfg.Transport[sec.Name()] = transport
	}
	return clientCfg
}
