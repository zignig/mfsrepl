package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	//"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/mesh"

	"mfs"
)

type stringset map[string]struct{}

func (ss stringset) Set(value string) error {
	ss[value] = struct{}{}
	return nil
}

func (ss stringset) String() string {
	return strings.Join(ss.slice(), ",")
}

func (ss stringset) slice() []string {
	slice := make([]string, 0, len(ss))
	for k := range ss {
		slice = append(slice, k)
	}
	sort.Strings(slice)
	return slice
}

func main() {
	fmt.Println("MFS replicator")
	peers := &stringset{}
	var (
		meshListen = flag.String("mesh", net.JoinHostPort("0.0.0.0", strconv.Itoa(mesh.Port)), "mesh listen address")
		nickname   = flag.String("nickname", mustHostname(), "peer nickname")
		channel    = flag.String("channel", "default", "gossip channel name")
		config     = flag.String("config", "./config.toml", "config file path")
	)
	flag.Var(peers, "peer", "initial peer (may be repeated)")
	flag.Parse()
	c := LoadConfig(*config)
	fmt.Println(c)
	logger := log.New(os.Stderr, *nickname+"> ", log.LstdFlags)

	host, portStr, err := net.SplitHostPort(*meshListen)
	if err != nil {
		logger.Fatalf("mesh address: %s: %v", *meshListen, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Fatalf("mesh address: %s: %v", *meshListen, err)
	}

	name, err := mesh.PeerNameFromString(c.PeerID)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	router := mesh.NewRouter(mesh.Config{
		Host:               host,
		Port:               port,
		ProtocolMinVersion: mesh.ProtocolMinVersion,
		Password:           []byte(c.Password),
		ConnLimit:          64,
		PeerDiscovery:      false,
		TrustedSubnets:     []*net.IPNet{},
	}, name, *nickname, mesh.NullOverlay{}, logger) //log.New(ioutil.Discard, "", 0))

	peer := newPeer(name, logger)
	gossip := router.NewGossip(*channel, peer)
	peer.register(gossip)
	r := mfs.NewIPfsfs()
	if r.Stat() {
		val := r.Mfs("share")
		peer.Insert(val.Hash)
	} else {
		logger.Printf("No ipfs node")
	}
	go info(router, logger, peer)
	go insert(logger, peer)
	func() {
		logger.Printf("mesh router starting (%s)", *meshListen)
		router.Start()
	}()
	defer func() {
		logger.Printf("mesh router stopping")
		router.Stop()
	}()

	//router.ConnectionMaker.InitiateConnections(peers.slice(), true)
	if len(c.Peers) > 0 {
		router.ConnectionMaker.InitiateConnections(c.Peers, true)
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()
	logger.Print(<-errs)
}

func insert(logger *log.Logger, peer *peer) {
	c := time.Tick(1 * time.Minute)
	for {
		select {
		case <-c:
			r := mfs.NewIPfsfs()
			if r.Stat() {
				val := r.Mfs("share")
				peer.Insert(val.Hash)
			} else {
				logger.Printf("No ipfs node")
			}
		}
	}
}
func info(r *mesh.Router, logger *log.Logger, peer *peer) {
	c := time.Tick(20 * time.Second)
	for {
		select {
		case <-c:
			for i, j := range r.Peers.Descriptions() {
				logger.Printf(" %v , %v [%v] -> %v ", i, j.NickName, j.Name, peer.st.set[j.Name])
			}
		}
	}
}

func mustHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return hostname
}

func genMac() (s string) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}
	buf[0] |= 2
	s = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
	return s
}
