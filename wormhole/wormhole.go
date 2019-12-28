package wormhole

import (
	"errors"
	"github.com/themakers/wormhole/wormhole/internal/data_channel"
	"github.com/themakers/wormhole/wormhole/internal/remote_peer"
)

var ErrPeerGone = errors.New("peer gone")

type (
	RemotePeer          = remote_peer.RemotePeer
	RemotePeerGenerated = remote_peer.RemotePeerGenerated
)

type DataChannel = data_channel.DataChannel

type RegisterUnnamedRefFunc = remote_peer.RegisterUnnamedRefFunc
