package wormhole

import (
	"github.com/themakers/wormhole/wormhole/internal/base"
	"github.com/themakers/wormhole/wormhole/internal/remote_peer"
)

type (
	RemotePeer          remote_peer.RemotePeer
	RemotePeerGenerated remote_peer.RemotePeerGenerated
)

type DataChannel base.DataChannel

type WireFormatHandler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(m []byte) (interface{}, error)
}
