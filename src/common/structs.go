package common

import (
	"net"
)

type MapKey struct {
	Origin    string
	MessageId uint32
}

// Messages

type RumorMessage struct {
	Origin   string
	Id       uint32
	LastIP   *net.IP
	LastPort *int
}

type PrivateMessage struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Id          uint32
}

type PeerStatus struct {
	Identifier string
	NextId     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

type GossipPacket struct {
	Rumor   *RumorMessage
	Status  *StatusPacket
	Private *PrivateMessage
}

// GUI Server Responses

type StatusResponse struct {
	Status string
}

type IdResponse struct {
	Id string
}

type PeerListResponse struct {
	Peers []string
	Hops  []string
}

// PeerList sorting

type PeerStatusList []PeerStatus

func (a PeerStatusList) Len() int      { return len(a) }
func (a PeerStatusList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a PeerStatusList) Less(i, j int) bool {
	if a[i].Identifier < a[j].Identifier {
		return true
	}
	if a[i].Identifier > a[j].Identifier {
		return false
	}
	return a[i].NextId < a[j].NextId
}
