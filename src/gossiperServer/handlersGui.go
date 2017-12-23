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
		log.Print("upgrade:", err)
		return
	} else {
		g.sock = c
	}

	g.sockLock.Lock()
	c.WriteJSON(common.JSONRequest{Type: "login", Name: g.name})
	g.sockLock.Unlock()
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
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
			req = common.JSONRequest{
				Type:   "answer",
				Answer: data.Answer,
			}

		case "candidate":
			log.Println("Received a candidate")
			req = common.JSONRequest{
				Type:      "candidate",
				Candidate: data.Candidate,
			}

		case "leave":
			log.Println("Received a leave")
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
			Data:        &req,
		}
		g.SendPrivateMessage(msg)
	}
}
