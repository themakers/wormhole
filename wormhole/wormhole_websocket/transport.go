package wormhole_websocket

import (
	"context"
	"github.com/themakers/wormhole/wormhole"
	"github.com/themakers/wormhole/wormhole/json_format"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func Acceptor(lp wormhole.LocalPeer) http.Handler {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(q *http.Request) bool {
			return true
		},
		EnableCompression: true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		c, err := upgrader.Upgrade(w, q, nil)
		if err != nil {
			lp.Log().Panic("error during websocket upgrade", zap.Error(err))
		}
		defer c.Close()

		if err := lp.(wormhole.LocalPeerTransport).HandleDataChannel(newWebSocketChan(q.Context(), lp.Log(), c)); err != nil {
			lp.Log().Panic("error serving websocket", zap.Error(err))
		}
	})
}

func Connect(ctx context.Context, lp wormhole.LocalPeer, addr string) error {
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	return lp.(wormhole.LocalPeerTransport).HandleDataChannel(newWebSocketChan(ctx, lp.Log(), c))
}

////////////////////////////////////////////////////////////////
//// Implementation
////

func newWebSocketChan(ctx context.Context, log *zap.Logger, conn *websocket.Conn) wormhole.DataChannel {
	return &webSocketChan{
		ctx:  ctx,
		conn: conn,
		wfh:  json_format.New(),
	}
}

var _ wormhole.DataChannel = new(webSocketChan)

type webSocketChan struct {
	ctx  context.Context
	conn *websocket.Conn
	wfh  wormhole.WireFormatHandler
}

func (c *webSocketChan) ReadMessage() (interface{}, error) {
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	log.Println(">>> incoming:", string(data))
	return c.wfh.Unmarshal(data)
}

func (c *webSocketChan) WriteMessage(m interface{}) error {
	data, err := c.wfh.Marshal(m)
	if err != nil {
		return err
	}
	log.Println(">>> outgoing:", string(data))
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *webSocketChan) Close() error {
	return c.conn.Close()
}

func (c *webSocketChan) Context() context.Context {
	return c.ctx
}
