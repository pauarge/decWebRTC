package gossiperServer

import (
	"time"
	"math/rand"
	"strings"
	"net"
	"strconv"
	"sort"
	"log"
	"fmt"
	"github.com/pauarge/decWebRTC/src/common"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}

func parsePeerSet(peers, gossipAddr string) map[string]bool {
	peerSet := make(map[string]bool)
	if last := len(peers) - 1; last >= 0 {
		if peers[last] == ',' {
			peers = peers[:last]
		}
		for _, i := range strings.Split(peers, ",") {
			if i != gossipAddr {
				_, err := net.ResolveUDPAddr("udp4", i)
				if err != nil {
					log.Fatal(err)
				}
				peerSet[i] = true
			}
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

func (g *Gossiper) encodeWant() common.StatusPacket {
	defer g.wantLock.RUnlock()
	g.wantLock.RLock()

	var x []common.PeerStatus
	for k := range g.want {
		x = append(x, common.PeerStatus{Identifier: k, NextId: g.want[k]})
	}
	sort.Sort(common.PeerStatusList(x))
	return common.StatusPacket{Want: x}
}

func (g *Gossiper) getPeerList(exclude string) []string {
	defer g.PeersLock.RUnlock()
	defer g.NextHopLock.RUnlock()
	g.PeersLock.RLock()
	g.NextHopLock.RLock()

	fmt.Print("PEERS ")
	var p []string
	for i := range g.Peers {
		//fmt.Print(i + ",")
		if i != exclude {
			p = append(p, i)
		}
	}
	fmt.Println()

	fmt.Print("DSDV ")
	for k, v := range g.NextHop {
		fmt.Print(k + ": " + v.String() + " ")
	}
	fmt.Println()
	return p
}

func (g *Gossiper) createRouteRumor() common.RumorMessage {
	defer g.wantLock.Unlock()
	defer g.MessagesLock.Unlock()
	g.MessagesLock.Lock()
	g.wantLock.Lock()

	msg := common.RumorMessage{Origin: g.Name, Id: g.counter, Text: ""}
	g.Messages[common.MapKey{Origin: g.Name, MessageId: g.counter}] = msg
	g.counter += 1
	g.want[g.Name] = g.counter
	return msg
}
