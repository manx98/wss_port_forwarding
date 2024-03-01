package main

import (
	"context"
	"flag"
	"github.com/manx98/wss_port_forwarding/client/config"
	"github.com/manx98/wss_port_forwarding/client/transport"
	"github.com/manx98/wss_port_forwarding/logger"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
)

func main() {
	configFile := flag.String("c", "client_config.ini", "config file")
	flag.Parse()
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	err := transport.Run(ctx, config.LoadConfig(*configFile))
	if err != nil {
		logger.Fatal("failed to run proxy server", zap.Error(err))
	}
}
