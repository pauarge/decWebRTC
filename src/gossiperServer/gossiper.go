package gossiperServer

import (
	"net"

	"log"
	"sync"
	"github.com/pauarge/peerster/gossiper/common"
)

type Gossiper struct {
	counter             uint32
	Name                string
	noforward           bool
	uiConn              *net.UDPConn
	gossipConn          *net.UDPConn
	channels            map[string]chan bool
	dataRequests        map[string]chan bool
	downloadMetas       map[string]chan bool
	files               map[string]common.StoredFile
	Messages            map[common.MapKey]common.RumorMessage
	NextHop             map[string]*net.UDPAddr
	Peers               map[string]bool
	PrivateMessages     map[string][]common.PrivateMessage
	searchRequests      map[common.MapKeySearchRequest]bool
	want                map[string]uint32
	channelsLock        *sync.RWMutex
	dataRequestsLock    *sync.RWMutex
	downloadMetasLock   *sync.RWMutex
	filesLock           *sync.RWMutex
	MessagesLock        *sync.RWMutex
	MsgOrderLock        *sync.RWMutex
	NextHopLock         *sync.RWMutex
	PeersLock           *sync.RWMutex
	PrivateMessagesLock *sync.RWMutex
	searchRequestsLock  *sync.RWMutex
	wantLock            *sync.RWMutex
	MsgOrder            []common.MapKey
}

func NewGossiper(uiPort int, gossipAddrRaw, name, peers string, noforward bool) *Gossiper {
	uiAddr := &net.UDPAddr{
		Port: uiPort,
		IP:   net.ParseIP(common.LocalIp),
	}
	gossipAddr, err := net.ResolveUDPAddr("udp4", gossipAddrRaw)
	if err != nil {
		log.Fatal(err)
	}
	uiConn, err := net.ListenUDP("udp", uiAddr)
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
		noforward:           noforward,
		uiConn:              uiConn,
		gossipConn:          gossipConn,
		channels:            make(map[string]chan bool),
		dataRequests:        make(map[string]chan bool),
		downloadMetas:       make(map[string]chan bool),
		files:               make(map[string]common.StoredFile),
		Messages:            make(map[common.MapKey]common.RumorMessage),
		NextHop:             make(map[string]*net.UDPAddr),
		Peers:               parsePeerSet(peers, gossipAddrRaw),
		PrivateMessages:     make(map[string][]common.PrivateMessage),
		searchRequests:      make(map[common.MapKeySearchRequest]bool),
		want:                make(map[string]uint32),
		channelsLock:        &sync.RWMutex{},
		dataRequestsLock:    &sync.RWMutex{},
		downloadMetasLock:   &sync.RWMutex{},
		filesLock:           &sync.RWMutex{},
		MessagesLock:        &sync.RWMutex{},
		MsgOrderLock:        &sync.RWMutex{},
		NextHopLock:         &sync.RWMutex{},
		PeersLock:           &sync.RWMutex{},
		PrivateMessagesLock: &sync.RWMutex{},
		searchRequestsLock:  &sync.RWMutex{},
		wantLock:            &sync.RWMutex{},
		MsgOrder:            nil,
	}
}

func (g *Gossiper) Start(rtimer int) {
	var wg sync.WaitGroup
	wg.Add(4)
	go g.listenUI(wg)
	go g.listenGossip(wg)
	go g.routeRumorGenerator(wg, rtimer)
	go g.antiEntropy(wg)
	g.iterativeRumorMongering("", g.createRouteRumor())
	wg.Wait()
}
