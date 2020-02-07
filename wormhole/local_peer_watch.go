package wormhole

import (
	"context"
	"errors"
	"sync"

	"github.com/rs/xid"
)

type peerWatcher struct {
	lock  sync.Mutex
	peer  RemotePeer
	chans map[string]chan RemotePeer

	//> Protected by localPeer.peersLock
	rc int
}

func (pw *peerWatcher) set(peer RemotePeer) {
	pw.lock.Lock()
	defer pw.lock.Unlock()

	if peer == pw.peer {
		return
	}

	pw.peer = peer

	for _, ch := range pw.chans {
		ch <- pw.peer
	}
}

func (pw *peerWatcher) ch() (chan RemotePeer, func()) {
	pw.lock.Lock()
	defer pw.lock.Unlock()

	ch := make(chan RemotePeer, 32)

	chid := xid.New().String()

	pw.chans[chid] = ch

	if pw.peer != nil {
		ch <- pw.peer
	}

	return ch, func() {
		pw.lock.Lock()
		defer pw.lock.Unlock()

		delete(pw.chans, chid)
	}
}

func (lp *localPeer) watchPeer(id string) (*peerWatcher, func()) {
	lp.peersLock.Lock()
	defer lp.peersLock.Unlock()

	pw, ok := lp.peers[id]
	if !ok {
		pw = &peerWatcher{rc: 0, chans: map[string]chan RemotePeer{}}
		lp.peers[id] = pw
	}

	pw.rc += 1

	return pw, func() {
		lp.peersLock.Lock()
		defer lp.peersLock.Unlock()

		pw.rc -= 1

		if pw.rc <= 0 {
			delete(lp.peers, id)
		}
	}
}

func (lp *localPeer) peerOnline(peer RemotePeer, id string) func() {
	pw, release := lp.watchPeer(id)

	pw.set(peer)

	return func() {
		pw.set(nil)
		release()
	}
}

var ErrTimeout = errors.New("timeout")

func (lp *localPeer) WaitFor(waitCtx context.Context, id string) RemotePeer {
	pw, release := lp.watchPeer(id)
	defer release()

	ch, releaseWatch := pw.ch()
	defer releaseWatch()

	select {
	case peer := <-ch:
		return peer
	case <-waitCtx.Done():
		return nil
	}
}
