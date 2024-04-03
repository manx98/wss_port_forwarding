package transport

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/manx98/wss_port_forwarding/forwarding_client"
	server_config "github.com/manx98/wss_port_forwarding/server/config"
	"github.com/manx98/wss_port_forwarding/utils"
)

func NewForwardingClient(ctx context.Context, server *server_config.ServerConfig, wsConn *websocket.Conn, local string) (*forwarding_client.ForwardingHandler, error) {
	tcpConn, err := utils.HandlerConn(local)
	if err != nil {
		return nil, fmt.Errorf("failed to tcpConn local: %w", err)
	}
	return forwarding_client.NewForwardingHandler(server.Password, wsConn, tcpConn, ctx, false), nil
}
