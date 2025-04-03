package handler

import (
	"encoding/json"
	"net/http"

	rtmp "github.com/sandrospengler/streamserver/pkg/rtmp"
	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
	"github.com/pion/webrtc/v3"
)

type PeerConnectionHandler struct{}

func (h PeerConnectionHandler) HandleCreatePeerConnection(c echo.Context) error {

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Error("Error creating PeerConnection", err)
		return err
	}

	if rtmp.VIDEO_TRACK == nil {
		rtmp.VIDEO_TRACK, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264,
		}, "video", "stream-0")
		if err != nil {
			return err
		}

		if rtmp.VIDEO_TRACK == nil {
			log.Println("VideoTrack did not initialize")
		}

		log.WithFields(log.Fields{
			"StreamID": rtmp.VIDEO_TRACK.ID(),
			"Kind":     rtmp.VIDEO_TRACK.Kind(),
			"Codec":    rtmp.VIDEO_TRACK.Codec(),
		}).Info("Created VideoTrack")

		_, err = peerConnection.AddTrack(rtmp.VIDEO_TRACK)
		if err != nil {
			log.Error("Error adding VideoTrack to PeerConnection", err)
			return err
		}
	}

	if rtmp.AUDIO_TRACK == nil {
		rtmp.AUDIO_TRACK, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeOpus,
		}, "audio", "stream-0")
		if err != nil {
			return err
		}

		if rtmp.AUDIO_TRACK == nil {
			log.Println("AudioTrack did not initialize")
		}

		log.Info("Created AudioTrack")

		_, err = peerConnection.AddTrack(rtmp.AUDIO_TRACK)
		if err != nil {
			log.Error("Error adding AudioTrack to PeerConnection", err)
			return err
		}
	}

	var offer webrtc.SessionDescription
	err = json.NewDecoder(c.Request().Body).Decode(&offer)
	if err != nil {
		log.Error("Error decoding offer", err)
		return err
	}

	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.Error("Error setting remote description", err)
		return err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Error("Error creating answer", err)
		return err
	}

	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Error("Error setting local description", err)
		return err
	}
	<-gatherComplete

	response, err := json.Marshal(peerConnection.LocalDescription())
	if err != nil {
		log.Error("Error marshaling answer", err)
		return err
	}

	c.Response().Header().Set("Content-Type", "application/json")
	return c.JSONBlob(http.StatusOK, response)
}
