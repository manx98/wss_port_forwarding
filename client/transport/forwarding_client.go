package transport

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/manx98/wss_port_forwarding/forwarding_client"
	server_config "github.com/manx98/wss_port_forwarding/server/config"
	"github.com/manx98/wss_port_forwarding/server/transport"
	"net"
	"net/http"
	"net/url"
)

func NewForwardingClient(ctx context.Context, server *server_config.ServerConfig, conn net.Conn, remote string) (*forwarding_client.ForwardingHandler, error) {
	u := url.URL{Scheme: "ws", Host: server.Bind, Path: server.Path}
	c, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), http.Header{
		transport.UserKey:     []string{server.User},
		transport.PasswordKey: []string{server.Password},
		transport.RemoteKey:   []string{remote},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}
	client := &forwarding_client.ForwardingHandler{
		WsClient: c,
		Conn:     conn,
		IsClient: true,
	}
	c.SetPongHandler(client.PingHandler)
	c.SetPongHandler(client.PongHandler)
	client.Ctx, client.Cancel = context.WithCancelCause(ctx)
	return client, nil
}
