package server

import (
	"net"
	"log"
	"github.com/dedis/protobuf"
	"github.com/pauarge/decWebRTC/src/common"
	"time"
)

func (g *Gossiper) sendStatusPacket(relay *net.UDPAddr) {
	msg := g.encodeWant()
	p := common.GossipPacket{Status: &msg}
	packetBytes, err := protobuf.Encode(&p)
	if err != nil {
		log.Println(err)
	} else {
		_, err = g.gossipConn.WriteToUDP(packetBytes, relay)
		if err != nil {
			log.Println(err)
			g.deletePeer(getRelayStr(relay))
		}
	}
}

func (g *Gossiper) sendPrivateMessage(msg common.PrivateMessage) {
	p := common.GossipPacket{Private: &msg}
	packetBytes, err := protobuf.Encode(&p)
	if err != nil {
		log.Println(err)
	} else {
		g.nextHopLock.RLock()
		relay, ok := g.nextHop[msg.Destination]
		g.nextHopLock.RUnlock()
		if ok {
			_, err = g.gossipConn.WriteToUDP(packetBytes, relay)
			if err != nil {
				log.Println(err)
				g.deletePeer(getRelayStr(relay))
			}
		} else {
			log.Println("Could not find a next hop for sending the private message.")
		}
	}
}

func (g *Gossiper) rumorMongering(address string, msg common.RumorMessage) {
	log.Println("Mongering route to " + address)
	p := common.GossipPacket{Rumor: &msg}
	packetBytes, err := protobuf.Encode(&p)
	if err != nil {
		log.Println(err)
	} else {
		g.channelsLock.RLock()
		ch := g.channels[address]
		g.channelsLock.RUnlock()

		addr, err := net.ResolveUDPAddr("udp4", address)
		_, err = g.gossipConn.WriteToUDP(packetBytes, addr)
		if err != nil {
			log.Println(err)
			g.deletePeer(address)
			ch <- true
		}
		select {
		case _ = <-ch:
			log.Println("Unlocked channel")
		case <-time.After(time.Second * common.TimeOutSecs):
			log.Println("Timeout on mongering")
			g.deletePeer(address)
		}
	}
	log.Println("Finished mongering to", address)
}

func (g *Gossiper) iterativeRumorMongering(exclude string, msg common.RumorMessage) {
	peers := g.getPeerList(exclude)
	for i := range peers {
		go g.rumorMongering(peers[i], msg)
	}
}
