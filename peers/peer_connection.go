package peers

import (
	"fmt"

	"github.com/pion/webrtc/v3"
)

type PeerConnection struct {
	pc *webrtc.PeerConnection
}

func NewPeerConnection(configuration webrtc.Configuration) (*PeerConnection, error) {
	pc, err := webrtc.NewPeerConnection(configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	return &PeerConnection{pc: pc}, nil
}
func (p *PeerConnection) Close() error {
	if p.pc == nil {
		return fmt.Errorf("peer connection is nil")
	}

	err := p.pc.Close()
	if err != nil {
		return fmt.Errorf("error closing peer connection: %w", err)
	}

	// If you have other resources to clean up, do it here.

	return nil
}
func (p *PeerConnection) AddTrack(track *webrtc.TrackLocalStaticSample) (*webrtc.RTPSender, error) {
	return p.pc.AddTrack(track)
}

func (p *PeerConnection) OnTrack(f func(*webrtc.TrackRemote, *webrtc.RTPReceiver)) {
	p.pc.OnTrack(f)
}

func (p *PeerConnection) OnICECandidate(f func(*webrtc.ICECandidate)) {
	p.pc.OnICECandidate(f)
}

func (p *PeerConnection) OnConnectionStateChange(f func(webrtc.PeerConnectionState)) {
	p.pc.OnConnectionStateChange(f)
}

func (p *PeerConnection) CreateOffer() (*webrtc.SessionDescription, error) {
	offer, err := p.pc.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating offer: %w", err)
	}
	return &offer, nil
}

func (p *PeerConnection) CreateAnswer(options *webrtc.AnswerOptions) (*webrtc.SessionDescription, error) {
	answer, err := p.pc.CreateAnswer(options)
	if err != nil {
		return nil, fmt.Errorf("error creating answer: %w", err)
	}
	return &answer, nil
}

func (p *PeerConnection) SetLocalDescription(desc webrtc.SessionDescription) error {
	return p.pc.SetLocalDescription(desc)
}

func (p *PeerConnection) SetRemoteDescription(desc webrtc.SessionDescription) error {
	return p.pc.SetRemoteDescription(desc)
}

func (p *PeerConnection) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	return p.pc.AddICECandidate(candidate)
}
