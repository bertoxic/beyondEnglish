package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/mux"

	"github.com/bertoxic/beyondEnglish/peers"
	"github.com/bertoxic/beyondEnglish/signaling"
)

type Server struct {
	router          *mux.Router
	signalingServer *signaling.Server
	peerManager     *peers.PeerManager
}

func NewServer() *Server {
	router := mux.NewRouter()
	signalingServer := signaling.NewServer()
	peerManager := peers.NewPeerManager(signalingServer)

	server := &Server{
		router:          router,
		signalingServer: signalingServer,
		peerManager:     peerManager,
	}

	server.routes()
	return server
}

func (s *Server) routes() {
	s.router.HandleFunc("/api/rooms", s.handleCreateRoom).Methods("POST")
	s.router.HandleFunc("/api/rooms/{roomID}/join", s.handleJoinRoom).Methods("POST")
	s.router.HandleFunc("/", s.renderhomepage).Methods("GET")
	s.router.HandleFunc("/ws", s.signalingServer.HandleWebSocket)

	// Serve static files
	//s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/build")))
}

func (s *Server) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		return
	}
	print("Request to handlecreateroom")
	roomID := s.peerManager.CreateRoom()
	json.NewEncoder(w).Encode(map[string]string{"roomID": roomID})
}

func (s *Server) renderhomepage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		return
	}
	renderHTMLTemplates(w, "template.html", nil)
	x, err := os.Getwd()
	if err != nil {
		println("error occureeed")
	}
	fmt.Println("Current Directory: ", x, " okko0o")

	// roomID := s.peerManager.CreateRoom()
	// json.NewEncoder(w).Encode(map[string]string{"roomID": roomID})

}

func (s *Server) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	roomID := vars["roomID"]

	if !s.peerManager.RoomExists(roomID) {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// The actual joining of the room will happen through WebSocket
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func main() {
	server := NewServer()

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", server.router))
}

func renderHTMLTemplates(w http.ResponseWriter, tmpl string, data interface{}) {

	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
