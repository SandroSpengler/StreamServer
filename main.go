package main

import (
	"github.com/sandrospengler/streamserver/pkg/rtmp"
	"log"
)

func main() {
	log.Println("starting server")

	rtmp.StartRTMPServer()
}
