package nowire

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type RawDataChannel interface {
	ReadMessage() ([]byte, error)
	WriteMessage([]byte) error
	Close() error
}

func WebSocketAcceptor(lp LocalPeer) http.Handler {
	var upgrader = websocket.Upgrader{}

	return http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		c, err := upgrader.Upgrade(w, q, nil)
		if err != nil {
			lp.Log().Panic("Error during websocket upgrade", zap.Error(err))
		}
		defer c.Close()

		lp.HandleDataChannel(newWebSockerChan(q.Context(), lp.Log(), c))
	})
}

func WebSocketConnect(ctx context.Context, lp LocalPeer, addr string) error {
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	return lp.HandleDataChannel(newWebSockerChan(ctx, lp.Log(), c))
}

func newWebSockerChan(ctx context.Context, log *zap.Logger, conn *websocket.Conn) DataChannel {
	return NewJSONDataChannel(ctx, log, &webSocketChan{conn: conn})
}

type webSocketChan struct {
	conn *websocket.Conn
}

func (c *webSocketChan) ReadMessage() ([]byte, error) {
	_, data, err := c.conn.ReadMessage()
	return data, err
}

func (c *webSocketChan) WriteMessage(p []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, p)
}

func (c *webSocketChan) Close() error {
	return c.conn.Close()
}
