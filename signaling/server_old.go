
 package signaling

// import (
// 	"encoding/json"
// 	"log"
// 	"net/http"
// 	"sync"

// 	"github.com/gorilla/websocket"
// )

// type Server struct {
// 	clients    map[*websocket.Conn]string // map of client connections to room names
// 	rooms      map[string]map[*websocket.Conn]bool // map of room names to client connections
// 	broadcast  chan []byte
// 	register   chan *websocket.Conn
// 	unregister chan *websocket.Conn
// 	mutex      sync.Mutex
// 	upgrader   websocket.Upgrader
// }

// func NewServer() *Server {
// 	return &Server{
// 		clients:    make(map[*websocket.Conn]string),
// 		rooms:      make(map[string]map[*websocket.Conn]bool),
// 		broadcast:  make(chan []byte),
// 		register:   make(chan *websocket.Conn),
// 		unregister: make(chan *websocket.Conn),
// 		upgrader: websocket.Upgrader{
// 			CheckOrigin: func(r *http.Request) bool {
// 				return true // Allow all origins
// 			},
// 		},
// 	}
// }

// func (s *Server) handleMessages() {
// 	for {
// 		select {
// 		case client := <-s.register:
// 			s.mutex.Lock()
// 			s.clients[client] = ""
// 			s.mutex.Unlock()
// 		case client := <-s.unregister:
// 			s.mutex.Lock()
// 			if roomName, ok := s.clients[client]; ok {
// 				delete(s.clients, client)
// 				if room, ok := s.rooms[roomName]; ok {
// 					delete(room, client)
// 					if len(room) == 0 {
// 						delete(s.rooms, roomName)
// 					}
// 				}
// 			}
// 			s.mutex.Unlock()
// 		case message := <-s.broadcast:
// 			var msg map[string]interface{}
// 			if err := json.Unmarshal(message, &msg); err != nil {
// 				log.Printf("Error unmarshaling message: %v", err)
// 				continue
// 			}

// 			roomName, ok := msg["roomName"].(string)
// 			if !ok {
// 				log.Println("Room name not found in message")
// 				continue
// 			}

// 			s.mutex.Lock()
// 			if room, ok := s.rooms[roomName]; ok {
// 				for client := range room {
// 					err := client.WriteMessage(websocket.TextMessage, message)
// 					if err != nil {
// 						log.Printf("Error sending message: %v", err)
// 						client.Close()
// 						delete(room, client)
// 						delete(s.clients, client)
// 					}
// 				}
// 			}
// 			s.mutex.Unlock()
// 		}
// 	}
// }

// func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
// 	ws, err := s.upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	defer ws.Close()

// 	s.register <- ws

// 	for {
// 		_, msg, err := ws.ReadMessage()
// 		if err != nil {
// 			log.Printf("Error reading message: %v", err)
// 			s.unregister <- ws
// 			break
// 		}

// 		var message map[string]interface{}
// 		if err := json.Unmarshal(msg, &message); err != nil {
// 			log.Printf("Error unmarshaling message: %v", err)
// 			continue
// 		}

// 		switch message["type"] {
// 		case "create":
// 			s.handleCreateRoom(ws, message)
// 		case "join":
// 			s.handleJoinRoom(ws, message)
// 		case "leave":
// 			s.handleLeaveRoom(ws, message)
// 		default:
// 			s.broadcast <- msg
// 		}
// 	}
// }

// func (s *Server) handleCreateRoom(ws *websocket.Conn, message map[string]interface{}) {
// 	roomName, ok := message["roomName"].(string)
// 	if !ok {
// 		log.Println("Invalid room name")
// 		return
// 	}

// 	s.mutex.Lock()
// 	defer s.mutex.Unlock()

// 	if _, exists := s.rooms[roomName]; exists {
// 		ws.WriteJSON(map[string]string{"type": "error", "message": "Room already exists"})
// 		return
// 	}

// 	s.rooms[roomName] = make(map[*websocket.Conn]bool)
// 	s.rooms[roomName][ws] = true
// 	s.clients[ws] = roomName

// 	ws.WriteJSON(map[string]string{"type": "created", "roomName": roomName})
// }

// func (s *Server) handleJoinRoom(ws *websocket.Conn, message map[string]interface{}) {
// 	roomName, ok := message["roomName"].(string)
// 	if !ok {
// 		log.Println("Invalid room name")
// 		return
// 	}

// 	s.mutex.Lock()
// 	defer s.mutex.Unlock()

// 	if room, exists := s.rooms[roomName]; exists {
// 		room[ws] = true
// 		s.clients[ws] = roomName
// 		ws.WriteJSON(map[string]string{"type": "joined", "roomName": roomName})

// 		// Notify other participants in the room
// 		for client := range room {
// 			if client != ws {
// 				client.WriteJSON(map[string]string{"type": "participant_joined", "roomName": roomName})
// 			}
// 		}
// 	} else {
// 		ws.WriteJSON(map[string]string{"type": "error", "message": "Room does not exist"})
// 	}
// }

// func (s *Server) handleLeaveRoom(ws *websocket.Conn, message map[string]interface{}) {
// 	s.mutex.Lock()
// 	defer s.mutex.Unlock()

// 	if roomName, ok := s.clients[ws]; ok {
// 		if room, exists := s.rooms[roomName]; exists {
// 			delete(room, ws)
// 			if len(room) == 0 {
// 				delete(s.rooms, roomName)
// 			} else {
// 				// Notify other participants in the room
// 				for client := range room {
// 					client.WriteJSON(map[string]string{"type": "participant_left", "roomName": roomName})
// 				}
// 			}
// 		}
// 		delete(s.clients, ws)
// 	}

// 	ws.WriteJSON(map[string]string{"type": "left"})
// }