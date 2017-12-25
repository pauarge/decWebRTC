package server

import (
	"net"

	"log"
	"sync"
	"github.com/pauarge/decWebRTC/src/common"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Gossiper struct {
	counter      uint32
	name         string
	gossipConn   *net.UDPConn
	channels     map[string]chan bool
	messages     map[common.MapKey]common.RumorMessage
	nextHop      map[string]*net.UDPAddr
	peers        map[string]bool
	want         map[string]uint32
	channelsLock *sync.RWMutex
	messagesLock *sync.RWMutex
	nextHopLock  *sync.RWMutex
	peersLock    *sync.RWMutex
	wantLock     *sync.RWMutex

	// GUI
	router   *mux.Router
	sock     *websocket.Conn
	sockLock *sync.RWMutex
}

func NewGossiper(gossipAddrRaw, name, peers string) *Gossiper {
	gossipAddr, err := net.ResolveUDPAddr("udp4", gossipAddrRaw)
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
		messages:     make(map[common.MapKey]common.RumorMessage),
		nextHop:      make(map[string]*net.UDPAddr),
		peers:        parsePeerSet(peers, gossipAddrRaw),
		want:         make(map[string]uint32),
		channelsLock: &sync.RWMutex{},
		messagesLock: &sync.RWMutex{},
		nextHopLock:  &sync.RWMutex{},
		peersLock:    &sync.RWMutex{},
		wantLock:     &sync.RWMutex{},
		router:       mux.NewRouter(),
		sockLock:     &sync.RWMutex{},
	}
}

func (g *Gossiper) Start(rtimer int) {
	var wg sync.WaitGroup
	wg.Add(4)
	go g.listenGUI()
	go g.listenGossip(wg)
	go g.routeRumorGenerator(wg, rtimer)
	go g.antiEntropy(wg)
	g.iterativeRumorMongering("", g.createRouteRumor())
	wg.Wait()
}
