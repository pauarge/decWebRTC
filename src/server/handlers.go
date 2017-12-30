package server

import (
	"net"
	"log"
	"github.com/pauarge/decWebRTC/src/common"
)

func (g *Gossiper) handleStatusPacket(msg common.StatusPacket, relay *net.UDPAddr) {
	relayStr := getRelayStr(relay)

	g.channelsLock.RLock()
	if ch, ok := g.channels[relayStr]; ok {
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
	g.wantLock.RUnlock()

	log.Println("Known peers:", g.getPeerList(""))

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
	g.wantLock.RLock()
	wantMsgOrigin := g.want[msg.Origin]
	g.wantLock.RUnlock()

	if wantMsgOrigin > msg.Id && msg.Origin != g.name && msg.LastPort == nil && msg.LastIP == nil {
		g.nextHopLock.Lock()
		g.nextHop[msg.Origin] = relay
		g.nextHopLock.Unlock()
		g.sendUserList()
	} else if wantMsgOrigin <= msg.Id || wantMsgOrigin == 0 {
		g.wantLock.Lock()
		g.want[msg.Origin] = msg.Id + 1
		g.wantLock.Unlock()

		if msg.Origin != g.name {
			g.nextHopLock.Lock()
			g.nextHop[msg.Origin] = relay
			g.nextHopLock.Unlock()
			g.sendUserList()
		}

		g.sendStatusPacket(relay)

		msg.LastIP = &relay.IP
		msg.LastPort = &relay.Port
		g.iterativeRumorMongering(getRelayStr(relay), msg)
	}
}

func (g *Gossiper) handlePrivateMessage(msg common.PrivateMessage) {
	msg.HopLimit -= 1
	if msg.Destination == g.name {
		if g.sock != nil {
			log.Println("Received a private message and forwarded to GUI")
			g.sockLock.Lock()
			g.sock.WriteJSON(msg.Data)
			g.sockLock.Unlock()
		} else {
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
			g.SendPrivateMessage(res)
		}
	} else if msg.HopLimit > 0 {
		log.Println("Forwaring private message")
		g.SendPrivateMessage(msg)
	}
}
