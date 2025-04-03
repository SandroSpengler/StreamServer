package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/sandrospengler/streamserver/pkg/http"
	"github.com/sandrospengler/streamserver/pkg/rtmp"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	log.Info("Starting StreamServer")

	go rtmp.StartRTMPServer()

	http.StartHTTPServer()
}
