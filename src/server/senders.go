package server

import (
	"net"
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
		log.Println(err)
	} else {
		_, err = g.gossipConn.WriteToUDP(packetBytes, relay)
		if err != nil {
			log.Println(err)
			g.peersLock.Lock()
			delete(g.peers, getRelayStr(relay))
			g.peersLock.Unlock()
			g.sendUserList()
		}
	}
}

func (g *Gossiper) SendPrivateMessage(msg common.PrivateMessage) {
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
				g.peersLock.Lock()
				delete(g.peers, getRelayStr(relay))
				g.peersLock.Unlock()
				g.sendUserList()
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
		addr, err := net.ResolveUDPAddr("udp4", address)
		_, err = g.gossipConn.WriteToUDP(packetBytes, addr)
		if err != nil {
			log.Println(err)
			g.peersLock.Lock()
			delete(g.peers, address)
			g.peersLock.Unlock()
			g.sendUserList()
		} else {
			g.channelsLock.Lock()
			g.channels[address] = make(chan bool)
			g.channelsLock.Unlock()
			ticker := time.NewTicker(time.Second * common.TimeOutSecs)
			go func() {
				for range ticker.C {
					g.channelsLock.Lock()
					if ch, ok := g.channels[address]; ok {
						ch <- true
						log.Println("Timeout on mongering")
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
	}
}

func (g *Gossiper) iterativeRumorMongering(exclude string, msg common.RumorMessage) {
	peers := g.getPeerList(exclude)
	for i := range peers {
		g.rumorMongering(peers[i], msg)
	}
}
