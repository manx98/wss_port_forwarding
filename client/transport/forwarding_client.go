package transport

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/manx98/wss_port_forwarding/forwarding_client"
	server_config "github.com/manx98/wss_port_forwarding/server/config"
	"github.com/manx98/wss_port_forwarding/server/transport"
	"github.com/manx98/wss_port_forwarding/utils"
	"net"
	"net/http"
	"net/url"
)

func NewForwardingClient(ctx context.Context, server *server_config.ServerConfig, conn net.Conn, remote string) (*forwarding_client.ForwardingHandler, error) {
	u := url.URL{Scheme: "ws", Host: server.Bind, Path: server.Path}
	remoteData, err := utils.EncryptAES([]byte(remote), server.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt remote info: %w", err)
	}
	c, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), http.Header{
		"Cookie": {
			fmt.Sprintf("%s=%s", transport.RemoteKey, base64.StdEncoding.EncodeToString(remoteData)),
			fmt.Sprintf("%s=%s", transport.RemoteMD5, utils.MD5([]byte(remote))),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}
	client := &forwarding_client.ForwardingHandler{
		Key:      server.Password,
		WsClient: c,
		Conn:     conn,
		IsClient: true,
	}
	c.SetPongHandler(client.PingHandler)
	c.SetPongHandler(client.PongHandler)
	client.Ctx, client.Cancel = context.WithCancelCause(ctx)
	return client, nil
}
