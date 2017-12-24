package gossiperServer

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

		var req common.JSONRequest
		switch data.Type {
		case "offer":
			log.Println("Sending offer to " + data.Name)
			req = common.JSONRequest{
				Type:  "offer",
				Offer: data.Offer,
				Name:  g.name,
			}

		case "answer":
			log.Println("Received an answer")
			connectedUser = data.Name
			req = common.JSONRequest{
				Type:   "answer",
				Answer: data.Answer,
			}

		case "candidate":
			log.Println("Received a candidate")
			connectedUser = data.Name
			req = common.JSONRequest{
				Type:      "candidate",
				Candidate: data.Candidate,
			}

		case "leave":
			log.Println("Received a leave")
			connectedUser = ""
			req = common.JSONRequest{
				Type: "leave",
			}

		default:
			log.Println("Did not understand the command")
			continue
		}

		msg := common.PrivateMessage{
			Origin:      g.name,
			Destination: data.Name,
			HopLimit:    common.MaxHops,
			Data:        req,
		}
		g.SendPrivateMessage(msg)
	}
}
