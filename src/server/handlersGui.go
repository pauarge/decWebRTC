package server

import (
	"net/http"
	"encoding/json"
	"github.com/pauarge/decWebRTC/src/common"
	"github.com/gorilla/websocket"
	"log"
)

var upgrader = websocket.Upgrader{}

func (g *Gossiper) echoHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	} else if g.sock != nil {
		c.WriteJSON(common.JSONRequest{Type: "alreadyGUI"})
		log.Println("")
		return
	} else {
		g.sock = c
	}

	log.Println("Web client connected through socket")
	g.sockLock.Lock()
	c.WriteJSON(common.JSONRequest{Type: "login", Name: g.name})
	g.sockLock.Unlock()
	g.sendUserList()
	defer c.Close()

	var connectedUser string

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			if connectedUser != "" {
				req := common.JSONRequest{
					Type: "leave",
					Name: g.name,
				}
				msg := common.PrivateMessage{
					Origin:      g.name,
					Destination: connectedUser,
					HopLimit:    common.MaxHops,
					Data:        req,
				}
				g.SendPrivateMessage(msg)
			}
			g.sock = nil
			log.Println("Web client disconnected")
			break
		}

		var data common.JSONRequest
		err = json.Unmarshal(message, &data)
		if err != nil {
			log.Println(err)
			break
		}

		switch data.Type {

		case "peer":
			log.Println("Adding peer from GUI")
			g.peersLock.Lock()
			g.peers[data.NewPeer] = true
			g.peersLock.Unlock()
			g.sendUserList()
			continue

		case "answer":
			log.Println("Received an answer")
			connectedUser = data.Target

		case "candidate":
			log.Println("Received a candidate")
			connectedUser = data.Target

		case "leave":
			connectedUser = ""

		default:
			log.Println("Received", data.Type)
		}

		msg := common.PrivateMessage{
			Origin:      g.name,
			Destination: data.Target,
			HopLimit:    common.MaxHops,
			Data:        data,
		}
		g.SendPrivateMessage(msg)
	}
}
