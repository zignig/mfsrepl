package main

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"os"
	"os/signal"
	"syscall"

	"mfs"
	"refshare"
)

var logger *logging.Logger

func main() {
	var (
		configPath = flag.String("config", "./repl.toml", "config file path")
		password   = flag.String("password", "", "password for mesh")
		peer       = flag.String("peer", "", "peer address")
		nickname   = flag.String("nickname", "", "Nickname for the node")
		level      = flag.Int("log", 2, "Logging Level")
	)
	flag.Parse()

	config := LoadConfig(*configPath, *peer, *password, *nickname)
	LogSetup(*level, "mfsrepl")
	logging.SetLevel(logging.DEBUG, "mfs")

	logger := GetLogger("cluster")
	logger.Critical("MFS replicator")
	cluster := NewCluster(config, logger)

	// Attach the widgets
	refPeer := refshare.NewPeer(cluster.Name, logger)
	cluster.Attach(refPeer, config.Channel)
	// Spin up the mesh
	go func() {
		cluster.Start()
	}()

	// Defer the Mesh Close
	defer func() {
		cluster.Stop()
	}()

	// Show the current peers
	cluster.Peers()
	// Show a list every 10 seconds
	go cluster.Info(30)

	// Create the Shares
	shares := mfs.NewShare(config.Shares)
	// Watch the shares
	go shares.Watch(10)
	// Run the primary event loop
	go Process(cluster, refPeer, shares, 10)
	// Run and Wait
	errs := make(chan error, 1)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()
	logger.Critical(<-errs)
}
