package http

import (
	"github.com/labstack/echo/v4"
	handler "github.com/sandrospengler/streamserver/pkg/http/handler/home"
	webrtc "github.com/sandrospengler/streamserver/pkg/http/handler/webrtc"
	log "github.com/sirupsen/logrus"
)

func StartHTTPServer() {

	port := ":8080"
	log.Info("HTTP-Server starting on port " + port)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	peerConnectionHandler := webrtc.PeerConnectionHandler{}

	homeHandler := handler.HomeHandler{}

	e.GET("/", homeHandler.HandleHomeShow)

	e.POST("createPeerConnection", peerConnectionHandler.HandleCreatePeerConnection)

	e.Static("/assets", "assets")

	err := e.Start(port)

	if err != nil {
		log.Fatal("Could not start http-server", err)
	}
}
