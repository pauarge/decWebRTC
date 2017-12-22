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
	c.WriteJSON(common.JSONRequest{Type: "login", Success: true, Name: g.Name})
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

		switch data.Type {
		case "offer":
			log.Println("Sending offer to " + data.Name)
			req := common.JSONRequest{
				Type:  "offer",
				Offer: data.Offer,
				Name:  g.Name,
			}
			msg := common.PrivateMessage{
				Origin:      g.Name,
				Destination: data.Name,
				HopLimit:    common.MaxHops,
				Data:        &req,
			}
			g.SendPrivateMessage(msg)

		case "answer":
			log.Println("Received an answer")
			req := common.JSONRequest{
				Type: "answer",
				Answer: data.Answer,
			}
			msg := common.PrivateMessage{
				Origin:      g.Name,
				Destination: data.Name,
				HopLimit:    common.MaxHops,
				Data:        &req,
			}
			g.SendPrivateMessage(msg)

		case "candidate":
			log.Println("Received a candidate")
			req := common.JSONRequest{
				Type: "candidate",
				Candidate: data.Candidate,
			}
			msg := common.PrivateMessage{
				Origin:      g.Name,
				Destination: data.Name,
				HopLimit:    common.MaxHops,
				Data:        &req,
			}
			g.SendPrivateMessage(msg)

			/*case "leave":
				log.Println("Received a leave")
				msg := common.PrivateMessage{
					Origin: s.gossiper.Name,
					Destination: data.Name,
					HopLimit: common.MaxHops,
					Data: "leave",
				}*/

		default:
			log.Println("Did not understand the command")
		}
	}
}
