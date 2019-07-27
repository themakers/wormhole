package wormhole

import (
	"github.com/themakers/wormhole/wormhole/internal/remote_peer"
)

/****************************************************************
** IFC LocalPeer
********/

func NewLocalPeer(cbs PeerCallbacks) LocalPeer {
	lp := &localPeer{
		cbs: cbs,
	}

	if lp.cbs == nil {
		lp.cbs = NewPeerCallbacks(
			func(rp RemotePeer) {},
			func(id string) {},
		)
	}

	return lp
}

type LocalPeer interface {
	__localPeer()
}

type LocalPeerGenerated interface {
	LocalPeer

	RegisterInterface(ifc string, constructor func(caller RemotePeer))
}

type LocalPeerTransport interface {
	LocalPeer

	HandleDataChannel(ch DataChannel) error
}

type PeerCallbacks interface {
	OnPeerConnected(peer RemotePeer)
	OnPeerDisconnected(id string)
}

func NewPeerCallbacks(opc func(peer RemotePeer), opd func(id string)) PeerCallbacks {
	return &peerCallbacksWrapper{opc: opc, opd: opd}
}

type peerCallbacksWrapper struct {
	opc func(peer RemotePeer)
	opd func(id string)
}

func (pcw *peerCallbacksWrapper) OnPeerConnected(peer RemotePeer) { pcw.opc(peer) }
func (pcw *peerCallbacksWrapper) OnPeerDisconnected(id string)    { pcw.opd(id) }

/****************************************************************
** IMPL LocalPeer
********/

var (
	_ LocalPeer          = new(localPeer)
	_ LocalPeerGenerated = new(localPeer)
)

type localPeer struct {
	cbs PeerCallbacks

	ctors []func(peer RemotePeer)
}

func (lp *localPeer) RegisterInterface(ifc string, constructor func(caller RemotePeer)) {
	lp.ctors = append(lp.ctors, constructor)
}

func (lp *localPeer) HandleDataChannel(dc DataChannel) error {
	rp := remote_peer.NewRemotePeer(dc)

	defer rp.Close()

	for _, ctor := range lp.ctors {
		ctor(rp)
	}

	go lp.cbs.OnPeerConnected(rp)
	defer (func() {
		go lp.cbs.OnPeerDisconnected("")
	})()

	return rp.ReceiverWorker()
}

func (lp *localPeer) __localPeer() {}
