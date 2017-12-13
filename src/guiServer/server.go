package guiServer

import (
	"github.com/pauarge/peerster/gossiper/common"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"github.com/pauarge/peerster/gossiper/gossiperServer"
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
	s.router.HandleFunc("/download", s.downloadHandler)
	s.router.HandleFunc("/file", s.fileHandler)
	s.router.HandleFunc("/id", s.idHandler)
	s.router.HandleFunc("/message", s.messageHandler)
	s.router.HandleFunc("/node", s.nodeHandler)
	s.router.HandleFunc("/search", s.searchHandler)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("static/")))

	http.Handle("/", s.router)

	log.Printf("Serving on HTTP port %s\n", common.GuiPort)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
