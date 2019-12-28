// +build !js,!wasm

package wormhole_websocket

import (
	"bytes"
	"context"
	"github.com/themakers/wormhole/wormhole"
	"github.com/themakers/wormhole/wormhole/wire_io"
	"github.com/themakers/wormhole/wormhole_msgp"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// TODO https://github.com/gobwas/ws ???

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
			panic(err)
		}
		defer c.Close()

		if err := lp.(wormhole.LocalPeerTransport).HandleDataChannel(newWebSocketChan(q.Context(), "", c), nil); err != nil && err != wormhole.ErrPeerGone && err != context.Canceled {
			panic(err)
		} else {
		}
	})
}

func Connect(ctx context.Context, lp wormhole.LocalPeer, addr string, pcbs wormhole.PeerCallbacks) error {
	c, _, err := websocket.DefaultDialer.DialContext(ctx, addr, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	return lp.(wormhole.LocalPeerTransport).HandleDataChannel(newWebSocketChan(ctx, addr, c), pcbs)
}

func StayConnected(ctx context.Context, lp wormhole.LocalPeer, pcbs wormhole.PeerCallbacks, addr string) {
	for {
		(func() {
			defer func() {
				if rec := recover(); rec != nil {
					time.Sleep(1 * time.Second)
				}
			}()

			if err := Connect(ctx, lp, addr, pcbs); err != nil {
				panic(err)
			}

			select {
			case <-ctx.Done():
				return
			default:
			}

			time.Sleep(1 * time.Second)
		})()
	}
}

////////////////////////////////////////////////////////////////
//// Implementation
////

func newWebSocketChan(ctx context.Context, addr string, conn *websocket.Conn) wormhole.DataChannel {
	return &webSocketChan{
		ctx:  ctx,
		addr: addr,
		conn: conn,
		wfh:  wormhole_msgp.Handler,
	}
}

var _ wormhole.DataChannel = new(webSocketChan)

type webSocketChan struct {
	ctx  context.Context
	addr string
	conn *websocket.Conn
	wfh  wire_io.Handler
}

func (c *webSocketChan) MessageReader() (sz int, vr wire_io.ValueReader, err error) {
	_, data, err := c.conn.ReadMessage()

	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			return 0, nil, err
		} else {
			return 0, nil, wormhole.ErrPeerGone
		}
	}

	return c.wfh.NewReader(bytes.NewReader(data))
}

func (c *webSocketChan) MessageWriter(sz int, mw func(wire_io.ValueWriter) error) error {
	//wc, err := c.conn.NextWriter(websocket.BinaryMessage)
	//if err != nil {
	//	return err
	//}

	buf := bytes.NewBuffer([]byte{})

	if err := c.wfh.NewWriter(sz, buf, func(vw wire_io.ValueWriter) error {
		if err := mw(vw); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
}

func (c *webSocketChan) Close() error {
	if err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}

	return nil
}

func (c *webSocketChan) Context() context.Context {
	return c.ctx
}

func (c *webSocketChan) Addr() string {
	return c.addr
}
