package server

import (
	"net"
	"log"
	"github.com/pauarge/decWebRTC/src/common"
	"strconv"
)

func (g *Gossiper) handleStatusPacket(msg common.StatusPacket, relay *net.UDPAddr) {
	relayStr := getRelayStr(relay)
	log.Println("Got status packet from", relayStr)

	g.channelsLock.RLock()
	if ch, ok := g.channels[relayStr]; ok {
		log.Println("Unlocked channel")
		ch <- true
	}
	g.channelsLock.RUnlock()

	remoteWant := parseWant(msg)
	g.wantLock.RLock()
	localWant := g.want
	synced := true
	var dest common.MapKey

	for i := range remoteWant {
		if remoteWant[i] > localWant[i] {
			synced = false
		} else if remoteWant[i] < localWant[i] {
			dest = common.MapKey{Origin: i, MessageId: localWant[i] - 1}
			break
		}
	}

	for i := range localWant {
		if remoteWant[i] > localWant[i] {
			synced = false
		} else if remoteWant[i] < localWant[i] {
			if remoteWant[i] == 0 {
				remoteWant[i] = 1
			}
			dest = common.MapKey{Origin: i, MessageId: localWant[i] - 1}
			break
		}
	}

	log.Println("Known peers:", g.getPeerList(""))
	log.Println("Local vector clock:", localWant)
	g.wantLock.RUnlock()

	if dest != (common.MapKey{}) {
		msg := common.RumorMessage{
			Origin: dest.Origin,
			Id:     dest.MessageId,
		}
		if dest.Origin != g.name {
			g.nextHopLock.RLock()
			nextHop := g.nextHop[dest.Origin]
			g.nextHopLock.RUnlock()
			msg.LastIP = &nextHop.IP
			msg.LastPort = &nextHop.Port
		}
		g.rumorMongering(relayStr, msg)
	} else if synced {
		log.Println("In sync with " + relayStr)
	} else {
		g.sendStatusPacket(relay)
	}
}

func (g *Gossiper) handleRumorMessage(msg common.RumorMessage, relay *net.UDPAddr) {
	if msg.Origin != g.name {
		g.wantLock.RLock()
		wantMsgOrigin := g.want[msg.Origin]
		g.wantLock.RUnlock()

		if wantMsgOrigin > msg.Id && msg.LastPort == nil && msg.LastIP == nil {
			g.nextHopLock.Lock()
			g.nextHop[msg.Origin] = relay
			g.nextHopLock.Unlock()
			g.sendUserList()
		} else if wantMsgOrigin <= msg.Id || wantMsgOrigin == 0 {
			if msg.LastPort != nil && msg.LastIP != nil {
				g.addPeer(msg.LastIP.String() + ":" + strconv.Itoa(*msg.LastPort))
			}

			g.nextHopLock.Lock()
			g.nextHop[msg.Origin] = relay
			g.nextHopLock.Unlock()
			g.sendUserList()

			g.wantLock.Lock()
			g.want[msg.Origin] = msg.Id + 1
			g.wantLock.Unlock()

			g.sendStatusPacket(relay)

			msg.LastIP = &relay.IP
			msg.LastPort = &relay.Port
			g.iterativeRumorMongering(getRelayStr(relay), msg)
		}
	}
}

func (g *Gossiper) handlePrivateMessage(msg common.PrivateMessage) {
	msg.HopLimit -= 1
	if msg.Destination == g.name {
		if g.sock != nil {
			log.Printf("Received a PM of type %s and forwarded to GUI\n", msg.Data.Type)
			g.sockLock.Lock()
			g.sock.WriteJSON(msg.Data)
			g.sockLock.Unlock()
		} else if msg.Data.Type != "initCallKO" {
			log.Println("Received a private message but could not forward it to GUI")
			res := common.PrivateMessage{
				Origin:      g.name,
				Destination: msg.Origin,
				HopLimit:    common.MaxHops,
				Data: common.JSONRequest{
					Type: "initCallKO",
					Name: g.name,
				},
			}
			g.sendPrivateMessage(res)
		}
	} else if msg.HopLimit > 0 {
		log.Println("Forwarding PM of type", msg.Data.Type)
		g.sendPrivateMessage(msg)
	}
}
