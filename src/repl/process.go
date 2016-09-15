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
	cluster.logger.Info("Starting Main Loop")
	for {
		select {
		case <-c:
			cluster.logger.Debugf("%v", cluster.peer.st)
		case update := <-shareUpdates:
			cluster.logger.Infof("SHARE UPDATE %v", update)
			cluster.peer.Insert(update.Path, update.NewHash)
		case update := <-remoteUpdates:
			cluster.logger.Infof("REMOTE UPDATE %v", update)
			//share.Backup("")
		}
	}
}
