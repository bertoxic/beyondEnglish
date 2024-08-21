 package peers

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"sync"

// 	"github.com/bertoxic/beyondEnglish/media"
// 	"github.com/bertoxic/beyondEnglish/signaling"
// 	"github.com/bertoxic/beyondEnglish/storage"
// 	"github.com/pion/webrtc/v3"
// )

// type PeerManager struct {
// 	signalingClient *signaling.Client
// 	peerConnections map[string]*PeerConnection
// 	localVideoTrack *media.VideoTrack
// 	localAudioTrack *media.AudioTrack
// 	recorder        storage.FileRecorder
// 	roomName        string
// 	mutex           sync.Mutex
// }

// func (pm *PeerManager) CreateRoom(roomName string) error {
// 	err := pm.signalingClient.CreateRoom(roomName)
// 	if err != nil {
// 		return fmt.Errorf("failed to create room: %w", err)
// 	}
// 	pm.roomName = roomName
// 	return nil
// }

// func (pm *PeerManager) JoinRoom(roomName string) error {
// 	err := pm.signalingClient.JoinRoom(roomName)
// 	if err != nil {
// 		return fmt.Errorf("failed to join room: %w", err)
// 	}
// 	pm.roomName = roomName
// 	return nil
// }

// // func (pm *PeerManager) JoinRoom(roomName string) error {
// //     // Create a join room message
// //     joinMessage := signaling.JoinRoomMessage{
// //         Type: "join",
// //         Room: roomName,
// //     }

// //     // Convert the message to JSON
// //     messageJSON, err := json.Marshal(joinMessage)
// //     if err != nil {
// //         return fmt.Errorf("failed to marshal join message: %w", err)
// //     }

// //     // Send the message using the existing SendMessage method
// //    err =
// //      pm.signalingClient.SendtoJoin(messageJSON)
// //     if err != nil {
// //         return fmt.Errorf("failed to join room: %w", err)
// //     }

// //     log.Printf("Joined room: %s", roomName)
// //     return nil
// // }

// func (pm *PeerManager) LeaveRoom() {
// 	pm.mutex.Lock()
// 	defer pm.mutex.Unlock()

// 	if pm.roomName != "" {
// 		pm.signalingClient.LeaveRoom(pm.roomName)
// 		for _, pc := range pm.peerConnections {
// 			pc.pc.Close()
// 		}
// 		pm.peerConnections = make(map[string]*PeerConnection)
// 		pm.roomName = ""
// 	}
// }

// // func (pm *PeerManager) LeaveRoom() {
// // 	pm.mutex.Lock()
// // 	defer pm.mutex.Unlock()

// // 	for _, pc := range pm.peerConnections {
// // 		pc.pc.Close()
// // 	}
// // 	pm.peerConnections = make(map[string]*PeerConnection)
// // }

// func (pm *PeerManager) handleSignalingMessage(message []byte) {
// 	var msg map[string]interface{}
// 	if err := json.Unmarshal(message, &msg); err != nil {
// 		log.Printf("Error unmarshaling message: %v", err)
// 		return
// 	}

// 	switch msg["type"] {
// 	case "offer":
// 		pm.handleOffer(msg["from"].(string), msg["offer"].(webrtc.SessionDescription))
// 	case "answer":
// 		pm.handleAnswer(msg["from"].(string), msg["answer"].(webrtc.SessionDescription))
// 	case "ice-candidate":
// 		pm.handleICECandidate(msg["from"].(string), msg["candidate"].(webrtc.ICECandidateInit))
// 	case "participant_joined":
// 		pm.handleParticipantJoined(msg["from"].(string))
// 	case "participant_left":
// 		pm.handleParticipantLeft(msg["from"].(string))
// 	}
// }

// // func (pm *PeerManager) handleSignalingMessage(message []byte) {
// // 	var msg map[string]interface{}
// // 	err := json.Unmarshal(message, &msg)
// // 	if err != nil {
// // 		log.Printf("Failed to parse signaling message: %v", err)
// // 		return
// // 	}

// // 	messageType, ok := msg["type"].(string)
// // 	if !ok {
// // 		log.Printf("Message type not found or not a string")
// // 		return
// // 	}

// // 	switch messageType {
// // 	case "offer":
// // 		offer := webrtc.SessionDescription{
// // 			Type: webrtc.SDPTypeOffer,
// // 			SDP:  msg["sdp"].(string),
// // 		}
// // 		peerID := msg["from"].(string)
// // 		err = pm.handleOffer(peerID, offer)
// // 		if err != nil {
// // 			log.Printf("Failed to handle offer: %v", err)
// // 		}
// // 	case "answer":
// // 		answer := webrtc.SessionDescription{
// // 			Type: webrtc.SDPTypeAnswer,
// // 			SDP:  msg["sdp"].(string),
// // 		}
// // 		peerID := msg["from"].(string)
// // 		err = pm.handleAnswer(peerID, answer)
// // 		if err != nil {
// // 			log.Printf("Failed to handle answer: %v", err)
// // 		}
// // 	case "ice-candidate":
// // 		candidate := webrtc.ICECandidateInit{
// // 			Candidate: msg["candidate"].(string),
// // 		}
// // 		peerID := msg["from"].(string)
// // 		err = pm.handleICECandidate(peerID, candidate)
// // 		if err != nil {
// // 			log.Printf("Failed to handle ICE candidate: %v", err)
// // 		}
// // 	default:
// // 		log.Printf("Unknown message type: %s", messageType)
// // 	}
// // }

