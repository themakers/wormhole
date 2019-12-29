// +build !js,!wasm

package wormhole_websocket

import (
	"bytes"
	"context"
	"github.com/gorilla/websocket"
	"github.com/themakers/wormhole/wormhole"
	"github.com/themakers/wormhole/wormhole/wire_io"
	"github.com/themakers/wormhole/wormhole_msgp"
	"io"
	"net/http"
	"sync"
	"time"
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

var getBuffer = func() func() (*bytes.Buffer, func()) {
	pool := sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 1024))
			//return new(bytes.Buffer)
		},
	}
	return func() (*bytes.Buffer, func()) {
		buf := pool.Get().(*bytes.Buffer)
		buf.Reset()
		return buf, func() {
			pool.Put(buf)
		}
	}
}()

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

func (c *webSocketChan) MessageReader() (sz int, vr wire_io.ValueReader, done func(), err error) {
	_, r, err := c.conn.NextReader()

	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			return 0, nil, nil, err
		} else {
			return 0, nil, nil, wormhole.ErrPeerGone
		}
	}

	buf, done := getBuffer()
	defer done()

	if _, err := io.Copy(buf, r); err != nil {
		return 0, nil, nil, err
	}

	return c.wfh.NewReader(buf)
}

func (c *webSocketChan) MessageWriter(sz int, mw func(wire_io.ValueWriter) error) error {
	buf, done := getBuffer()
	defer done()

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

//BenchmarkBasic-8       	    5361	    210164 ns/op	   29703 B/op	     185 allocs/op
//BenchmarkHTTPBasic-8   	    8775	    143634 ns/op	    6400 B/op	      82 allocs/op

//BenchmarkBasic-8       	    5164	    235032 ns/op	   29473 B/op	     181 allocs/op
//BenchmarkHTTPBasic-8   	    8853	    132581 ns/op	    6391 B/op	      82 allocs/op
