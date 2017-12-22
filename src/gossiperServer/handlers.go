package gossiperServer

import (
	"net"
	"strconv"
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
		g.MessagesLock.RLock()
		msg := g.Messages[dest]
		g.MessagesLock.RUnlock()
		g.rumorMongering(relayStr, msg)
	} else if !synced {
		g.sendStatusPacket(relay)
	}
}

func (g *Gossiper) handleRumorMessage(msg common.RumorMessage, relay *net.UDPAddr) {
	g.wantLock.RLock()
	wantMsgOrigin := g.want[msg.Origin]
	g.wantLock.RUnlock()

	if wantMsgOrigin > msg.Id && msg.Origin != g.Name && msg.LastPort == nil && msg.LastIP == nil {
		g.NextHopLock.Lock()
		g.NextHop[msg.Origin] = relay
		g.NextHopLock.Unlock()
	} else if wantMsgOrigin == msg.Id || wantMsgOrigin == 0 {
		if msg.LastPort != nil && msg.LastIP != nil {
			g.PeersLock.Lock()
			g.Peers[msg.LastIP.String()+":"+strconv.Itoa(*msg.LastPort)] = true
			g.PeersLock.Unlock()
		}

		g.wantLock.Lock()
		g.want[msg.Origin] = msg.Id + 1
		g.wantLock.Unlock()

		g.MessagesLock.Lock()
		g.Messages[common.MapKey{Origin: msg.Origin, MessageId: msg.Id}] = msg
		g.MessagesLock.Unlock()

		g.NextHopLock.Lock()
		g.NextHop[msg.Origin] = relay
		g.NextHopLock.Unlock()

		g.sendStatusPacket(relay)

		msg.LastIP = &relay.IP
		msg.LastPort = &relay.Port
		go g.iterativeRumorMongering(getRelayStr(relay), msg)
	}
}

func (g *Gossiper) handlePrivateMessage(msg common.PrivateMessage) {
	msg.HopLimit -= 1
	if msg.Destination == g.Name {
		g.sock.WriteJSON(msg.Data)
	} else if msg.HopLimit > 0 {
		g.SendPrivateMessage(msg)
	}
}
