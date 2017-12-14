package gossiperServer

import (
	"net"

	"log"
	"sync"
	"github.com/pauarge/decWebRTC/src/common"
)

type Gossiper struct {
	counter             uint32
	Name                string
	gossipConn          *net.UDPConn
	channels            map[string]chan bool
	Messages            map[common.MapKey]common.RumorMessage
	NextHop             map[string]*net.UDPAddr
	Peers               map[string]bool
	PrivateMessages     map[string][]common.PrivateMessage
	want                map[string]uint32
	channelsLock        *sync.RWMutex
	MessagesLock        *sync.RWMutex
	NextHopLock         *sync.RWMutex
	PeersLock           *sync.RWMutex
	PrivateMessagesLock *sync.RWMutex
	wantLock            *sync.RWMutex
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
		counter:             1,
		Name:                name,
		gossipConn:          gossipConn,
		channels:            make(map[string]chan bool),
		Messages:            make(map[common.MapKey]common.RumorMessage),
		NextHop:             make(map[string]*net.UDPAddr),
		Peers:               parsePeerSet(peers, gossipAddrRaw),
		PrivateMessages:     make(map[string][]common.PrivateMessage),
		want:                make(map[string]uint32),
		channelsLock:        &sync.RWMutex{},
		MessagesLock:        &sync.RWMutex{},
		NextHopLock:         &sync.RWMutex{},
		PeersLock:           &sync.RWMutex{},
		PrivateMessagesLock: &sync.RWMutex{},
		wantLock:            &sync.RWMutex{},
	}
}

func (g *Gossiper) Start(rtimer int) {
	var wg sync.WaitGroup
	wg.Add(3)
	go g.listenGossip(wg)
	go g.routeRumorGenerator(wg, rtimer)
	go g.antiEntropy(wg)
	g.iterativeRumorMongering("", g.createRouteRumor())
	wg.Wait()
}
