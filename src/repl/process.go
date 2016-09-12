package main

import (
	"mfs"
	"time"
)

// Process
func Process(cluster *Cluster, share *mfs.Share, interval int) {
	c := time.Tick(time.Duration(interval) * time.Second)
	// get the channels from the constructs
	shareUpdates := share.UpdateChannel()
	remoteUpdates := cluster.peer.UpdateChannel()
	// loop and wait for events
	for {
		select {
		case <-c:
			cluster.logger.Printf("BOOP")
			cluster.logger.Printf("%v", cluster.peer.st)
		case update := <-shareUpdates:
			cluster.logger.Printf("SHARE UPDATE %v", update)
			cluster.peer.Insert(update.Path, update.NewHash)
		case update := <-remoteUpdates:
			cluster.logger.Printf("REMOTE UPDATE %v", update)
		}
	}
}

//}
