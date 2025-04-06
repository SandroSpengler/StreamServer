package broadcast

import (
	"sync"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

type Broadcaster struct {
	Tracks []*webrtc.TrackLocalStaticSample
	lock   sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		Tracks: make([]*webrtc.TrackLocalStaticSample, 0),
	}
}

func (b *Broadcaster) AddTrack(track *webrtc.TrackLocalStaticSample) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.Tracks = append(b.Tracks, track)
}

func (b *Broadcaster) WriteSample(sample media.Sample) error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	var err error
	for _, track := range b.Tracks {
		if writeErr := track.WriteSample(sample); writeErr != nil {
			err = writeErr // save the last error if any
		}
	}
	return err
}
