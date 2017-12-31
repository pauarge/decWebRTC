package main

import (
	"flag"
	"github.com/pauarge/decWebRTC/src/server"
	"github.com/pauarge/decWebRTC/src/common"
)

func main() {
	gossipPortPtr := flag.Int("gossipPort", common.GossipPort, "Port in which gossips are listened")
	guiPortPtr := flag.Int("guiPort", common.GuiPort, "Port in which the GUI is offered")
	namePtr := flag.String("name", "", "Name of the node")
	peersPtr := flag.String("peers", "", "List of peers")
	rtimerPtr := flag.Int("rtimer", common.DefaultRTimer, "How many seconds the peer waits between two "+
		"route rumor messagese conds for rtimer")
	flag.Parse()

	if *namePtr == "" {
		flag.PrintDefaults()
	} else {
		g := server.NewGossiper(*gossipPortPtr, *namePtr, *peersPtr)
		g.Start(*guiPortPtr, *rtimerPtr)
	}
}
