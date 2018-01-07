package server

import (
	"net"
	"log"
	"sync"
	"strconv"
	"github.com/gorilla/websocket"
	"github.com/pauarge/decWebRTC/src/stund"
)

type Gossiper struct {
	channels     map[string]chan bool
	counter      uint32
	name         string
	gossipConn   *net.UDPConn
	nextHop      map[string]*net.UDPAddr
	peers        map[string]bool
	want         map[string]uint32
	channelsLock *sync.RWMutex
	nextHopLock  *sync.RWMutex
	peersLock    *sync.RWMutex
	wantLock     *sync.RWMutex

	// GUI
	sock     *websocket.Conn
	sockLock *sync.RWMutex
}

func NewGossiper(gossipPort int, name, peers string) *Gossiper {
	gossipAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:"+strconv.Itoa(gossipPort))
	if err != nil {
		log.Fatal(err)
	}
	gossipConn, err := net.ListenUDP("udp", gossipAddr)
	if err != nil {
		log.Fatal(err)
	}
	g := &Gossiper{
		channels:     make(map[string]chan bool),
		counter:      1,
		name:         name,
		gossipConn:   gossipConn,
		nextHop:      make(map[string]*net.UDPAddr),
		peers:        make(map[string]bool),
		want:         make(map[string]uint32),
		channelsLock: &sync.RWMutex{},
		nextHopLock:  &sync.RWMutex{},
		peersLock:    &sync.RWMutex{},
		wantLock:     &sync.RWMutex{},
		sockLock:     &sync.RWMutex{},
	}
	for _, p := range parsePeerSet(peers) {
		g.addPeer(p)
	}
	return g
}

func (g *Gossiper) Start(guiPort, rtimer int, disableGui bool) {
	if !disableGui {
		go g.listenGUI(guiPort)
	}
	go stund.RunStun()

	var wg sync.WaitGroup
	wg.Add(3)
	go g.listenGossip(wg)
	go g.routeRumorGenerator(wg, rtimer)
	go g.antiEntropy(wg)
	g.iterativeRumorMongering("", g.createRouteRumor())
	wg.Wait()
}
