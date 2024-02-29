package transport

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/manx98/wss_port_forwarding/logger"
	"github.com/manx98/wss_port_forwarding/server/config"
	"go.uber.org/zap"
	"net/http"
)

const (
	UserKey     = "user"
	PasswordKey = "password"
	RemoteKey   = "remote"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Run(parent context.Context, server *config.ServerConfig) error {
	mux := http.NewServeMux()
	mux.HandleFunc(server.Path, func(writer http.ResponseWriter, request *http.Request) {
		handleWebSocket(parent, server, writer, request)
	})
	s := &http.Server{Addr: server.Bind, Handler: mux}
	go func() {
		<-parent.Done()
		s.Shutdown(context.Background())
	}()
	return s.ListenAndServe()
}

func handleWebSocket(ctx context.Context, server *config.ServerConfig, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(UserKey) != server.User && r.Header.Get(PasswordKey) != server.Password {
		w.WriteHeader(403)
		return
	}
	local := r.Header.Get(RemoteKey)
	if local == "" {
		w.WriteHeader(400)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("failed to upgrade websocket", zap.Error(err))
		return
	}
	defer conn.Close()
	client, err := NewForwardingClient(ctx, server, conn, local)
	if err != nil {
		logger.Error("failed to create forwarding client", zap.Error(err))
		return
	}
	logger.Debug("handler new proxy", zap.String("local", local), zap.String("user", r.RemoteAddr))
	err = client.Handler()
	if err != nil {
		logger.Error("failed to handle client request", zap.Error(err))
		return
	}
}
