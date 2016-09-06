package main

import (
	"flag"
	"fmt"
	//"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mfs"
)

func main() {
	fmt.Println("MFS replicator")
	var (
		configPath = flag.String("config", "./repl.toml", "config file path")
		password   = flag.String("password", "", "password for mesh")
		peer       = flag.String("peer", "", "peer address")
	)
	flag.Parse()

	config := LoadConfig(*configPath, *peer, *password)
	config.Print()
	logger := log.New(os.Stderr, config.Nickname+"> ", log.LstdFlags)

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
	go cluster.Info(5)

	// Run and Wait
	errs := make(chan error, 1)
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
