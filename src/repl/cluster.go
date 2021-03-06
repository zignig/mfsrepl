package main

import (
	"crypto/rand"
	"fmt"
	"github.com/op/go-logging"
	"github.com/weaveworks/mesh"
	"net"
	"os"

	"strconv"
	"time"
)

type Cluster struct {
	Name mesh.PeerName

	config *Config
	logger *logging.Logger
	router *mesh.Router
	names  map[string]string
}

func NewCluster(config *Config, logger *logging.Logger) (cl *Cluster) {
	cl = &Cluster{config: config, logger: logger}
	host, portStr, err := net.SplitHostPort(config.Listen)
	if err != nil {
		logger.Fatalf("mesh address: %s: %v", config.Listen, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Fatalf("mesh address: %s: %v", config.Listen, err)
	}
	name, err := mesh.PeerNameFromString(config.PeerID)
	if err != nil {
		logger.Fatalf("%v", err)
	}
	cl.Name = name
	// log wrapper for mesh internal logger
	wrap := NewLogWrap(logger)
	router, err := mesh.NewRouter(mesh.Config{
		Host:               host,
		Port:               port,
		ProtocolMinVersion: mesh.ProtocolMinVersion,
		Password:           []byte(config.Password),
		ConnLimit:          64,
		PeerDiscovery:      config.Discovery,
		TrustedSubnets:     []*net.IPNet{},
	}, name, config.Nickname, mesh.NullOverlay{}, wrap) //log.New(ioutil.Discard, "", 0))
	cl.router = router
	if err != nil {
		panic("mesh fail")
	}

	//peer := refshare.NewPeer(name, logger)
	//gossip := router.NewGossip(config.Channel, peer)
	//peer.Register(gossip)
	// bind the peer
	//cl.peer = peer
	cl.names = make(map[string]string)
	return cl
}

// Crazy interface hack
type Widget interface {
	mesh.Gossiper
	Register(send mesh.Gossip)
}

func (cl *Cluster) Attach(widget Widget, channel string) {
	gossip, _ := cl.router.NewGossip(channel, widget)
	widget.Register(gossip)
}

func (cl *Cluster) GetNames() {
	stat := mesh.NewStatus(cl.router)
	for _, j := range stat.Peers {
		cl.names[j.Name] = j.NickName
	}

}

func (cl *Cluster) Start() {
	if len(cl.config.Peers) > 0 {
		cl.router.ConnectionMaker.InitiateConnections(cl.config.Peers, true)
	}
	cl.logger.Infof("mesh router starting (%s)", cl.config.Listen)
	cl.router.Start()
}

func (cl *Cluster) Stop() {
	cl.logger.Critical("mesh router stopping")
	cl.router.Stop()
}

func (cl *Cluster) Peers() {
	cl.GetNames()
	for i, j := range cl.router.Peers.Descriptions() {
		cl.logger.Infof(" %v , %v [%v] ", i, j.NickName, j.Name)
		//cl.logger.Infof("NAMES %v", cl.names)
	}
}

func (cl *Cluster) Info(interval int) {
	c := time.Tick(time.Duration(interval) * time.Second)
	for {
		select {
		case <-c:
			cl.Peers()
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
