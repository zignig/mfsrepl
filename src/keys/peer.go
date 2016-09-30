package keys

import (
	"github.com/op/go-logging"

	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/weaveworks/mesh"
)

// Peer encapsulates state and implements mesh.Gossiper.
// It should be passed to mesh.Router.NewGossip,
// and the resulting Gossip registered in turn,
// before calling mesh.Router.Start.
var logger = logging.MustGetLogger("keys")

type peer struct {
	st      *state
	send    mesh.Gossip
	actions chan<- func()
	quit    chan struct{}
	logger  *logging.Logger
}

// peer implements mesh.Gossiper.
var _ mesh.Gossiper = &peer{}

// Construct a peer with empty state.
// Be sure to register a channel, later,
// so we can make outbound communication.
func NewPeer(logger *logging.Logger) *peer {
	actions := make(chan func())
	p := &peer{
		st:      newState(),
		send:    nil, // must .register() later
		actions: actions,
		quit:    make(chan struct{}),
		//update:  make(chan ident, 10),
		logger: logger,
	}
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
func (p *peer) Gossip() (complete mesh.GossipData) {
	logger.Critical("KEY GOSSIP")
	complete = p.st.copy()
	logger.Criticalf("data -> %v\n", complete)
	//p.logger.Printf("Gossip => complete %v", complete.(*state).set)
	return complete
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *peer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	var set map[string]*SignedKey
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		return nil, err
	}
	delta = p.st.mergeDelta(set)
	p.SpoolMerge(delta)
	return delta, nil
}

func (p *peer) SpoolMerge(delta mesh.GossipData) {
	//	if delta != nil {
	//		for node, values := range delta.(*state).set {
	//		}
	//	}
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *peer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	var set map[string]*SignedKey
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		fmt.Println(err)
		return nil, err
	}
	received = p.st.mergeReceived(set)
	p.SpoolMerge(received)
	return received, nil
}

// Merge the gossiped data represented by buf into our state.
func (p *peer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	p.logger.Info(" unicast , %s", src)
	var set map[string]*SignedKey
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		return err
	}

	complete := p.st.mergeComplete(set)
	p.logger.Debug("OnGossipUnicast %s %v => complete %v", src, set, complete)
	return nil
}
