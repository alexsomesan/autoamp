package main

import (
	"log"
	"os"
	"strings"
	"time"
)

type PlayEvent int
type PlayState int
type PlayChannel chan PlayEvent

const (
	DefaultAmpControlDev = "/dev/ampcontrol"
	DefaultALSAStatusDev = "/proc/asound/card1/pcm0p/sub0/status"
	DefaultTickMs        = 333

	EventPlaying PlayEvent = iota
	EventClosed

	StateStopped PlayState = iota
	StatePlaying
)

func main() {
	var state = StateStopped

	ampDev, err := os.OpenFile(DefaultAmpControlDev, os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open AMP control device: ", err)
	}
	log.Println("Amp control device opened")
	defer func() {
		cerr := ampDev.Close()
		if cerr != nil {
			log.Fatal("Failed to close AMP control device: ", err)
		}
		log.Println("Amp control device closed")
	}()

	log.Println("AMP control running...")

	evch := make(PlayChannel)

	for {
		select {
		case <-time.Tick(DefaultTickMs * time.Millisecond):
			go readPlayState(evch)
		case event := <-evch:
			switch event {
			case EventPlaying:
				state = startAmp(state, ampDev)
			case EventClosed:
				state = stopAmp(state, ampDev)
			}
		}
	}
}

func readPlayState(ch PlayChannel) {
	as, err := os.ReadFile(DefaultALSAStatusDev)
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(string(as), "closed") {
		ch <- EventClosed
	}
	if strings.Contains(string(as), "state: RUNNING") {
		ch <- EventPlaying
	}
}

func stopAmp(s PlayState, f *os.File) PlayState {
	if s != StatePlaying {
		return s
	}
	i, err := f.Write([]byte{'0'})
	if err != nil {
		log.Fatal(err)
	}
	if i != 1 {
		log.Println("Failed to write 0 to control device")
		return s
	}
	log.Println("Stopped")
	return StateStopped
}

func startAmp(s PlayState, f *os.File) PlayState {
	if s != StateStopped {
		return s
	}
	i, err := f.Write([]byte{'1'})
	if err != nil {
		log.Fatal(err)
	}
	if i != 1 {
		log.Println("Failed to write 1 to control device")
		return s
	}
	log.Println("Playing")
	return StatePlaying
}
