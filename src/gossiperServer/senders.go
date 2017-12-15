package gossiperServer

import (
	"net"
	"fmt"
	"log"
	"time"
	"github.com/dedis/protobuf"
	"github.com/pauarge/decWebRTC/src/common"
)

func (g *Gossiper) sendStatusPacket(relay *net.UDPAddr) {
	msg := g.encodeWant()
	p := common.GossipPacket{Status: &msg}
	packetBytes, err := protobuf.Encode(&p)
	if err != nil {
		log.Fatal(err)
	}
	_, err = g.gossipConn.WriteToUDP(packetBytes, relay)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Gossiper) sendPrivateMessage(dest string, p common.GossipPacket) {
	packetBytes, err := protobuf.Encode(&p)
	if err != nil {
		log.Fatal(err)
	}
	g.NextHopLock.RLock()
	relay, ok := g.NextHop[dest]
	g.NextHopLock.RUnlock()
	if ok {
		g.gossipConn.WriteToUDP(packetBytes, relay)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (g *Gossiper) iterativeRumorMongering(exclude string, msg common.RumorMessage) {
	peers := g.getPeerList(exclude)
	for i := range peers {
		g.rumorMongering(peers[i], msg)
	}
}

func (g *Gossiper) rumorMongering(address string, msg common.RumorMessage) {
	fmt.Println("MONGERING ROUTE to " + address)
	p := common.GossipPacket{Rumor: &msg}
	packetBytes, err := protobuf.Encode(&p)
	if err != nil {
		log.Fatal(err)
	}
	addr, err := net.ResolveUDPAddr("udp4", address)
	_, err = g.gossipConn.WriteToUDP(packetBytes, addr)
	if err != nil {
		log.Fatal(msg, err)
	}
	g.channelsLock.Lock()
	g.channels[address] = make(chan bool)
	g.channelsLock.Unlock()
	ticker := time.NewTicker(time.Second * common.TimeOutSecs)
	go func() {
		for range ticker.C {
			g.channelsLock.Lock()
			if ch, ok := g.channels[address]; ok {
				ch <- true
				fmt.Println("TIMEOUT ON MONGERING")
			}
			g.channelsLock.Unlock()
			return
		}
	}()
	g.channelsLock.RLock()
	ch := g.channels[address]
	g.channelsLock.RUnlock()
	_ = <-ch
	ticker.Stop()
	g.channelsLock.Lock()
	delete(g.channels, address)
	g.channelsLock.Unlock()
}