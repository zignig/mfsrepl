package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
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
	fmt.Println("Blorp Blorp")
	r := mfs.NewIPfsfs()
	peers := &stringset{}
	var (
		meshListen = flag.String("mesh", net.JoinHostPort("0.0.0.0", strconv.Itoa(mesh.Port)), "mesh listen address")
		hwaddr     = flag.String("hwaddr", "", "MAC address, i.e. mesh peer ID")
		nickname   = flag.String("nickname", mustHostname(), "peer nickname")
		password   = flag.String("password", "", "password (optional)")
		channel    = flag.String("channel", "default", "gossip channel name")
	)
	flag.Var(peers, "peer", "initial peer (may be repeated)")
	flag.Parse()

	logger := log.New(os.Stderr, *nickname+"> ", log.LstdFlags)

	host, portStr, err := net.SplitHostPort(*meshListen)
	if err != nil {
		logger.Fatalf("mesh address: %s: %v", *meshListen, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Fatalf("mesh address: %s: %v", *meshListen, err)
	}

	name, err := mesh.PeerNameFromString(genMac())
	if err != nil {
		logger.Fatalf("%s: %v", *hwaddr, err)
	}

	router := mesh.NewRouter(mesh.Config{
		Host:               host,
		Port:               port,
		ProtocolMinVersion: mesh.ProtocolMinVersion,
		Password:           []byte(*password),
		ConnLimit:          64,
		PeerDiscovery:      true,
		TrustedSubnets:     []*net.IPNet{},
	}, name, *nickname, mesh.NullOverlay{}, log.New(ioutil.Discard, "", 0))

	peer := newPeer(name, logger)
	gossip := router.NewGossip(*channel, peer)
	peer.register(gossip)
	go info(router, logger, peer)
	func() {
		logger.Printf("mesh router starting (%s)", *meshListen)
		router.Start()
	}()
	defer func() {
		logger.Printf("mesh router stopping")
		router.Stop()
	}()

	if r.Stat() {
		val := r.Mfs("share")
		peer.Insert(val.Hash)
	} else {
		logger.Printf("No ipfs node")
	}

	router.ConnectionMaker.InitiateConnections(peers.slice(), true)

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()
	logger.Print(<-errs)
}

func info(r *mesh.Router, logger *log.Logger, peer *peer) {
	c := time.Tick(20 * time.Second)
	for {
		select {
		case <-c:
			for i, j := range r.Peers.Descriptions() {
				logger.Printf(" %v , %v ", i, j)
			}
			fmt.Println(peer.st)
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
