
package storage

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

type FileRecorder struct {
	videoWriter *ivfwriter.IVFWriter
	audioWriter *oggwriter.OggWriter
	mutex       sync.Mutex
}

func NewFileRecorder(filename string) (*FileRecorder, error) {
	videoFileName := fmt.Sprintf("%s.ivf", filename)
	audioFileName := fmt.Sprintf("%s.ogg", filename)
	videoFile, err := os.Create(fmt.Sprintf("%s.ivf", filename))
	if err != nil {
		return nil, fmt.Errorf("failed to create video file: %w", err)
	}

	audioFile, err := os.Create(fmt.Sprintf("%s.ogg", filename))
	if err != nil {
		videoFile.Close()
		return nil, fmt.Errorf("failed to create audio file: %w", err)
	}

	videoWriter, err := ivfwriter.New(videoFileName)
	if err != nil {
		videoFile.Close()
		audioFile.Close()
		return nil, fmt.Errorf("failed to create IVF writer: %w", err)
	}

	audioWriter, err := oggwriter.New(audioFileName, 48000, 2)
	if err != nil {
		videoWriter.Close()
		videoFile.Close()
		audioFile.Close()
		return nil, fmt.Errorf("failed to create Ogg writer: %w", err)
	}

	return &FileRecorder{
		videoWriter: videoWriter,
		audioWriter: audioWriter,
	}, nil
}

func (fr *FileRecorder) AddTrack(track *webrtc.TrackRemote) {
	go func() {
		for {
			rtpPacket, _, err := track.ReadRTP()
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Printf("Failed to read RTP packet: %v\n", err)
				continue
			}

			fr.mutex.Lock()
			if track.Kind() == webrtc.RTPCodecTypeVideo {
				fr.videoWriter.WriteRTP(rtpPacket)
			} else if track.Kind() == webrtc.RTPCodecTypeAudio {
				fr.audioWriter.WriteRTP(rtpPacket)
			}
			fr.mutex.Unlock()
		}
	}()
}

func (fr *FileRecorder) Stop(){
	fr.mutex.Lock()
	defer fr.mutex.Unlock()

	if fr.videoWriter != nil {
		fr.videoWriter.Close()
	}
	if fr.audioWriter != nil {	
		fr.audioWriter.Close()
	}
}