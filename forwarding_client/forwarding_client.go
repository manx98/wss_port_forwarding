package forwarding_client

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/manx98/wss_port_forwarding/utils"
	"net"
	"sync"
	"time"
)

type message struct {
	Type int    `json:"type"`
	Data []byte `json:"data"`
}

type ForwardingHandler struct {
	key       []byte
	wsClient  *websocket.Conn
	conn      net.Conn
	ctx       context.Context
	cancel    context.CancelCauseFunc
	lock      sync.Mutex
	isClient  bool
	writeChan chan *message
}

func (c *ForwardingHandler) handleWrite() error {
	for c.ctx.Err() == nil {
		_, p, err := c.wsClient.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read data from server: %w", err)
		}
		data, err := utils.DecryptAES(p, c.key)
		if err != nil {
			return fmt.Errorf("failed to decrypt data from remote: %w", err)
		}
		_, err = c.conn.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write data to local connect: %w", err)
		}
	}
	return context.Cause(c.ctx)
}

func (c *ForwardingHandler) handleWriteToWs() error {
	for c.ctx.Err() == nil {
		select {
		case msg := <-c.writeChan:
			data, err := utils.EncryptAES(msg.Data, c.key)
			if err != nil {
				return fmt.Errorf("failed to encrypt data from local: %w", err)
			}
			err = c.wsClient.WriteMessage(msg.Type, data)
			if err != nil {
				return fmt.Errorf("failed to write data to server: %w", err)
			}
		}
	}
	return context.Cause(c.ctx)
}

func (c *ForwardingHandler) handleRead() error {
	var buf []byte
	buf1 := make([]byte, 16*1024)
	buf2 := make([]byte, 16*1024)
	isBuf1 := false
	for c.ctx.Err() == nil {
		isBuf1 = !isBuf1
		if isBuf1 {
			buf = buf1
		} else {
			buf = buf2
		}
		n, err := c.conn.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read data from local connect: %w", err)
		}
		select {
		case c.writeChan <- &message{websocket.BinaryMessage, buf[:n]}:
		case <-c.ctx.Done():
		}
	}
	return context.Cause(c.ctx)
}

func (c *ForwardingHandler) Handler() error {
	defer c.cancel(context.Canceled)
	go func() {
		err := c.handleWriteToWs()
		c.cancel(err)
	}()
	go func() {
		err := c.handleWrite()
		c.cancel(err)
	}()
	go func() {
		err := c.handleRead()
		c.cancel(err)
	}()
	if c.isClient {
		go c.intervalPing()
	}
	<-c.ctx.Done()
	c.wsClient.Close()
	c.conn.Close()
	return context.Cause(c.ctx)
}

func (c *ForwardingHandler) PingHandler(appData string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.wsClient.WriteMessage(websocket.PongMessage, nil)
}

func (c *ForwardingHandler) PongHandler(appData string) error {
	return nil
}

func (c *ForwardingHandler) intervalPing() {
	for c.ctx.Err() == nil {
		select {
		case <-time.After(time.Second * 10):
			select {
			case c.writeChan <- &message{websocket.PingMessage, nil}:
			case <-c.ctx.Done():
			}
		case <-c.ctx.Done():
		}
	}
	return
}

func NewForwardingHandler(key []byte, wsClient *websocket.Conn, tcpConn net.Conn, ctx context.Context, isClient bool) *ForwardingHandler {
	h := &ForwardingHandler{
		key:       key,
		wsClient:  wsClient,
		conn:      tcpConn,
		isClient:  isClient,
		writeChan: make(chan *message),
	}
	h.wsClient.SetPongHandler(h.PingHandler)
	h.wsClient.SetPongHandler(h.PongHandler)
	h.ctx, h.cancel = context.WithCancelCause(ctx)
	return h
}
