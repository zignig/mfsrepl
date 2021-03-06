package refshare

import (
	"github.com/op/go-logging"

	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/weaveworks/mesh"
	"mfs"
	"time"
)

// Peer encapsulates state and implements mesh.Gossiper.
// It should be passed to mesh.Router.NewGossip,
// and the resulting Gossip registered in turn,
// before calling mesh.Router.Start.
type Peer struct {
	st      *state
	send    mesh.Gossip
	actions chan<- func()
	quit    chan struct{}
	update  chan mfs.Update
	logger  *logging.Logger
}

// peer implements mesh.Gossiper.
var _ mesh.Gossiper = &Peer{}

// Construct a peer with empty state.
// Be sure to register a channel, later,
// so we can make outbound communication.
func NewPeer(self mesh.PeerName, logger *logging.Logger) *Peer {
	actions := make(chan func())
	p := &Peer{
		st:      newState(self),
		send:    nil, // must .register() later
		actions: actions,
		quit:    make(chan struct{}),
		update:  make(chan mfs.Update, 10),
		logger:  logger,
	}
	go p.loop(actions)
	return p
}

// getUpdate
// return the update channel
func (p *Peer) UpdateChannel() (update chan mfs.Update) {
	return p.update
}

func (p *Peer) loop(actions <-chan func()) {
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
func (p *Peer) Register(send mesh.Gossip) {
	p.actions <- func() { p.send = send }
}

// Return the current value of the counter.
func (p *Peer) get() refs {
	return p.st.get()
}

func (p *Peer) Insert(name, value string) (result refs) {
	c := make(chan struct{})
	p.actions <- func() {
		defer close(c)
		st := p.st.insert(name, value)
		//p.logger.Debugf("Insert data %v", st)
		if p.send != nil {
			p.send.GossipBroadcast(st)
		} else {
			p.logger.Critical("no sender configured; not broadcasting update right now")
		}
		result = st.get()
	}
	<-c
	return result
}

func (p *Peer) stop() {
	close(p.quit)
}

// Return a copy of our complete state.
func (p *Peer) Gossip() (complete mesh.GossipData) {
	complete = p.st.copy()
	//p.logger.Printf("Gossip => complete %v", complete.(*state).set)
	return complete
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *Peer) OnGossip(buf []byte) (delta mesh.GossipData, err error) {
	var set map[mesh.PeerName]refs
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		return nil, err
	}
	delta = p.st.mergeDelta(set)
	p.SpoolMerge(delta)
	return delta, nil
}

func (p *Peer) SpoolMerge(delta mesh.GossipData) {
	if delta != nil {
		for node, values := range delta.(*state).set {
			//fmt.Println("source ->", node)
			for key, value := range values {
				//fmt.Println("delta ", key, value)
				u := mfs.Update{
					Path:     key,
					NewHash:  value,
					Stamp:    time.Now(),
					PeerName: node.String(),
				}
				p.update <- u
			}
		}
	}
}

// Merge the gossiped data represented by buf into our state.
// Return the state information that was modified.
func (p *Peer) OnGossipBroadcast(src mesh.PeerName, buf []byte) (received mesh.GossipData, err error) {
	var set map[mesh.PeerName]refs
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		fmt.Println(err)
		return nil, err
	}
	received = p.st.mergeReceived(set)
	p.SpoolMerge(received)
	return received, nil
}

// Merge the gossiped data represented by buf into our state.
func (p *Peer) OnGossipUnicast(src mesh.PeerName, buf []byte) error {
	p.logger.Info(" unicast , %s", src)
	var set map[mesh.PeerName]refs
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&set); err != nil {
		return err
	}

	complete := p.st.mergeComplete(set)
	p.logger.Debug("OnGossipUnicast %s %v => complete %v", src, set, complete)
	return nil
}
