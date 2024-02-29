package transport

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/manx98/wss_port_forwarding/forwarding_client"
	server_config "github.com/manx98/wss_port_forwarding/server/config"
	"net"
)

func NewForwardingClient(ctx context.Context, server *server_config.ServerConfig, conn *websocket.Conn, local string) (*forwarding_client.ForwardingHandler, error) {
	dial, err := net.Dial("tcp", local)
	if err != nil {
		return nil, fmt.Errorf("failed to dial local: %w", err)
	}
	client := &forwarding_client.ForwardingHandler{
		WsClient: conn,
		Conn:     dial,
	}
	conn.SetPongHandler(client.PingHandler)
	conn.SetPongHandler(client.PongHandler)
	client.Ctx, client.Cancel = context.WithCancelCause(ctx)
	return client, nil
}
