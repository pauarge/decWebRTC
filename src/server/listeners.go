package server

import (
	"sync"
	"log"
	"time"
	"net"
	"strconv"
	"net/http"
	"github.com/pauarge/decWebRTC/src/common"
	"github.com/dedis/protobuf"
)

func (g *Gossiper) listenGUI(guiPort int) {
	http.HandleFunc("/echo", g.echoHandler)
	http.Handle("/", http.FileServer(http.Dir("src/static/")))

	log.Printf("Serving on HTTP port %d\n", guiPort)
	log.Fatal(http.ListenAndServe(":"+ strconv.Itoa(guiPort), nil))
}

func (g *Gossiper) listenGossip(wg sync.WaitGroup) {
	defer wg.Done()
	for {
		var buf = make([]byte, common.BufferSize)
		n, relay, err := g.gossipConn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		g.peersLock.Lock()
		g.peers[getRelayStr(relay)] = true
		g.peersLock.Unlock()
		m := common.GossipPacket{}
		err = protobuf.Decode(buf[0:n], &m)
		if err == nil {
			if m.Rumor != nil {
				go g.handleRumorMessage(*m.Rumor, relay)
			} else if m.Private != nil {
				go g.handlePrivateMessage(*m.Private)
			} else if m.Status != nil {
				go g.handleStatusPacket(*m.Status, relay)
			}
		} else {
			log.Println(err)
		}
	}
}

func (g *Gossiper) antiEntropy(wg sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * common.LongTimeOutSecs)
	for range ticker.C {
		var p []string
		g.peersLock.RLock()
		for i := range g.peers {
			p = append(p, i)
		}
		g.peersLock.RUnlock()
		if len(p) > 0 {
			e, _ := pickRandomElem(p)
			addr, err := net.ResolveUDPAddr("udp4", e)
			if err != nil {
				log.Fatal(err)
			}
			g.sendStatusPacket(addr)
		}
	}
}

func (g *Gossiper) routeRumorGenerator(wg sync.WaitGroup, rtimer int) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(rtimer) * time.Second)
	for range ticker.C {
		msg := g.createRouteRumor()
		g.iterativeRumorMongering("", msg)
	}
}
