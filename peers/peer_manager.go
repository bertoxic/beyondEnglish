// File: peers/peer_manager.go (updated)

package peers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"

	//"github.com/bertoxic/beyondEnglish/media"
	"github.com/bertoxic/beyondEnglish/signaling"
	"github.com/bertoxic/beyondEnglish/storage"
)

type Room struct {
	ID       string
	Peers    map[string]*PeerConnection
	Recorder *storage.FileRecorder
}

type PeerManager struct {
	signalingServer *signaling.Server
	rooms           map[string]*Room
	mutex           sync.Mutex
}

func NewPeerManager(signalingServer *signaling.Server) *PeerManager {
	pm := &PeerManager{
		signalingServer: signalingServer,
		rooms:           make(map[string]*Room),
	}

	signalingServer.SetMessageHandler(pm.handleSignalingMessage)
	return pm
}

func (pm *PeerManager) CreateRoom() string {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	roomID := uuid.New().String()
	pm.rooms[roomID] = &Room{
		ID:    roomID,
		Peers: make(map[string]*PeerConnection),
	}
	return roomID
}

func (pm *PeerManager) RoomExists(roomID string) bool {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	_, exists := pm.rooms[roomID]
	return exists
}

func (pm *PeerManager) JoinRoom(roomID string, peerID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	room, exists := pm.rooms[roomID]
	if !exists {
		return fmt.Errorf("room %s does not exist", roomID)
	}

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := NewPeerConnection(config)
	if err != nil {
		return fmt.Errorf("failed to create peer connection: %w", err)
	}

	room.Peers[peerID] = peerConnection

	// Set up event handlers for the peer connection
	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		pm.signalingServer.SendToPeer(peerID, signaling.Message{
			Type: "ice-candidate",
			Data: c.ToJSON(),
		})
	})

	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("Received remote track from peer %s in room %s", peerID, roomID)
		if room.Recorder != nil {
			room.Recorder.AddTrack(remoteTrack)
		}
	})

	return nil
}

func (pm *PeerManager) LeaveRoom(roomID string, peerID string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	room, exists := pm.rooms[roomID]
	if !exists {
		return
	}

	peerConnection, exists := room.Peers[peerID]
	if exists {
		peerConnection.Close()
		delete(room.Peers, peerID)
	}

	if len(room.Peers) == 0 {
		if room.Recorder != nil {
			room.Recorder.Stop()
		}
		delete(pm.rooms, roomID)
	}
}

