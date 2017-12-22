package gossiperServer

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
	Name         string
	gossipConn   *net.UDPConn
	channels     map[string]chan bool
	Messages     map[common.MapKey]common.RumorMessage
	NextHop      map[string]*net.UDPAddr
	Peers        map[string]bool
	want         map[string]uint32
	channelsLock *sync.RWMutex
	MessagesLock *sync.RWMutex
	NextHopLock  *sync.RWMutex
	PeersLock    *sync.RWMutex
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
		Name:         name,
		gossipConn:   gossipConn,
		channels:     make(map[string]chan bool),
		Messages:     make(map[common.MapKey]common.RumorMessage),
		NextHop:      make(map[string]*net.UDPAddr),
		Peers:        parsePeerSet(peers, gossipAddrRaw),
		want:         make(map[string]uint32),
		channelsLock: &sync.RWMutex{},
		MessagesLock: &sync.RWMutex{},
		NextHopLock:  &sync.RWMutex{},
		PeersLock:    &sync.RWMutex{},
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
