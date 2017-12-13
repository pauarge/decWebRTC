package guiServer

import (
	"net/http"
	"encoding/json"
	"strings"
	"sort"
	"github.com/pauarge/peerster/gossiper/common"
	"encoding/hex"
	"log"
)

func (s *Server) messageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		r.ParseForm()
		if r.Form["Destination"] != nil {
			text := strings.Join(r.Form["Message"], "")
			dest := strings.Join(r.Form["Destination"], "")
			go s.gossiper.HandlePrivateMessageClient(common.PrivateMessage{Destination: dest, Text: text})
		} else {
			go s.gossiper.HandlePeerMessage(common.PeerMessage{Text: strings.Join(r.Form["Message"], "")})
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(common.StatusResponse{"message sent"})
	} else {
		ch := make(chan common.MessageList)
		go s.gossiper.GetMessages(ch)
		ml := <-ch
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ml)
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

func (s *Server) fileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		r.ParseForm()
		path := strings.Join(r.Form["Path"], "")
		go s.gossiper.HandleFileUpload(path)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(common.StatusResponse{"file uploaded"})
	}
}

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		r.ParseForm()
		node := strings.Join(r.Form["Destination"], "")
		filename := strings.Join(r.Form["FileName"], "")
		hash := strings.Join(r.Form["HashValue"], "")
		h, err := hex.DecodeString(hash)
		if err != nil {
			log.Fatal(err)
		}
		msg := common.DataRequest{Destination: node, FileName: filename, HashValue: h}
		go s.gossiper.HandleDataRequestClient(msg)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(common.StatusResponse{"file requested"})
	}
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		r.ParseForm()
		keywords := strings.Join(r.Form["Keywords"], "")
		budget := strings.Join(r.Form["Budget"], "")
		go s.gossiper.HandleKeywords(keywords, budget)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(common.StatusResponse{"search requested"})
	}
}
