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
			//cluster.logger.Debugf("%v", cluster.peer.st)
		case update := <-shareUpdates:
			cluster.logger.Infof("SHARE UPDATE %v", update)
			cluster.peer.Insert(update.Path, update.NewHash)
		case update := <-remoteUpdates:
			// update the names
			cluster.GetNames()
			// convert to name from peerid
			val, ok := cluster.names[update.PeerName]
			if ok {
				update.PeerName = val
				cluster.logger.Infof("%v", update)
				share.SubmitUpdate(update)
			}
			//share.Mkdir("/"+update.Path+"/"+update.PeerName, true)
		}
	}
}
