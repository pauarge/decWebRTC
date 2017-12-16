package guiServer

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/pauarge/decWebRTC/src/common"
	"github.com/pauarge/decWebRTC/src/gossiperServer"
)

type Server struct {
	router   *mux.Router
	gossiper *gossiperServer.Gossiper
}

func NewServer(g *gossiperServer.Gossiper) *Server {
	return &Server{
		router:   mux.NewRouter(),
		gossiper: g,
	}
}

func (s *Server) Start() {
	s.router.HandleFunc("/echo", s.echoHandler)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("static/")))

	http.Handle("/", s.router)

	log.Printf("Serving on HTTP port %s\n", common.GuiPort)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
