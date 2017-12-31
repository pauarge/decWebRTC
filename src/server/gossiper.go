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
	counter      uint32
	name         string
	gossipConn   *net.UDPConn
	channels     map[string]chan bool
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
	return &Gossiper{
		counter:      1,
		name:         name,
		gossipConn:   gossipConn,
		channels:     make(map[string]chan bool),
		nextHop:      make(map[string]*net.UDPAddr),
		peers:        parsePeerSet(peers),
		want:         make(map[string]uint32),
		channelsLock: &sync.RWMutex{},
		nextHopLock:  &sync.RWMutex{},
		peersLock:    &sync.RWMutex{},
		wantLock:     &sync.RWMutex{},
		sockLock:     &sync.RWMutex{},
	}
}

func (g *Gossiper) Start(guiPort, rtimer int) {
	var wg sync.WaitGroup
	wg.Add(3)
	go g.listenGUI(guiPort)
	go stund.RunStun()
	go g.listenGossip(wg)
	go g.routeRumorGenerator(wg, rtimer)
	go g.antiEntropy(wg)
	g.iterativeRumorMongering("", g.createRouteRumor())
	wg.Wait()
}
