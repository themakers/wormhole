// +build !js,!wasm

package wormhole_websocket

import (
	"context"
	"github.com/themakers/wormhole/wormhole"
	"github.com/themakers/wormhole/wormhole/json_format"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
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
			panic(err)
		}
		defer c.Close()

		if err := lp.(wormhole.LocalPeerTransport).HandleDataChannel(newWebSocketChan(q.Context(), "", c), nil); err != nil && err != wormhole.ErrPeerGone {
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

func StayConnected(ctx context.Context, lp wormhole.LocalPeer, addr string) {
	for {
		(func() {
			defer func() {
				if rec := recover(); rec != nil {
					time.Sleep(1 * time.Second)
				}
			}()

			if err := Connect(ctx, lp, addr, wormhole.NewPeerCallbacks(func(peer wormhole.RemotePeer) {

			}, func(id string) {

			})); err != nil {
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

//func ConnectFGSDFH(ctx context.Context, lp wormhole.LocalPeer, addr string) (wormhole.RemotePeer, error) {
//	ctx, cancel := context.WithCancel(ctx)
//
//	//dialer := &websocket.Dialer{}
//
//	c, _, err := websocket.DefaultDialer.DialContext(ctx, addr, nil)
//	if err != nil {
//		res.Err <- err
//		cancel()
//		return
//	}
//	defer c.Close()
//
//	go (func() {
//		defer cancel()
//		if err := lp.(wormhole.LocalPeerTransport).HandleDataChannel(
//			newWebSocketChan(ctx, c),
//			wormhole.NewPeerCallbacks(func(peer wormhole.RemotePeer) {
//				res.Peer <- peer
//			}, func(id string) {
//				cancel()
//			}),
//		); err != nil {
//			res.Err <- err
//		} else {
//			res.Err <- nil
//		}
//	})()
//
//	return
//}

////////////////////////////////////////////////////////////////
//// Implementation
////

func newWebSocketChan(ctx context.Context, addr string, conn *websocket.Conn) wormhole.DataChannel {
	return &webSocketChan{
		ctx:  ctx,
		addr: addr,
		conn: conn,
		wfh:  json_format.New(),
	}
}

var _ wormhole.DataChannel = new(webSocketChan)

type webSocketChan struct {
	ctx  context.Context
	addr string
	conn *websocket.Conn
	wfh  wormhole.WireFormatHandler
}

func (c *webSocketChan) ReadMessage() (interface{}, error) {
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			return nil, err
		} else {
			return nil, wormhole.ErrPeerGone
		}
	}

	return c.wfh.Unmarshal(data)
}

func (c *webSocketChan) WriteMessage(m interface{}) error {
	data, err := c.wfh.Marshal(m)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.BinaryMessage, data)
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
