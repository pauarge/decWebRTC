package server

import (
	"net"
	"strconv"
	"github.com/pauarge/decWebRTC/src/common"
	"log"
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
			dest = common.MapKey{Origin: i, MessageId: remoteWant[i]}
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
			dest = common.MapKey{Origin: i, MessageId: remoteWant[i]}
			break
		}
	}
	g.wantLock.RUnlock()

	if dest != (common.MapKey{}) {
		g.messagesLock.RLock()
		msg := g.messages[dest]
		g.messagesLock.RUnlock()
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
	} else if wantMsgOrigin == msg.Id || wantMsgOrigin == 0 {
		if msg.LastPort != nil && msg.LastIP != nil {
			g.peersLock.Lock()
			g.peers[msg.LastIP.String()+":"+strconv.Itoa(*msg.LastPort)] = true
			g.peersLock.Unlock()
		}

		g.wantLock.Lock()
		g.want[msg.Origin] = msg.Id + 1
		g.wantLock.Unlock()

		g.messagesLock.Lock()
		g.messages[common.MapKey{Origin: msg.Origin, MessageId: msg.Id}] = msg
		g.messagesLock.Unlock()

		g.nextHopLock.Lock()
		g.nextHop[msg.Origin] = relay
		g.nextHopLock.Unlock()

		g.sendUserList()

		g.sendStatusPacket(relay)

		msg.LastIP = &relay.IP
		msg.LastPort = &relay.Port
		g.iterativeRumorMongering(getRelayStr(relay), msg)
	}
}

func (g *Gossiper) handlePrivateMessage(msg common.PrivateMessage) {
	msg.HopLimit -= 1
	if msg.Destination == g.name {
		g.sockLock.Lock()
		g.sock.WriteJSON(msg.Data)
		g.sockLock.Unlock()
	} else if msg.HopLimit > 0 {
		log.Println("Forwaring private message")
		g.SendPrivateMessage(msg)
	}
}
