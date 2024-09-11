package signaling

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type JoinRoomMessage struct {
	Type string `json:"type"`
	Room string `json:"room"`
}
type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	onMessage func([]byte)
	mutex     sync.Mutex
}

func NewClient() *Client {
	return &Client{
		send: make(chan []byte, 256),
	}
}

func (c *Client) CreateRoom(roomName string) error {
	message := Message{
		Type: "create",
		Data: map[string]interface{}{
			"roomName": roomName,
		},
	}
	return c.sendJSON(message)
}

func (c *Client) JoinRoom(roomName string) error {
	message := Message{
		Type: "join",
		Data: map[string]interface{}{
			"roomName": roomName,
		},
	}
	
	return c.sendJSON(message)
}

func (c *Client) LeaveRoom(roomName string) error {
	message := Message{
		Type: "leave",
		Data: map[string]interface{}{
			"roomName": roomName,
		},
	}
	return c.sendJSON(message)
}

func (c *Client) Connect(url string) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	c.conn = conn

	go c.readPump()
	go c.writePump()

	return nil
}

func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}
	close(c.send)
}

func (c *Client) Send(message []byte) {
	c.send <- message
}

func (c *Client) SendtoJoin(messageJSON []byte) error {
	err := c.conn.WriteMessage(websocket.TextMessage, messageJSON)
	c.send <- messageJSON
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)

	}
	return nil
}

func (c *Client) SetOnMessage(callback func([]byte)) {
	c.onMessage = callback
}

func (c *Client) readPump() {
	defer c.Close()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if c.onMessage != nil {
			c.onMessage(message)
		}
	}
}

func (c *Client) writePump() {
	defer c.Close()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}

func (c *Client) SendOffer(roomName string, offer interface{}) error {
	message := Message{
		Type: "offer",
		Data: map[string]interface{}{
			"roomName": roomName,
			"offer":    offer,
		},
	}
	return c.sendJSON(message)
}

func (c *Client) SendAnswer(roomName string, answer interface{}) error {
	message := Message{
		Type: "answer",
		Data: map[string]interface{}{
			"roomName": roomName,
			"answer":   answer,
		},
	}
	return c.sendJSON(message)
}

func (c *Client) SendICECandidate(roomName string, candidate interface{}) error {
	message := Message{
		Type: "ice-candidate",
		Data: map[string]interface{}{
			"roomName": roomName,
			"candidate": candidate,
		},
	}
	return c.sendJSON(message)
}

func (c *Client) sendJSON(v interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.conn.WriteJSON(v)
}
