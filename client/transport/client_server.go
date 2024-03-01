package transport

import (
	"context"
	"fmt"
	"github.com/manx98/wss_port_forwarding/client/config"
	"github.com/manx98/wss_port_forwarding/logger"
	server_config "github.com/manx98/wss_port_forwarding/server/config"
	"go.uber.org/zap"
	"net"
	"sync"
)

func Run(parent context.Context, cfg *config.ClientConfig) error {
	ctx, cancel := context.WithCancelCause(parent)
	wg := sync.WaitGroup{}
	for _, t := range cfg.Transport {
		wg.Add(1)
		go func(tCfg *config.Transport) {
			defer wg.Done()
			err := runTCPListener(ctx, cfg.Server, tCfg)
			if err != nil {
				cancel(err)
			}
		}(t)
	}
	wg.Wait()
	return context.Cause(ctx)
}

func runTCPListener(parent context.Context, server *server_config.ServerConfig, transport *config.Transport) error {
	logger.Debug("server tcp connect", zap.Any("transport", transport))
	listen, err := net.Listen("tcp", transport.Local)
	if err != nil {
		return fmt.Errorf("failed to listen %s: %w", transport.Local, err)
	}
	for parent.Err() == nil {
		accept, err01 := listen.Accept()
		if err01 != nil {
			return err01
		}
		logger.Debug("accept new connection", zap.Any("connect", accept.RemoteAddr().String()))
		go func() {
			err02 := handlerTcpConnection(parent, accept, server, transport.Remote)
			if err02 != nil {
				logger.Error("failed to handle connection", zap.String("remote", accept.RemoteAddr().String()), zap.Error(err02))
			}
		}()
	}
	return context.Cause(parent)
}

func handlerTcpConnection(ctx context.Context, conn net.Conn, server *server_config.ServerConfig, remote string) error {
	client, err := NewForwardingClient(ctx, server, conn, remote)
	if err != nil {
		conn.Close()
		return err
	}
	return client.Handler()
}