func (pm *PeerManager) handleSignalingMessage(message signaling.Message) {
	switch message.Type {
	case "join-room":
		var joinMessage struct {
			RoomID string `json:"roomID"`
			PeerID string `json:"peerID"`
		}
        dataBytes, err := json.Marshal(message.Data)
        if err != nil {
            log.Printf("Failed to marshal message.Data: %v", err)
            return
        }

        if err := json.Unmarshal(dataBytes, &joinMessage); err != nil {
            log.Printf("Failed to unmarshal join-room message: %v", err)
            return
        }
		err = pm.JoinRoom(joinMessage.RoomID, joinMessage.PeerID)
		if err != nil {
			log.Printf("Failed to join room: %v", err)
			return
		}
		// Notify other peers in the room
		pm.notifyPeersInRoom(joinMessage.RoomID, joinMessage.PeerID, "peer-joined")

	case "leave-room":
		var leaveMessage struct {
			RoomID string `json:"roomID"`
			PeerID string `json:"peerID"`
		}
        dataBytes, err := json.Marshal(message.Data)
        if err != nil {
            log.Printf("Failed to marshal message.Data: %v", err)
            return
        }

        if err := json.Unmarshal(dataBytes, &leaveMessage); err != nil {
            log.Printf("Failed to unmarshal join-room message: %v", err)
            return
        }
		pm.LeaveRoom(leaveMessage.RoomID, leaveMessage.PeerID)
		// Notify other peers in the room
		pm.notifyPeersInRoom(leaveMessage.RoomID, leaveMessage.PeerID, "peer-left")

	case "offer":
		var offerMessage struct {
			RoomID string                    `json:"roomID"`
			PeerID string                    `json:"peerID"`
			Offer  webrtc.SessionDescription `json:"offer"`
		}
        dataBytes, err := json.Marshal(message.Data)
        if err != nil {
            log.Printf("Failed to marshal message.Data: %v", err)
            return
        }

        if err := json.Unmarshal(dataBytes, &offerMessage); err != nil {
            log.Printf("Failed to unmarshal join-room message: %v", err)
            return
        }
		pm.handleOffer(offerMessage.RoomID, offerMessage.PeerID, offerMessage.Offer)

	case "answer":
		var answerMessage struct {
			RoomID string                    `json:"roomID"`
			PeerID string                    `json:"peerID"`
			Answer webrtc.SessionDescription `json:"answer"`
		}
        dataBytes, err := json.Marshal(message.Data)
        if err != nil {
            log.Printf("Failed to marshal message.Data: %v", err)
            return
        }

        if err := json.Unmarshal(dataBytes, &answerMessage); err != nil {
            log.Printf("Failed to unmarshal join-room message: %v", err)
            return
        }
		pm.handleAnswer(answerMessage.RoomID, answerMessage.PeerID, answerMessage.Answer)

	case "ice-candidate":
		var candidateMessage struct {
			RoomID    string                  `json:"roomID"`
			PeerID    string                  `json:"peerID"`
			Candidate webrtc.ICECandidateInit `json:"candidate"`
		}
        dataBytes, err := json.Marshal(message.Data)
        if err != nil {
            log.Printf("Failed to marshal message.Data: %v", err)
            return
        }

        if err := json.Unmarshal(dataBytes, &candidateMessage); err != nil {
            log.Printf("Failed to unmarshal join-room message: %v", err)
            return
        }
		pm.handleICECandidate(candidateMessage.RoomID, candidateMessage.PeerID, candidateMessage.Candidate)
	}
}

func (pm *PeerManager) handleOffer(roomID string, peerID string, offer webrtc.SessionDescription) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	room, exists := pm.rooms[roomID]
	if !exists {
		log.Printf("Room %s does not exist", roomID)
		return
	}

	peerConnection, exists := room.Peers[peerID]
	if !exists {
		log.Printf("Peer %s not found in room %s", peerID, roomID)
		return
	}

	err := peerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.Printf("Failed to set remote description: %v", err)
		return
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Printf("Failed to create answer: %v", err)
		return
	}

	err = peerConnection.SetLocalDescription(*answer)
	if err != nil {
		log.Printf("Failed to set local description: %v", err)
		return
	}

	pm.signalingServer.SendToPeer(peerID, signaling.Message{
		Type: "answer",
		Data: answer,
	})
}

func (pm *PeerManager) handleAnswer(roomID string, peerID string, answer webrtc.SessionDescription) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	room, exists := pm.rooms[roomID]
	if !exists {
		log.Printf("Room %s does not exist", roomID)
		return
	}

	peerConnection, exists := room.Peers[peerID]
	if !exists {
		log.Printf("Peer %s not found in room %s", peerID, roomID)
		return
	}

	err := peerConnection.SetRemoteDescription(answer)
	if err != nil {
		log.Printf("Failed to set remote description: %v", err)
		return
	}
}

func (pm *PeerManager) handleICECandidate(roomID string, peerID string, candidate webrtc.ICECandidateInit) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	room, exists := pm.rooms[roomID]
	if !exists {
		log.Printf("Room %s does not exist", roomID)
		return
	}

	peerConnection, exists := room.Peers[peerID]
	if !exists {
		log.Printf("Peer %s not found in room %s", peerID, roomID)
		return
	}

	err := peerConnection.AddICECandidate(candidate)
	if err != nil {
		log.Printf("Failed to add ICE candidate: %v", err)
		return
	}
}

func (pm *PeerManager) notifyPeersInRoom(roomID string, excludePeerID string, eventType string) {
	room, exists := pm.rooms[roomID]
	if !exists {
		return
	}

	for peerID := range room.Peers {
		if peerID != excludePeerID {
			pm.signalingServer.SendToPeer(peerID, signaling.Message{
				Type: eventType,
				Data: map[string]string{"peerID": excludePeerID},
			})
		}
	}
	}