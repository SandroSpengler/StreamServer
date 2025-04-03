package rtmp

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"
	"time"

	"github.com/Glimesh/go-fdkaac/fdkaac"
	"github.com/hraban/opus"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	log "github.com/sirupsen/logrus"

	flvtag "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"
	gortmp "github.com/yutopp/go-rtmp"
	rtmpmsg "github.com/yutopp/go-rtmp/message"
)

var VIDEO_TRACK *webrtc.TrackLocalStaticSample = nil
var AUDIO_TRACK *webrtc.TrackLocalStaticSample = nil

type RTMPHandler struct {
	gortmp.DefaultHandler
	peerConnection *webrtc.PeerConnection
	audioDecoder   *fdkaac.AacDecoder
	audioEncoder   *opus.Encoder
}

func StartRTMPServer() {
	port := ":1935"
	log.Info("RTMP-Server starting on port " + port)

	tcpAddr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		log.Fatal("Could not resolve tcp address", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal("Could not listen on tcp address", err)
	}

	server := gortmp.NewServer(&gortmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			return conn, &gortmp.ConnConfig{
				Handler: &RTMPHandler{
					peerConnection: nil,
					audioEncoder:   newOpusEncoder(),
					audioDecoder:   fdkaac.NewAacDecoder(),
				},

				ControlState: rtmp.StreamControlStateConfig{
					DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
				},
			}
		},
	})
	if err := server.Serve(listener); err != nil {
		log.Fatal("Could not serve", err)
	}
}

func newOpusEncoder() *opus.Encoder {
	encoder, err := opus.NewEncoder(48000, 2, opus.AppAudio)
	if err != nil {
		log.Panicf("Failed to create Opus encoder: %v", err)
	}
	encoder.SetMaxBandwidth(opus.Fullband)
	encoder.SetComplexity(10)
	encoder.SetBitrate(128000)
	// encoder.SetBitrateToMax()
	encoder.SetInBandFEC(true)
	return encoder
}

func (h *RTMPHandler) OnAudio(timestamp uint32, payload io.Reader) error {
	start := time.Now()

	var audio flvtag.AudioData
	if err := flvtag.DecodeAudioData(payload, &audio); err != nil {
		return err
	}

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, audio.Data); err != nil {
		return err
	}

	if data.Len() <= 0 {
		log.Println("no audio data", timestamp, payload)
		return nil
	}

	datas := data.Bytes()

	var opusBuffer []byte = make([]byte, 4000)

	if audio.AACPacketType == flvtag.AACPacketTypeSequenceHeader {
		log.Println("Created new codec ", hex.EncodeToString(datas))
		if err := h.audioDecoder.InitRaw(datas); err != nil {
			log.Println("error initializing stream", err)
		}
		return nil
	}

	pcm, err := h.audioDecoder.Decode(datas)
	if err != nil {
		log.Println("decode error: ", hex.EncodeToString(datas), err)
		return nil
	}

	pcmInt16 := bytesToInt16(pcm)

	// Resample from 1024 to 960 samples
	resampledPCM := resample(pcmInt16, 1024, 960)

	// Debugging: Log PCM length and frame size
	frameSize := 960 * 2 // 960 samples * 2 channels
	log.Printf("Resampled PCM length: %d, Opus frame size: %d", len(resampledPCM), frameSize)

	// Ensure PCM length is a multiple of frameSize
	if len(resampledPCM)%frameSize != 0 {
		log.Printf("Warning: Trimming PCM size from %d to %d\n", len(resampledPCM), (len(resampledPCM)/frameSize)*frameSize)
		resampledPCM = resampledPCM[:(len(resampledPCM)/frameSize)*frameSize]
	}

	encodedBytes, err := h.audioEncoder.Encode(resampledPCM, opusBuffer)
	if err != nil {
		log.Println("Opus encoding error:", err)
		return err
	}

	elapsed := time.Since(start).Microseconds()
	log.Printf("FFmpeg audio conversion took %d µs", elapsed)

	return AUDIO_TRACK.WriteSample(media.Sample{
		Data:     opusBuffer[:encodedBytes],
		Duration: 20 * time.Millisecond,
	})
}

func resample(in []int16, inRate int, outRate int) []int16 {
	if inRate == outRate {
		return in
	}

	outLen := int(float64(len(in)) * float64(outRate) / float64(inRate))
	out := make([]int16, outLen)

	for i := 0; i < outLen; i++ {
		inIndex := float64(i) * float64(inRate) / float64(outRate)
		inIndexInt := int(inIndex)

		if inIndexInt >= len(in)-1 {
			out[i] = in[len(in)-1]
		} else {
			inFrac := inIndex - float64(inIndexInt)
			out[i] = int16(float64(in[inIndexInt])*(1.0-inFrac) + float64(in[inIndexInt+1])*inFrac)
		}
	}
	return out
}

func bytesToInt16(pcm []byte) []int16 {
	// Trim PCM data before converting to int16
	if len(pcm)%4 != 0 {
		pcm = pcm[:len(pcm)-(len(pcm)%4)]
	}

	int16Data := make([]int16, len(pcm)/2)
	for i := 0; i < len(int16Data); i++ {
		int16Data[i] = int16(pcm[2*i]) | int16(pcm[2*i+1])<<8
	}
	return int16Data
}

const headerLengthField = 4

func (h *RTMPHandler) OnVideo(timestamp uint32, payload io.Reader) error {
	if VIDEO_TRACK == nil {
		// Because there is no WebRTC video track, we ignore the video data
		return nil
	}

	var video flvtag.VideoData
	if err := flvtag.DecodeVideoData(payload, &video); err != nil {
		return err
	}

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, video.Data); err != nil {
		return err
	}

	outBuf := []byte{}
	videoBuffer := data.Bytes()
	for offset := 0; offset < len(videoBuffer); {
		bufferLength := int(binary.BigEndian.Uint32(videoBuffer[offset : offset+headerLengthField]))
		if offset+bufferLength >= len(videoBuffer) {
			break
		}

		offset += headerLengthField
		outBuf = append(outBuf, []byte{0x00, 0x00, 0x00, 0x01}...)
		outBuf = append(outBuf, videoBuffer[offset:offset+bufferLength]...)

		offset += int(bufferLength)
	}

	//log.Info("Video data received")

	return VIDEO_TRACK.WriteSample(media.Sample{
		Data:     outBuf,
		Duration: time.Second / 30,
	})
}

func (h *RTMPHandler) OnConnect(timestamp uint32, cmd *rtmpmsg.NetConnectionConnect) error {
	log.Printf("RTMP OnConnect: %#v", cmd)
	return nil
}

func (h *RTMPHandler) OnCreateStream(timestamp uint32, cmd *rtmpmsg.NetConnectionCreateStream) error {
	log.Printf("RTMP OnCreateStream: %#v", cmd)
	return nil
}

func (h *RTMPHandler) OnPublish(_ *rtmp.StreamContext, timestamp uint32, cmd *rtmpmsg.NetStreamPublish) error {
	log.Printf("RTMP OnPublish: %#v", cmd)
	return nil
}

func (h *RTMPHandler) OnClose() {
	log.Printf("OnClose")
}