// func (pm *PeerManager) handleParticipantJoined(peerID string) {
// 	// Create a new peer connection for the joined participant
// 	_, err := pm.getOrCreatePeerConnection(peerID)
// 	if err != nil {
// 		log.Printf("Error creating peer connection for new participant: %v", err)
// 		return
// 	}

// 	// Create and send an offer to the new participant
// 	offer, err := pm.createOffer(peerID)
// 	if err != nil {
// 		log.Printf("Error creating offer for new participant: %v", err)
// 		return
// 	}

// 	err = pm.signalingClient.SendOffer(pm.roomName, peerID, offer)
// 	if err != nil {
// 		log.Printf("Error sending offer to new participant: %v", err)
// 	}
// }

// func (pm *PeerManager) handleParticipantLeft(peerID string) {
// 	pm.mutex.Lock()
// 	defer pm.mutex.Unlock()

// 	if pc, ok := pm.peerConnections[peerID]; ok {
// 		pc.pc.Close()
// 		delete(pm.peerConnections, peerID)
// 	}
// }

// func (pm *PeerManager) createOffer(peerID string) (webrtc.SessionDescription, error) {
// 	pc, err := pm.getOrCreatePeerConnection(peerID)
// 	if err != nil {
// 		return webrtc.SessionDescription{}, err
// 	}

// 	offer, err := pc.CreateOffer()
// 	if err != nil {
// 		return webrtc.SessionDescription{}, fmt.Errorf("failed to create offer: %w", err)
// 	}

// 	err = pc.SetLocalDescription(offer)
// 	if err != nil {
// 		return webrtc.SessionDescription{}, fmt.Errorf("failed to set local description: %w", err)
// 	}

// 	return *offer, nil
// }

// func NewPeerManager(signalingClient *signaling.Client) *PeerManager {
// 	pm := &PeerManager{
// 		signalingClient: signalingClient,
// 		peerConnections: make(map[string]*PeerConnection),
// 	}

// 	signalingClient.SetOnMessage(pm.handleSignalingMessage)
// 	return pm
// }

// func (pm *PeerManager) SetupLocalMedia() error {
// 	var err error
// 	pm.localVideoTrack, err = media.NewVideoTrack()
// 	if err != nil {
// 		return fmt.Errorf("failed to create video track: %w", err)
// 	}

// 	pm.localAudioTrack, err = media.NewAudioTrack()
// 	if err != nil {
// 		return fmt.Errorf("failed to create audio track: %w", err)
// 	}

// 	return nil
// }

// func (pm *PeerManager) SetRecorder(recorder storage.FileRecorder) {
// 	pm.recorder = recorder
// }

// func (pm *PeerManager) handleOffer(peerID string, offer webrtc.SessionDescription) error {
// 	pc, err := pm.getOrCreatePeerConnection(peerID)
// 	if err != nil {
// 		return err
// 	}

// 	err = pc.SetRemoteDescription(offer)
// 	if err != nil {
// 		return fmt.Errorf("failed to set remote description: %w", err)
// 	}

// 	answer, err := pc.CreateAnswer()
// 	if err != nil {
// 		return fmt.Errorf("failed to create answer: %w", err)
// 	}

// 	err = pc.SetLocalDescription(answer)
// 	if err != nil {
// 		return fmt.Errorf("failed to set local description: %w", err)
// 	}

// 	return pm.signalingClient.SendAnswer(peerID, answer)
// }

// func (pm *PeerManager) handleAnswer(peerID string, answer webrtc.SessionDescription) error {
// 	pc, err := pm.getOrCreatePeerConnection(peerID)
// 	if err != nil {
// 		return err
// 	}

// 	return pc.SetRemoteDescription(answer)
// }

// func (pm *PeerManager) handleICECandidate(peerID string, candidate webrtc.ICECandidateInit) error {
// 	pc, err := pm.getOrCreatePeerConnection(peerID)
// 	if err != nil {
// 		return err
// 	}

// 	return pc.AddICECandidate(candidate)
// }

// func (pm *PeerManager) getOrCreatePeerConnection(peerID string) (*PeerConnection, error) {
// 	pm.mutex.Lock()
// 	defer pm.mutex.Unlock()

// 	if pc, ok := pm.peerConnections[peerID]; ok {
// 		return pc, nil
// 	}

// 	config := webrtc.Configuration{
// 		ICEServers: []webrtc.ICEServer{
// 			{
// 				URLs: []string{"stun:stun.l.google.com:19302"},
// 			},
// 		},
// 	}

// 	pc, err := NewPeerConnection(config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
// 		if c == nil {
// 			return
// 		}
// 		pm.signalingClient.SendICECandidate(peerID, c.ToJSON())
// 	})

// 	pc.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
// 		log.Printf("Received remote track: %s", remoteTrack.ID())
// 		if &pm.recorder != nil {
// 			pm.recorder.AddTrack(remoteTrack)
// 		}
// 	})

// 	if pm.localVideoTrack != nil {
// 		_, err = pc.AddTrack(pm.localVideoTrack.TrackLocal())
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to add local video track: %w", err)
// 		}
// 	}

// 	if pm.localAudioTrack != nil {
// 		_, err = pc.AddTrack(pm.localAudioTrack.TrackLocal())
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to add local audio track: %w", err)
// 		}
// 	}

// 	pm.peerConnections[peerID] = pc
// 	return pc, nil
// }
