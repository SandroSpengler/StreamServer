package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pion/webrtc/v4"
	rtmp "github.com/sandrospengler/streamserver/pkg/rtmp"
	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
)

type PeerConnectionHandler struct{}

func (h PeerConnectionHandler) HandleCreatePeerConnection(c echo.Context) error {

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Error("Error creating PeerConnection", err)
		return err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(rtmp.VIDEO_CODEC, "video", "stream-0")
	if err != nil {
		return err
	}

	rtmp.VideoBroadcaster.AddTrack(videoTrack)

	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		log.Error("Error adding VideoTrack to PeerConnection", err)
		return err
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(rtmp.AUDIO_CODEC, "audio", "stream-0")
	if err != nil {
		return err
	}
	rtmp.AudioBroadcaster.AddTrack(audioTrack)

	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		log.Error("Error adding AudioTrack to PeerConnection", err)
		return err
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
