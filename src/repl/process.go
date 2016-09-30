package main

import (
	"mfs"
	"refshare"
	//	"time"
)

// Process
func Process(cluster *Cluster, peer *refshare.Peer, share *mfs.Share, interval int) {
	//c := time.Tick(time.Duration(interval) * time.Second)
	// get the channels from the constructs
	shareUpdates := share.UpdateChannel()
	remoteUpdates := peer.UpdateChannel()
	// loop and wait for events
	cluster.logger.Info("Starting Main Loop")
	for {
		select {
		//case <-c:
		//	cluster.logger.Debugf("%v", cluster.peer.st)
		case shareupdate := <-shareUpdates:
			cluster.logger.Debug("SHARE UPDATE %v", shareupdate)
			peer.Insert(shareupdate.Path, shareupdate.NewHash)
		case update := <-remoteUpdates:
			// update the names
			cluster.GetNames()
			// convert to name from peerid
			// only write active peers
			val, ok := cluster.names[update.PeerName]
			if ok {
				update.PeerName = val
				cluster.logger.Debug("INCOMING UPDATE %v", update)
				share.SubmitUpdate(update)
			}
			//share.Mkdir("/"+update.Path+"/"+update.PeerName, true)
			cluster.logger.Debug("UPDATE FINISHED")
		}
	}
}
