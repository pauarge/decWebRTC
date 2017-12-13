package gossiperServer

import (
	"sync"
	"github.com/pauarge/peerster/gossiper/common"
	"log"
	"github.com/dedis/protobuf"
	"time"
	"net"
)

func (g *Gossiper) listenUI(wg sync.WaitGroup) {
	defer wg.Done()
	for {
		var buf = make([]byte, common.BufferSize)
		n, _, err := g.uiConn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		m := common.PrivateMessage{}
		err = protobuf.Decode(buf[0:n], &m)
		if err == nil {
			go g.HandlePrivateMessageClient(m)
		} else {
			m := common.DataRequest{}
			err := protobuf.Decode(buf[0:n], &m)
			if err == nil {
				go g.HandleDataRequestClient(m)
			} else {
				m := common.PeerMessage{}
				err := protobuf.Decode(buf[0:n], &m)
				if err == nil {
					go g.HandlePeerMessage(m)
				} else {
					log.Println(err)
				}
			}
		}
	}
}

func (g *Gossiper) listenGossip(wg sync.WaitGroup) {
	defer wg.Done()
	for {
		var buf = make([]byte, common.BufferSize)
		n, relay, err := g.gossipConn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		g.PeersLock.Lock()
		g.Peers[getRelayStr(relay)] = true
		g.PeersLock.Unlock()
		m := common.GossipPacket{}
		err = protobuf.Decode(buf[0:n], &m)
		if err == nil {
			if m.Rumor != nil {
				go g.handleRumorMessage(*m.Rumor, relay)
			} else if m.Private != nil {
				go g.handlePrivateMessage(*m.Private)
			} else if m.Status != nil {
				go g.handleStatusPacket(*m.Status, relay)
			} else if m.DataRequest != nil {
				go g.handleDataRequest(*m.DataRequest, relay)
			} else if m.DataReply != nil {
				go g.handleDataReply(*m.DataReply, relay)
			} else if m.SearchRequest != nil {
				go g.handleSearchRequest(*m.SearchRequest, relay)
			} else if m.SearchReply != nil {
				go g.handleSearchReply(*m.SearchReply, relay)
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
		g.PeersLock.RLock()
		for i := range g.Peers {
			p = append(p, i)
		}
		g.PeersLock.RUnlock()
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

func (g *Gossiper) GetMessages(ch chan common.MessageList) {
	defer g.PrivateMessagesLock.RUnlock()
	defer g.MessagesLock.RUnlock()
	g.MessagesLock.RLock()
	g.PrivateMessagesLock.RLock()
	ml := common.MessageList{}
	for i := range g.MsgOrder {
		ml.Messages = append(ml.Messages, g.Messages[g.MsgOrder[i]])
	}
	ml.PrivateMessages = g.PrivateMessages
	ch <- ml
}
