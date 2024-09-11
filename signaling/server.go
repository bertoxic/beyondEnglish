package signaling

import (
	"encoding/json"
	"fmt"
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
	print(fmt.Sprintf("Added peer %s", peerID))
}

func (s *Server) RemovePeer(peerID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Remove the peer from the map
	delete(s.peers, peerID)
}
func (s *Server) GetPeerIDByConn(conn *websocket.Conn) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for peerID, peerConn := range s.peers {
		if peerConn == conn {
			return peerID, true
		}
	}

	// Return false if no matching connection is found
	return "", false
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
	conn.WriteJSON(Message{
		Type: "peer-id",
		Data: map[string]interface{}{"peerID": peerID},
	})
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
	case "offer":
		room, ok := message.Data.(map[string]interface{})["roomName"].(string)
		if !ok {
			log.Println("Invalid room data in offer message")
			return
		}
		offer := message.Data.(map[string]interface{})["offer"]
		s.handleOffer(conn, room, offer)
	case "answer":
		room, ok := message.Data.(map[string]interface{})["roomName"].(string)
		if !ok {
			log.Println("Invalid room data in answer message")
			return
		}
		answer := message.Data.(map[string]interface{})["answer"]
		s.handleAnswer(conn, room, answer)
	case "ice-candidate":
		room, ok := message.Data.(map[string]interface{})["roomName"].(string)
		if !ok {
			log.Println("Invalid room data in ice-candidate message")
			return
		}
		candidate := message.Data.(map[string]interface{})["candidate"]
		s.handleICECandidate(conn, room, candidate)
	default:
		log.Println("Unknown message type:", message.Type)
	}
}
func (s *Server) handleAnswer(conn *websocket.Conn, room string, answer interface{}) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if connections, ok := s.rooms[room]; ok {
		for otherConn := range connections {
			if otherConn != conn {
				message := Message{
					Type: "answer",
					Data: map[string]interface{}{
						"roomName": room,
						"answer":   answer,
					},
				}
				if err := otherConn.WriteJSON(message); err != nil {
					log.Printf("error sending answer to peer: %v", err)
				}
			}
		}
	}
}

func (s *Server) handleICECandidate(conn *websocket.Conn, room string, candidate interface{}) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if connections, ok := s.rooms[room]; ok {
		for otherConn := range connections {
			if otherConn != conn {
				message := Message{
					Type: "ice-candidate",
					Data: map[string]interface{}{
						"roomName":  room,
						"candidate": candidate,
					},
				}
				if err := otherConn.WriteJSON(message); err != nil {
					log.Printf("error sending ICE candidate to peer: %v", err)
				}
			}
		}
	}
}
func (s *Server) handleOffer(conn *websocket.Conn, room string, offer interface{}) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if connections, ok := s.rooms[room]; ok {
		for otherConn := range connections {
			if otherConn != conn {
				message := Message{
					Type: "offer",
					Data: map[string]interface{}{
						"roomName": room,
						"offer":    offer,
					},
				}
				if err := otherConn.WriteJSON(message); err != nil {
					log.Printf("error sending offer to peer: %v", err)
				}
			}
		}
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

	var peerID string
	for id, peerConn := range s.peers {
		if peerConn == conn {
			peerID = id
			break
		}
	}

	// Auto-join the creator to the room
	s.rooms[roomName][conn] = true
	log.Printf("Connection about to user %s auto-joined to room %s", peerID, roomName)

	// Notify other users (if any) about the new user
	data := map[string]string{"peerID": peerID, "roomName": roomName}
	message := Message{
		Type: "user-joined",
		Data: data,
	}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}
	//s.broadcastToRoom(conn, roomName, []byte(fmt.Sprintf(`{"type": "user-joined", "data": {"peerID": "%s", "roomName": "%s"}}`, peerID, roomName)))
	s.broadcastToRoom(conn, roomName, messageBytes)

	log.Printf("Connection has now auto-joined to room %s", roomName)

}

func (s *Server) handleJoin(conn *websocket.Conn, room string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// if _, ok := s.rooms[room]; !ok {
	// 	s.rooms[room] = make(map[*websocket.Conn]bool)
	// }
	if _, ok := s.rooms[room]; !ok {
		log.Printf("Room %s does not exist ohh", room)
		log.Print("rooom lissst is...", s.rooms)
		err := conn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "message": "Room does not exist oo"}`))
		if err != nil {
			log.Printf("error sending error message to peer: %v", err)
			return
		}

	}

	s.rooms[room][conn] = true

	// Find the peerID for this connection
	var peerID string
	for id, peerConn := range s.peers {
		if peerConn == conn {
			peerID = id
			break
		}
	}

	// Notify other users in the room about the new user
	for otherConn := range s.rooms[room] {
		if otherConn != conn {
			otherConn.WriteJSON(Message{
				Type: "user-joined",
				Data: map[string]interface{}{
					"peerID":   peerID,
					"roomName": room,
				},
			})
		}
	}

	log.Printf("Userboss %s joined room: %s", peerID, room)

	s.broadcastToRoom(conn,
		room,
		[]byte(fmt.Sprintf(`{"type": "user-joined", "data": {"peerID": "%s"}}`, peerID)))

}

func (s *Server) handleLeave(conn *websocket.Conn, room string) {
	s.removeConnection(conn)
	s.broadcastToRoom(conn, room, []byte(`{"type": "user-left"}`))
	log.Printf("User %s left room: %s", s.peers, room)

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
	log.Print("broadcasting to room 000y HELLLPP")
	// s.mutex.Lock()
	// defer s.mutex.Unlock()
	log.Print("broadcasting to room HELLLPP")
	if connections, ok := s.rooms[room]; ok {
		for conn := range connections {
			//if conn != sender {
			conn.WriteMessage(websocket.TextMessage, message)
			//}
		}
	} else {

		log.Print("did not broadcast to room HELLLPP")
	}
}
