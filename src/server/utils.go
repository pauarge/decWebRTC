package server

import (
	"time"
	"math/rand"
	"strings"
	"net"
	"strconv"
	"sort"
	"log"
	"github.com/pauarge/decWebRTC/src/common"
)

func parsePeerSet(peers string) map[string]bool {
	peerSet := make(map[string]bool)
	if last := len(peers) - 1; last >= 0 {
		if peers[last] == ',' {
			peers = peers[:last]
		}
		for _, i := range strings.Split(peers, ",") {
			_, err := net.ResolveUDPAddr("udp4", i)
			if err != nil {
				log.Fatal(err)
			}
			peerSet[i] = true
		}
	}
	return peerSet
}

func pickRandomElem(elems []string) (string, []string) {
	if len(elems) < 1 {
		log.Fatal("ELEMS HAS NO ELEMENTS")
	}
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)
	i := r.Intn(len(elems))
	e := elems[i]
	return e, append(elems[:i], elems[i+1:]...)
}

func getRelayStr(relay *net.UDPAddr) string {
	return relay.IP.String() + ":" + strconv.Itoa(relay.Port)
}

func parseWant(msg common.StatusPacket) map[string]uint32 {
	res := make(map[string]uint32)
	for _, v := range msg.Want {
		res[v.Identifier] = v.NextId
	}
	return res
}

func (g *Gossiper) addPeer(addr string) {
	g.peersLock.RLock()
	_, ok := g.peers[addr]
	g.peersLock.RUnlock()

	// Only checking peer connectivity if we don't have it
	if !ok {
		timeout := time.Duration(common.TimeOutSecs * time.Second)
		_, err := net.DialTimeout("udp", addr, timeout)
		if err != nil {
			log.Println("Peer unreachable:", err)
		} else {
			log.Println("Added peer", addr)
			g.peersLock.Lock()
			g.peers[addr] = true
			g.peersLock.Unlock()
			g.sendUserList()
		}
	}
}

func (g *Gossiper) deletePeer(addr string) {
	defer g.peersLock.Unlock()
	defer g.nextHopLock.Unlock()
	defer g.wantLock.Unlock()
	g.peersLock.Lock()
	g.nextHopLock.Lock()
	g.wantLock.Lock()

	delete(g.peers, addr)
	delete(g.nextHop, addr)
	delete(g.want, addr)
	g.sendUserList()
}

func (g *Gossiper) encodeWant() common.StatusPacket {
	defer g.wantLock.RUnlock()
	g.wantLock.RLock()

	var x []common.PeerStatus
	for k := range g.want {
		x = append(x, common.PeerStatus{Identifier: k, NextId: g.want[k]})
	}

	return common.StatusPacket{Want: x}
}

func (g *Gossiper) getPeerList(exclude string) []string {
	defer g.peersLock.RUnlock()
	g.peersLock.RLock()

	var peers []string
	for i := range g.peers {
		if i != exclude {
			peers = append(peers, i)
		}
	}

	sort.Strings(peers)
	return peers
}

func (g *Gossiper) sendUserList() {
	if g.sock != nil {
		defer g.nextHopLock.RUnlock()
		g.nextHopLock.RLock()

		var users []string
		for k := range g.nextHop {
			users = append(users, k)
		}

		sort.Strings(users)
		g.sockLock.Lock()
		g.sock.WriteJSON(common.JSONRequest{Type: "users", Users: users, Peers: g.getPeerList("")})
		g.sockLock.Unlock()
	}
}

func (g *Gossiper) createRouteRumor() common.RumorMessage {
	defer g.wantLock.Unlock()
	g.wantLock.Lock()

	msg := common.RumorMessage{Origin: g.name, Id: g.counter}
	g.counter += 1
	g.want[g.name] = g.counter
	return msg
}
