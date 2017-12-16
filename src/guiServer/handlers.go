package guiServer

import (
	"net/http"
	"encoding/json"
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

	c.WriteJSON(common.JSONRequest{Type: "login", Success: true, Name: s.gossiper.Name})
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
			msg := common.PrivateMessage{
				Origin: s.gossiper.Name,
				Destination: data.Name,
				HopLimit: common.MaxHops,
				Data: data.Offer,
			}
			s.gossiper.SendPrivateMessage(msg)


		case "answer":
			log.Println("Received an answer")

		case "candidate":
			log.Println("Received a candidate")
		case "leave":
			log.Println("Received a leave")
		default:
			log.Println("Did not understand the command")
		}
	}
}
