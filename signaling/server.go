package signaling

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Server struct {
	rooms          map[string]map[*websocket.Conn]bool
	mutex          sync.RWMutex
	peers          map[string]*websocket.Conn
	upgrader       websocket.Upgrader
	messageHandler func(Message)
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewServer() *Server {
	return &Server{
		rooms: make(map[string]map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
		},
		peers: make(map[string]*websocket.Conn),
	}
}
func (s *Server) Start() {
	// This method is now empty, as we're handling WebSocket connections in HandleWebSocket
}

func (s *Server) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, connections := range s.rooms {
		for conn := range connections {
			conn.Close()
		}
	}
	s.peers = make(map[string]*websocket.Conn) // Clear the peers map
}
func (s *Server) SetMessageHandler(handler func(Message)) {
	s.messageHandler = handler
}

func (s *Server) SendToPeer(peerID string, message Message) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Look up the connection associated with the peerID
	conn, ok := s.peers[peerID]
	if !ok {
		log.Printf("peerID %s not found", peerID)
		return
	}

	// Send the message to the specific peer
	if err := conn.WriteJSON(message); err != nil {
		log.Printf("error sending message to peer %s: %v", peerID, err)
	}
}

func (s *Server) AddPeer(peerID string, conn *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Add the peer to the map
	s.peers[peerID] = conn
}

func (s *Server) RemovePeer(peerID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Remove the peer from the map
	delete(s.peers, peerID)
}

// Update handleMessage to use the new Message type and call the messageHandler
// func (s *Server) handleMessage(conn *websocket.Conn, msg []byte) {
// 	var message Message
// 	if err := json.Unmarshal(msg, &message); err != nil {
// 		log.Printf("error parsing message: %v", err)
// 		return
// 	}

// 	if s.messageHandler != nil {
// 		s.messageHandler(message)
// 	}

//		// Existing switch statement can be removed or modified based on your needs
//	}
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		return
	}
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	peerID := r.URL.Query().Get("peerID")
	if peerID == "" {
		peerID = uuid.New().String() // Replace with a real UUID generator

	}

	// Add the peer to the server
	s.AddPeer(peerID, conn)

	defer s.RemovePeer(peerID)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			s.removeConnection(conn)
			break
		}
		s.handleMessage(conn, msg)
	}
}

func (s *Server) handleMessage(conn *websocket.Conn, msg []byte) {
	var message Message
	if err := json.Unmarshal(msg, &message); err != nil {
		log.Printf("error parsing message: %v", err)
		return
	}

	if s.messageHandler != nil {
		s.messageHandler(message)
	}

	switch message.Type {
	case "create":
		log.Println("message is create")
		log.Println(message.Data)
		room, ok := message.Data.(map[string]interface{})["roomName"].(string)

		if !ok {
			log.Println("Invalid room data in create message")
			return
		}
		s.handleCreate(conn, room)
	case "join":
		log.Println("message is join")

		// Assuming message.Data is a map or struct with a "room" field
		room, ok := message.Data.(map[string]interface{})["roomName"].(string)
		if !ok {
			log.Println("Invalid room data in join message")
			return
		}
		s.handleJoin(conn, room)
	case "leave":
		log.Println("message is leave")

		room, ok := message.Data.(map[string]interface{})["roomName"].(string)
		if !ok {
			log.Println("Invalid room data in leave message")
			return
		}
		s.handleLeave(conn, room)
	case "offer", "answer", "ice-candidate":
		room, ok := message.Data.(map[string]interface{})["roomName"].(string)
		if !ok {
			log.Println("Invalid room data in signaling message")
			return
		}
		s.broadcastToRoom(conn, room, msg)
	default:
		log.Println("Unknown message type:", message.Type)
	}
}

func (s *Server) handleCreate(conn *websocket.Conn, roomName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.rooms[roomName]; exists {
		log.Printf("Room %s already exists", roomName)
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "message": "Room already exists"}`))
		return
	}

	// Create the room
	s.rooms[roomName] = make(map[*websocket.Conn]bool)
	log.Printf("Room %s created", roomName)

	// Auto-join the creator to the room
	s.rooms[roomName][conn] = true
	log.Printf("Connection auto-joined to room %s", roomName)

	// Notify other users (if any) about the new user
	s.broadcastToRoom(conn, roomName, []byte(`{"type": "user-joined", "roomName": "`+roomName+`"}`))
}

func (s *Server) handleJoin(conn *websocket.Conn, room string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// if _, ok := s.rooms[room]; !ok {
	// 	s.rooms[room] = make(map[*websocket.Conn]bool)
	// }
	if _, ok := s.rooms[room]; !ok {
		log.Printf("Room %s does not exist", room)
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "message": "Room does not exist"}`))
		return
	}
	s.rooms[room][conn] = true

	// Notify other users in the room about the new user
	s.broadcastToRoom(conn, room, []byte(`{"type": "user-joined"}`))
}

func (s *Server) handleLeave(conn *websocket.Conn, room string) {
	s.removeConnection(conn)
	s.broadcastToRoom(conn, room, []byte(`{"type": "user-left"}`))
}

func (s *Server) removeConnection(conn *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for room, connections := range s.rooms {
		if _, ok := connections[conn]; ok {
			delete(connections, conn)
			if len(connections) == 0 {
				delete(s.rooms, room)
			}
			break
		}
	}
	// Remove from peers
	for peerID, peerConn := range s.peers {
		if peerConn == conn {
			delete(s.peers, peerID)
			break
		}
	}
}

func (s *Server) broadcastToRoom(sender *websocket.Conn, room string, message []byte) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if connections, ok := s.rooms[room]; ok {
		for conn := range connections {
			if conn != sender {
				conn.WriteMessage(websocket.TextMessage, message)
			}
		}
	}
}
