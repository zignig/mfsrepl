package keys

import (
	"github.com/op/go-logging"

	"bytes"
	"encoding/gob"
	//	"fmt"
	"github.com/weaveworks/mesh"
)

// Peer encapsulates state and implements mesh.Gossiper.
// It should be passed to mesh.Router.NewGossip,
// and the resulting Gossip registered in turn,
// before calling mesh.Router.Start.
var logger = logging.MustGetLogger("keys")

type peer struct {
	st       *state
	send     mesh.Gossip
	actions  chan<- func()
	quit     chan struct{}
	logger   *logging.Logger
	keyStore *KeyStore
}

// peer implements mesh.Gossiper.
var _ mesh.Gossiper = &peer{}

// Construct a peer with empty state.
// Be sure to register a channel, later,
// so we can make outbound communication.
func NewPeer(keypath string, logger *logging.Logger) *peer {
	actions := make(chan func())
	p := &peer{
		st:      newState(),
		send:    nil, // must .register() later
		actions: actions,
		quit:    make(chan struct{}),
		//update:  make(chan ident, 10),
		logger: logger,
	}
	ks, err := NewKeyStore(keypath)
	if err != nil {
		logger.Fatalf("Keystore fail %v", err)
	}
	p.keyStore = ks
	p.loadAllKeys()
	go p.loop(actions)
	return p
}

// getUpdate
// return the update channel
//func (p *peer) UpdateChannel() (update chan ident) {
//	return p.update
//}

func (p *peer) loop(actions <-chan func()) {
	for {
		select {
		case f := <-actions:
			f()
		case <-p.quit:
			return
		}
	}
}

func (p *peer) loadAllKeys() {
	keys, err := p.keyStore.ListKeys("public")
	if err != nil {
		logger.Critical(err)
	}
	for _, i := range keys {
		logger.Debugf("Load Key %s", i)
		k, _ := p.keyStore.GetPublic(i, "public")
		//fmt.Println(k, e)
		p.st.insert(k)
	}
}

// register the result of a mesh.Router.NewGossip.
func (p *peer) Register(send mesh.Gossip) {
	p.actions <- func() { p.send = send }
}

func (p *peer) Insert(sigK *SignedKey) (result *SignedKey) {
	c := make(chan struct{})
	p.actions <- func() {
		defer close(c)
		st := p.st.insert(sigK)
		//p.logger.Debugf("Insert data %v", st)
		if p.send != nil {
			p.send.GossipBroadcast(st)
		} else {
			p.logger.Critical("no sender configured; not broadcasting update right now")
		}
		//		result = st.get()
	}
	<-c
	return result
}

func (p *peer) stop() {
	close(p.quit)
}

// Return a copy of our complete state.
// TODO , get a small random selection of keys
func (p *peer) Gossip() (complete mesh.GossipData) {
	logger.Critical("KEY GOSSIP")
	complete = p.st.copy()
	//logger.Criticalf("data -> %v\n", complete.(*state).set)
	return complete
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *peer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	var set map[string]*SignedKey
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		return nil, err
	}
	for i, j := range set {
		if p.keyStore.HaveKey(i, "public") == false {
			logger.Criticalf("ADDING KEY %v", i)
			err := p.keyStore.TryInsert(j, "public")
			logger.Critical(err)
		}
	}
	return delta, nil
}

func (p *peer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	return nil, nil
}

func (p *peer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	return nil
}
