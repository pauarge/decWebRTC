package main

import (
	"flag"
	"github.com/pauarge/decWebRTC/src/gossiperServer"
	"github.com/pauarge/decWebRTC/src/guiServer"
	"github.com/pauarge/decWebRTC/src/common"
)

func main() {
	gossipAddrPtr := flag.String("gossipAddr", "", "Address and port in which gossips are listened")
	namePtr := flag.String("name", "", "Name of the node")
	peersPtr := flag.String("peers", "", "List of peers")
	rtimerPtr := flag.Int("rtimer", common.DefaultRTimer, "How many seconds the peer waits between two "+
		"route rumor messagese conds for rtimer")
	flag.Parse()

	if *gossipAddrPtr == "" || *namePtr == "" {
		flag.PrintDefaults()
	} else {
		g := gossiperServer.NewGossiper(*gossipAddrPtr, *namePtr, *peersPtr)
		go guiServer.NewServer(g).Start()
		g.Start(*rtimerPtr)
	}
}
