package peers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v3"

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

func (pm *PeerManager) CreateRoom(roomID string) string {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

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

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		pm.signalingServer.SendToPeer(peerID, signaling.Message{
			Type: "ice-candidate",
			Data: map[string]interface{}{
				"roomName":  roomID,
				"peerID":    peerID,
				"candidate": c.ToJSON(),
			},
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
	log.Printf("roooom data 0000 isss %s", message.Type)
	log.Printf("roooom data 0000xxxx isss %s", message.Type)
	switch message.Type {
	case "create":
		roomData, ok := message.Data.(map[string]interface{})
		log.Printf("roooom data 1 isss %s", roomData)

		if !ok {
			log.Printf("Invalid data format for create message")
			return
		}
		roomID, ok := roomData["roomName"].(string)
		if !ok {
			log.Printf("Invalid room name in create message")
			return
		}
		pm.CreateRoom(roomID)

	case "join":
		joinData, ok := message.Data.(map[string]interface{})
		if !ok {
			log.Printf("Invalid data format for join message")
			return
		}
		roomID, ok := joinData["roomName"].(string)
		if !ok {
			log.Printf("Invalid room name in join message")
			return
		}
		peerID, ok := joinData["peerID"].(string)
		if !ok {
			log.Printf("Invalid peer ID in join message")
			return
		}
		err := pm.JoinRoom(roomID, peerID)
		if err != nil {
			log.Printf("Failed to join room: %v", err)
			return
		}

	case "leave":
		leaveData, ok := message.Data.(map[string]interface{})
		if !ok {
			log.Printf("Invalid data format for leave message")
			return
		}
		roomID, ok := leaveData["roomName"].(string)
		if !ok {
			log.Printf("Invalid room name in leave message")
			return
		}
		peerID, ok := leaveData["peerID"].(string)
		if !ok {
			log.Printf("Invalid peer ID in leave message")
			return
		}
		pm.LeaveRoom(roomID, peerID)

	case "offer":
		offerData, ok := message.Data.(map[string]interface{})
		if !ok {
			log.Printf("Invalid data format for offer message")
			return
		}
		roomID, ok := offerData["roomName"].(string)
		if !ok {
			log.Printf("Invalid room name in offer message")
			return
		}
		peerID, ok := offerData["peerID"].(string)
		if !ok {
			log.Printf("Invalid peer ID in offer message")
			return
		}
		// offer, ok := offerData["offer"].(webrtc.SessionDescription)
		// if !ok {
		// 	log.Printf("Invalid offer in offer message")
		// 	return
		// }
		offerMap, ok := offerData["offer"].(map[string]interface{})
		if !ok {
			log.Printf("Invalid offer format in offer message")
			return
		}

		offerJSON, err := json.Marshal(offerMap) // Convert map to JSON
		if err != nil {
			log.Printf("Failed to marshal offer data: %v", err)
			return
		}

		var offer webrtc.SessionDescription
		if err := json.Unmarshal(offerJSON, &offer); err != nil {
			log.Printf("Failed to unmarshal offer to webrtc.SessionDescription: %v", err)
			return
		}
		pm.handleOffer(roomID, peerID, offer)

	case "answer":
		answerData, ok := message.Data.(map[string]interface{})
		if !ok {
			log.Printf("Invalid data format for answer message")
			return
		}
		roomID, ok := answerData["roomName"].(string)
		if !ok {
			log.Printf("Invalid room name in answer message")
			return
		}
		peerID, ok := answerData["peerID"].(string)
		if !ok {
			log.Printf("Invalid peer ID in answer message")
			return
		}
		answer, ok := answerData["answer"].(webrtc.SessionDescription)
		if !ok {
			log.Printf("Invalid answer in answer message")
			return
		}
		pm.handleAnswer(roomID, peerID, answer)

	case "ice-candidate":
		candidateData, ok := message.Data.(map[string]interface{})
		if !ok {
			log.Printf("Invalid data format for ice-candidate message")
			return
		}
		roomID, ok := candidateData["roomName"].(string)
		if !ok {
			log.Printf("Invalid room name in ice-candidate message ox %s", roomID)
			return
		}
		peerID, ok := candidateData["peerID"].(string)
		if !ok {
			log.Printf("Invalid peer ID in ice-candidate message")
			return
		}
		candidateInitMap, ok := candidateData["candidate"].(map[string]interface{})
		if !ok {
			log.Printf("Invalid candidate in ice-candidate message")
			return
		}
		ICECandidateJSON, err := json.Marshal(candidateInitMap) // Convert map to JSON
		if err != nil {
			log.Printf("Failed to marshal offer data: %v", err)
			return
		}
		var ICECandidate webrtc.ICECandidateInit
		if err := json.Unmarshal(ICECandidateJSON, &ICECandidate); err != nil {
			log.Printf("Failed to unmarshal offer to webrtc.ICECandidate: %v", err)
			return
		}
		pm.handleICECandidate(roomID, peerID, ICECandidate)
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
		Data: map[string]interface{}{
			"roomName": roomID,
			"peerID":   peerID,
			"answer":   answer,
		},
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
