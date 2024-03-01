package transport

import (
	"context"
	"encoding/base64"
	"github.com/gorilla/websocket"
	"github.com/manx98/wss_port_forwarding/logger"
	"github.com/manx98/wss_port_forwarding/server/config"
	"github.com/manx98/wss_port_forwarding/utils"
	"go.uber.org/zap"
	"net/http"
)

const (
	RemoteKey = "remote"
	RemoteMD5 = "remote_md5"
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
	local, err := r.Cookie(RemoteKey)
	if err != nil {
		logger.Error("failed to get remote", zap.String("request_remote", r.RemoteAddr), zap.Error(err))
		w.WriteHeader(400)
		return
	}
	remoteMd5Value, err := r.Cookie(RemoteMD5)
	if err != nil {
		logger.Error("failed to get remote_md5", zap.String("request_remote", r.RemoteAddr), zap.Error(err))
		w.WriteHeader(400)
		return
	}
	localData, err := base64.StdEncoding.DecodeString(local.Value)
	if err != nil {
		logger.Error("failed to decode remote", zap.String("request_remote", r.RemoteAddr), zap.String("local", local.Value), zap.Error(err))
		w.WriteHeader(400)
		return
	}
	localData, err = utils.DecryptAES(localData, server.Password)
	if err != nil {
		logger.Error("failed to decrypt remote", zap.String("request_remote", r.RemoteAddr), zap.String("local", local.Value), zap.Error(err))
		w.WriteHeader(400)
		return
	}
	realLocalMd5 := utils.MD5(localData)
	if realLocalMd5 != remoteMd5Value.Value {
		logger.Debug("request header remote_md5 is valid", zap.String("request_remote", r.RemoteAddr), zap.String("except_md5", realLocalMd5), zap.String("real_md5", remoteMd5Value.Value))
		w.WriteHeader(403)
		return
	}
	local.Value = string(localData)
	if local.Value == "" {
		logger.Error("failed to remote header is empty!", zap.String("request_remote", r.RemoteAddr), zap.String("local", local.Value))
		w.WriteHeader(400)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("failed to upgrade websocket", zap.Error(err))
		return
	}
	defer conn.Close()
	client, err := NewForwardingClient(ctx, server, conn, local.Value)
	if err != nil {
		logger.Error("failed to create forwarding client", zap.Error(err))
		return
	}
	logger.Debug("handler new proxy", zap.String("local", local.Value), zap.String("user", r.RemoteAddr))
	err = client.Handler()
	if err != nil {
		logger.Error("failed to handle client request", zap.Error(err))
		return
	}
}
