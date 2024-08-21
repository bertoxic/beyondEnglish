package media

import (
	"fmt"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type VideoTrack struct {
	track *webrtc.TrackLocalStaticSample
}

func NewVideoTrack() (*VideoTrack, error) {
	track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "pion")
	if err != nil {
		return nil, fmt.Errorf("failed to create video track: %w", err)
	}

	return &VideoTrack{track: track}, nil
}

func (vt *VideoTrack) TrackLocal() *webrtc.TrackLocalStaticSample {
	return vt.track
}

func (vt *VideoTrack) WriteSample(sample media.Sample) error {
	return vt.track.WriteSample(sample)
}

func (vt *VideoTrack) StartDummyVideo() {
	go func() {
		for {
			// Create a dummy video frame (you'd replace this with actual video capture)
			dummyFrame := []byte{0x00, 0x00, 0x00} // Replace with actual video data
			err := vt.WriteSample(media.Sample{
				Data:     dummyFrame,
				Duration: time.Second / 30, // 30 fps
			})
			if err != nil {
				fmt.Printf("Failed to write video sample: %v\n", err)
			}
			time.Sleep(time.Second / 30)
		}
	}()
}