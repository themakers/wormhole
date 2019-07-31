package wormhole

import (
	"context"
	"github.com/themakers/wormhole/wormhole/internal/remote_peer"
	"sync"
)

/****************************************************************
** IFC LocalPeer
********/

func NewLocalPeer(cbs PeerCallbacks) LocalPeer {
	lp := &localPeer{
		cbs:   cbs,
		peers: map[string]*peerWatcher{},
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
	Close()
	__localPeer()
}

// Interface to be used by generated code
type LocalPeerGenerated interface {
	LocalPeer

	RegisterInterface(ifc string, constructor func(caller RemotePeer))
	WaitFor(waitCtx context.Context, id string) RemotePeer
}

// Interface to be used by transport implementations
type LocalPeerTransport interface {
	LocalPeer

	HandleDataChannel(ch DataChannel, pcbs PeerCallbacks) error
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

	peers     map[string]*peerWatcher
	peersLock sync.RWMutex
}

func (lp *localPeer) RegisterInterface(ifc string, constructor func(caller RemotePeer)) {
	lp.ctors = append(lp.ctors, constructor)
}

func (lp *localPeer) HandleDataChannel(dc DataChannel, pcbs PeerCallbacks) error {
	rp := remote_peer.NewRemotePeer(dc)

	defer rp.Close()

	for _, ctor := range lp.ctors {
		ctor(rp)
	}

	// FIXME
	if dc.Addr() != "" {
		defer lp.peerOnline(rp, dc.Addr())()
	}

	go lp.cbs.OnPeerConnected(rp)
	if pcbs != nil {
		go pcbs.OnPeerConnected(rp)
	}
	defer (func() {
		go lp.cbs.OnPeerDisconnected("")
		if pcbs != nil {
			go pcbs.OnPeerDisconnected("")
		}
	})()

	return rp.ReceiverWorker()
}

func (lp *localPeer) Close() {
	lp.peersLock.Lock()
	defer lp.peersLock.Unlock()

	for _, pw := range lp.peers {
		if pw.peer != nil {
			pw.peer.Close()
		}
	}
}

func (lp *localPeer) __localPeer() {}
