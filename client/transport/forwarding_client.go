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
	wsConn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), http.Header{
		"Cookie": {
			fmt.Sprintf("%s=%s", transport.RemoteKey, base64.StdEncoding.EncodeToString(remoteData)),
			fmt.Sprintf("%s=%s", transport.RemoteMD5, utils.MD5([]byte(remote))),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}
	return forwarding_client.NewForwardingHandler(server.Password, wsConn, conn, ctx, true), nil
}
