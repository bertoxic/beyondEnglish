package media

import (
	"fmt"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type AudioTrack struct {
	track *webrtc.TrackLocalStaticSample
}

func NewAudioTrack() (*AudioTrack, error) {
	track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if err != nil {
		return nil, fmt.Errorf("failed to create audio track: %w", err)
	}

	return &AudioTrack{track: track}, nil
}

func (at *AudioTrack) TrackLocal() *webrtc.TrackLocalStaticSample {
	return at.track
}

func (at *AudioTrack) WriteSample(sample media.Sample) error {
	return at.track.WriteSample(sample)
}

func (at *AudioTrack) StartDummyAudio() {
	go func() {
		for {
			// Create a dummy audio sample (you'd replace this with actual audio capture)
			dummySample := []byte{0x00, 0x00, 0x00} // Replace with actual audio data
			err := at.WriteSample(media.Sample{
				Data:     dummySample,
				Duration: time.Millisecond * 20, // 20ms audio frames
			})
			if err != nil {
				fmt.Printf("Failed to write audio sample: %v\n", err)
			}
			time.Sleep(time.Millisecond * 20)
		}
	}()
}