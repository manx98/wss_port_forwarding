package forwarding_client

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/manx98/wss_port_forwarding/utils"
	"net"
	"time"
)

type message struct {
	Type int    `json:"type"`
	Data []byte `json:"data"`
}

type ForwardingHandler struct {
	key       []byte
	wsConn    *websocket.Conn
	tcpConn   net.Conn
	ctx       context.Context
	cancel    context.CancelCauseFunc
	isClient  bool
	writeChan chan *message
}

func (c *ForwardingHandler) handleWrite() error {
	for c.ctx.Err() == nil {
		_, p, err := c.wsConn.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read data from server: %w", err)
		}
		data, err := utils.DecryptAES(p, c.key)
		if err != nil {
			return fmt.Errorf("failed to decrypt data from remote: %w", err)
		}
		_, err = c.tcpConn.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write data to local connect: %w", err)
		}
	}
	return context.Cause(c.ctx)
}

func (c *ForwardingHandler) handleWriteToWs() error {
	for c.ctx.Err() == nil {
		var msg *message
		if c.isClient {
			select {
			case <-time.After(time.Second * 10):
				msg = &message{websocket.PingMessage, nil}
			case msg = <-c.writeChan:
			case <-c.ctx.Done():
			}
		} else {
			select {
			case msg = <-c.writeChan:
			case <-c.ctx.Done():
			}
		}
		if msg != nil {
			data, err := utils.EncryptAES(msg.Data, c.key)
			if err != nil {
				return fmt.Errorf("failed to encrypt data from local: %w", err)
			}
			err = c.wsConn.WriteMessage(msg.Type, data)
			if err != nil {
				return fmt.Errorf("failed to write data to server: %w", err)
			}
		}
	}
	return context.Cause(c.ctx)
}

func (c *ForwardingHandler) handleTcpRead() error {
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
		n, err := c.tcpConn.Read(buf)
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
		err := c.handleTcpRead()
		c.cancel(err)
	}()
	<-c.ctx.Done()
	c.wsConn.Close()
	c.tcpConn.Close()
	return context.Cause(c.ctx)
}

func (c *ForwardingHandler) PongHandler(appData string) error {
	return nil
}

func NewForwardingHandler(key []byte, wsConn *websocket.Conn, tcpConn net.Conn, ctx context.Context, isClient bool) *ForwardingHandler {
	h := &ForwardingHandler{
		key:       key,
		wsConn:    wsConn,
		tcpConn:   tcpConn,
		isClient:  isClient,
		writeChan: make(chan *message),
	}
	h.wsConn.SetPongHandler(h.PongHandler)
	h.ctx, h.cancel = context.WithCancelCause(ctx)
	return h
}
