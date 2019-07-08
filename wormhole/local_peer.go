package wormhole

import "go.uber.org/zap"

/****************************************************************
** IFC LocalPeer
********/

func NewLocalPeer(log *zap.Logger, cbs PeerCallbacks) LocalPeer {
	lp := &localPeer{
		log: log,
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
	HandleDataChannel(ch DataChannel) error

	Log() *zap.Logger
}

type LocalPeerGenerated interface {
	RegisterInterface(ifc string, constructor func(caller RemotePeer))
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
	log *zap.Logger
	cbs PeerCallbacks

	ctors []func(RemotePeer)
}

func (lp *localPeer) RegisterInterface(ifc string, constructor func(caller RemotePeer)) {
	lp.ctors = append(lp.ctors, constructor)
}

func (lp *localPeer) HandleDataChannel(dc DataChannel) error {
	rp := newRemotePeer(lp.log, dc)

	defer rp.close()

	for _, ctor := range lp.ctors {
		ctor(rp)
	}

	go lp.cbs.OnPeerConnected(rp)
	defer (func() {
		lp.cbs.OnPeerDisconnected("")
	})()

	return rp.run()
}

func (lp *localPeer) Log() *zap.Logger { return lp.log }
