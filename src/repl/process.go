package main

import (
	"mfs"
	//	"time"
)

// Process
func Process(cluster *Cluster, share *mfs.Share, interval int) {
	//c := time.Tick(time.Duration(interval) * time.Second)
	// get the channels from the constructs
	shareUpdates := share.UpdateChannel()
	remoteUpdates := cluster.peer.UpdateChannel()
	// loop and wait for events
	cluster.logger.Info("Starting Main Loop")
	for {
		select {
		//case <-c:
		//	cluster.logger.Debugf("%v", cluster.peer.st)
		case shareupdate := <-shareUpdates:
			cluster.logger.Critical("SHARE UPDATE %v", shareupdate)
			cluster.peer.Insert(shareupdate.Path, shareupdate.NewHash)
		case update := <-remoteUpdates:
			// update the names
			cluster.GetNames()
			// convert to name from peerid
			// only write active peers
			val, ok := cluster.names[update.PeerName]
			if ok {
				update.PeerName = val
				cluster.logger.Critical("INCOMING UPDATE %v", update)
				share.SubmitUpdate(update)
			}
			//share.Mkdir("/"+update.Path+"/"+update.PeerName, true)
			cluster.logger.Critical("UPDATE FINISHED")
		}
	}
}
