package transport

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/manx98/wss_port_forwarding/forwarding_client"
	server_config "github.com/manx98/wss_port_forwarding/server/config"
	"net"
)

func NewForwardingClient(ctx context.Context, server *server_config.ServerConfig, wsClient *websocket.Conn, local string) (*forwarding_client.ForwardingHandler, error) {
	tcpConn, err := net.Dial("tcp", local)
	if err != nil {
		return nil, fmt.Errorf("failed to tcpConn local: %w", err)
	}
	return forwarding_client.NewForwardingHandler(server.Password, wsClient, tcpConn, ctx, false), nil
}
