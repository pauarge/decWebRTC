package guiServer

import (
	"net/http"
	"encoding/json"
	"strings"
	"sort"
	"github.com/pauarge/decWebRTC/src/common"
	"github.com/gorilla/websocket"
	"log"
)

var upgrader = websocket.Upgrader{}

func (s *Server) echoHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var data common.JSONRequest
		err = json.Unmarshal(message, &data)
		if err != nil {
			log.Println(err)
			break
		}

		switch data.Type {
		case "login":
			log.Println("USER LOGGED " + data.Name)
		case "offer":
			log.Println("Received an offer")
		case "answer":
			log.Println("Received an answer")
		case "candidate":
			log.Println("Received a candidate")
		case "leave":
			log.Println("Received a leave")
		default:
			log.Println("Did not understand the command")
		}

		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (s *Server) nodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		r.ParseForm()
		s.gossiper.Peers[strings.Join(r.Form["Address"], "")] = true
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(common.StatusResponse{"peer added"})
	} else {
		w.WriteHeader(http.StatusOK)
		var keys []string
		var hops []string
		s.gossiper.PeersLock.RLock()
		for k := range s.gossiper.Peers {
			keys = append(keys, k)
		}
		s.gossiper.PeersLock.RUnlock()
		sort.Strings(keys)
		res := common.PeerListResponse{}
		for _, k := range keys {
			res.Peers = append(res.Peers, k)
		}
		s.gossiper.NextHopLock.RLock()
		for k := range s.gossiper.NextHop {
			hops = append(hops, k)
		}
		s.gossiper.NextHopLock.RUnlock()
		sort.Strings(hops)
		res.Hops = hops
		json.NewEncoder(w).Encode(res)
	}
}

func (s *Server) idHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		r.ParseForm()
		s.gossiper.Name = strings.Join(r.Form["Id"], "")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(common.StatusResponse{"node name updated"})
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(common.IdResponse{s.gossiper.Name})
	}
}
