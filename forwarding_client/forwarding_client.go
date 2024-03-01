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

type ForwardingHandler struct {
	Key      []byte
	WsClient *websocket.Conn
	Conn     net.Conn
	Ctx      context.Context
	Cancel   context.CancelCauseFunc
	lock     sync.Mutex
	IsClient bool
}

func (c *ForwardingHandler) handleWrite() error {
	for c.Ctx.Err() == nil {
		_, p, err := c.WsClient.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read data from server: %w", err)
		}
		data, err := utils.DecryptAES(p, c.Key)
		if err != nil {
			return fmt.Errorf("failed to decrypt data from remote: %w", err)
		}
		_, err = c.Conn.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write data to local connect: %w", err)
		}
	}
	return context.Cause(c.Ctx)
}

func (c *ForwardingHandler) handleRead() error {
	buf := make([]byte, 16*1024)
	for c.Ctx.Err() == nil {
		n, err := c.Conn.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read data from local connect: %w", err)
		}
		data, err := utils.EncryptAES(buf[:n], c.Key)
		if err != nil {
			return fmt.Errorf("failed to encrypt data from local: %w", err)
		}
		err = c.WsClient.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			return fmt.Errorf("failed to write data to server: %w", err)
		}
	}
	return context.Cause(c.Ctx)
}

func (c *ForwardingHandler) Handler() error {
	defer c.Cancel(context.Canceled)
	go func() {
		err := c.handleWrite()
		c.Cancel(err)
	}()
	go func() {
		err := c.handleRead()
		c.Cancel(err)
	}()
	if c.IsClient {
		go func() {
			err := c.intervalPing()
			c.Cancel(fmt.Errorf("failed to send ping msg: %w", err))
		}()
	}
	<-c.Ctx.Done()
	c.WsClient.Close()
	c.Conn.Close()
	return context.Cause(c.Ctx)
}

func (c *ForwardingHandler) PingHandler(appData string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.WsClient.WriteMessage(websocket.PongMessage, nil)
}

func (c *ForwardingHandler) PongHandler(appData string) error {
	return nil
}

func (c *ForwardingHandler) sendPing() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.WsClient.WriteMessage(websocket.PingMessage, nil)
}

func (c *ForwardingHandler) intervalPing() error {
	for c.Ctx.Err() == nil {
		select {
		case <-time.After(time.Second * 10):
			err := c.sendPing()
			if err != nil {
				return err
			}
		case <-c.Ctx.Done():
		}
	}
	return context.Cause(c.Ctx)
}
