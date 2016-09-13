package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/op/go-logging"
	"mfs"
)

var logger *logging.Logger

func main() {
	fmt.Println("MFS replicator")
	var (
		configPath = flag.String("config", "./repl.toml", "config file path")
		password   = flag.String("password", "", "password for mesh")
		peer       = flag.String("peer", "", "peer address")
		nickname   = flag.String("nickname", "", "Nickname for the node")
		level      = flag.Int("log", 0, "Logging Level")
	)
	flag.Parse()

	config := LoadConfig(*configPath, *peer, *password, *nickname)
	LogSetup(*level, "mfsrepl")

	logger := GetLogger("cluster")
	logger.Critical("NAARG")
	cluster := NewCluster(config, logger)

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
	go cluster.Info(10)

	// Create the Shares
	shares := mfs.NewShare(config.Shares, logger)
	// Watch the shares
	go shares.Watch(10)
	// Run the primary event loop
	go Process(cluster, shares, 5)
	// Run and Wait
	errs := make(chan error, 1)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()
	logger.Critical(<-errs)
}
