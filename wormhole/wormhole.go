package wormhole

import (
	"errors"
	"github.com/themakers/wormhole/wormhole/internal/base"
	"github.com/themakers/wormhole/wormhole/internal/remote_peer"
)

var ErrPeerGone = errors.New("peer gone")

type (
	RemotePeer          remote_peer.RemotePeer
	RemotePeerGenerated remote_peer.RemotePeerGenerated
)

type DataChannel base.DataChannel

type WireFormatHandler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(m []byte) (interface{}, error)
}
