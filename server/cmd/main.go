package main

import (
	"context"
	"flag"
	"github.com/manx98/wss_port_forwarding/logger"
	"github.com/manx98/wss_port_forwarding/server/config"
	"github.com/manx98/wss_port_forwarding/server/transport"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
)

func main() {
	configFile := flag.String("c", "config.ini", "config file")
	flag.Parse()
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	err := transport.Run(ctx, config.LoadConfig(*configFile))
	if err != nil {
		logger.Fatal("failed to run proxy server", zap.Error(err))
	}
}
