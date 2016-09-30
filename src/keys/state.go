package keys

import (
	"bytes"
	"sync"

	"encoding/gob"

	"github.com/op/go-logging"
	"github.com/weaveworks/mesh"
)

var log = logging.MustGetLogger("keyset")

// state is an implementation of a G-counter.
type state struct {
	mtx sync.RWMutex
	set map[string]*SignedKey
}

// state implements GossipData.
var _ mesh.GossipData = &state{}

// Construct an empty state object, ready to receive updates.
// This is suitable to use at program start.
// Other peers will populate us with data.
func newState() *state {
	return &state{
		set: make(map[string]*SignedKey),
	}
}

func (st *state) String() string {
	s := "\n"
	for i, _ := range st.set {
		s += i + "\n"
	}
	return s
}

func (st *state) get(fp string) (result *SignedKey) {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	return st.set[fp]
}

func (st *state) insert(sigK *SignedKey) (state *state) {
	return
}

func (st *state) copy() *state {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	return &state{
		set: st.set,
	}
}

// Encode serializes our complete state to a slice of byte-slices.
// In this simple example, we use a single JSON-encoded buffer.
func (st *state) Encode() [][]byte {
	st.mtx.RLock()
	defer st.mtx.RUnlock()
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(st.set); err != nil {
		panic(err)
	}
	return [][]byte{buf.Bytes()}
}

// Merge merges the other GossipData into this one,
// and returns our resulting, complete state.
func (st *state) Merge(other mesh.GossipData) (complete mesh.GossipData) {
	return st.mergeComplete(other.(*state).copy().set)
}

// Merge the set into our state, abiding increment-only semantics.
// Return a non-nil mesh.GossipData representation of the received set.
func (st *state) mergeReceived(set map[string]*SignedKey) (received mesh.GossipData) {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	for peer, v := range set {
		st.set[peer] = v
	}

	return &state{
		set: set, // all remaining elements were novel to us
	}
}

// Return any key/values that have been mutated, or nil if nothing changed.
// TODO this needs to check sub keys
func (st *state) mergeDelta(set map[string]*SignedKey) (delta mesh.GossipData) {
	st.mtx.Lock()
	defer st.mtx.Unlock()
	for peer, v := range set {
		// Do we have the key in our data
		if _, ok := st.set[peer]; ok {
			delete(set, peer) // requirement: it's not part of a delta
			continue
		}
		st.set[peer] = v
	}

	//log.Debugf("%v -> %v", set, delta)
	if len(set) <= 0 {
		return nil // per OnGossip requirements
	}
	return &state{
		set: set, // all remaining elements were novel to us
	}
}

// Merge the set into our state, abiding increment-only semantics.
// Return our resulting, complete state.
func (st *state) mergeComplete(set map[string]*SignedKey) (complete mesh.GossipData) {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	for peer, v := range set {
		st.set[peer] = v
	}

	return &state{
		set: st.set, // n.b. can't .copy() due to lock contention
	}
}
