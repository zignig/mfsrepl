package main

import (
	"mfs"
	"time"
)

func Process(cluster *Cluster, share *mfs.Share, interval int) {
	c := time.Tick(time.Duration(interval) * time.Second)
	shareUpdates := share.UpdateChannel()
	for {
		select {
		case <-c:
			cluster.logger.Printf("BOOP")
		case update := <-shareUpdates:
			cluster.logger.Printf("SHARE UPDATE %v", update)
		}
	}
}

//func Stuff() {
//	r := mfs.NewIPfsfs()
//	if r.Stat() {
//		val := r.Mfs("share")
//		peer.Insert(val.Hash)
//	} else {
//		logger.Printf("No ipfs node")
//	}
//	// Insert the share
//	go insert(logger, peer)
//}
