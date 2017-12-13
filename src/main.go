package main

import (
	"flag"
	"github.com/pauarge/peerster/gossiper/gossiperServer"
	"github.com/pauarge/peerster/gossiper/guiServer"
	"github.com/pauarge/peerster/gossiper/common"
)

func main() {
	gossipAddrPtr := flag.String("gossipAddr", "", "Address and port in which gossips are listened")
	guiPtr := flag.Bool("gui", false, "Activate GUI")
	namePtr := flag.String("name", "", "Name of the node")
	noForwardPtr := flag.Bool("noforward", false, "Disable forwarding of rumor or point to point messages")
	peersPtr := flag.String("peers", "", "List of peers")
	rtimerPtr := flag.Int("rtimer", common.DefaultRTimer, "How many seconds the peer waits between two "+
		"route rumor messagese conds for rtimer")
	uIPortPtr := flag.Int("UIPort", -1, "Port in which the UI has to run")
	flag.Parse()

	if *uIPortPtr == -1 || *gossipAddrPtr == "" || *namePtr == "" {
		flag.PrintDefaults()
	} else {
		g := gossiperServer.NewGossiper(*uIPortPtr, *gossipAddrPtr, *namePtr, *peersPtr, *noForwardPtr)
		if *guiPtr {
			go guiServer.NewServer(g).Start()
		}
		g.Start(*rtimerPtr)
	}
}
