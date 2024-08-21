 package storage

// import (
// 	"context"
// 	"fmt"
// 	"io"

// 	"cloud.google.com/go/storage"
// 	"github.com/pion/webrtc/v3"
// 	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
// 	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
// )

// type CloudRecorder struct {
// 	bucket      *storage.BucketHandle
// 	videoWriter *ivfwriter.IVFWriter
// 	audioWriter *oggwriter.OggWriter
// 	videoObject *storage.Writer
// 	audioObject *storage.Writer
// }

// func NewCloudRecorder(bucketName, filename string) (*CloudRecorder, error) {
// 	ctx := context.Background()
// 	client, err := storage.NewClient(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create storage client: %w", err)
// 	}

// 	bucket := client.Bucket(bucketName)

// 	videoObject := bucket.Object(fmt.Sprintf("%s.ivf", filename)).NewWriter(ctx)
// 	audioObject := bucket.Object(fmt.Sprintf("%s.ogg", filename)).NewWriter(ctx)

// 	videoWriter, err := ivfwriter.New(videoObject)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create IVF writer: %w", err)
// 	}

// 	audioWriter, err := oggwriter.New(audioObject, 48000, 2)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create Ogg writer: %w", err)
// 	}

// 	return &CloudRecorder{
// 		bucket:      bucket,
// 		videoWriter: videoWriter,
// 		audioWriter: audioWriter,
// 		videoObject: videoObject,
// 		audioObject: audioObject,
// 	}, nil
// }

// func (cr *CloudRecorder) AddTrack(track *webrtc.TrackRemote) {
// 	go func() {
// 		for {
// 			rtpPacket, _, err := track.ReadRTP()
// 			if err != nil {
// 				if err == io.EOF {
// 					return
// 				}
// 				fmt.Printf("Failed to read RTP packet: %v\n", err)
// 				continue
// 			}

// 			if track.Kind() == webrtc.RTPCodecTypeVideo {
// 				cr.videoWriter.WriteRTP(rtpPacket)
// 			} else if track.Kind() == webrtc.RTPCodecTypeAudio {
// 				cr.audioWriter.WriteRTP(rtpPacket)
// 			}
// 		}
// 	}()
// }

// func (cr *CloudRecorder) Stop() {
// 	if cr.videoWriter != nil {
// 		cr.videoWriter.Close()
// 	}
// 	if cr.audioWriter != nil {
// 		cr.audioWriter.Close()
// 	}
// 	cr.videoObject.Close()
// 	cr.audioObject.Close()
// }
