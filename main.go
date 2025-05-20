package main

import (
	"log"
	"os"
	"strings"
	"time"
)

type PlayEvent int
type PlayChannel chan PlayEvent

const (
	DefaultAmpControlDev = "/dev/ampcontrol"
	DefaultALSAStatusDev = "/proc/asound/card1/pcm0p/sub0/status"
	DefaultTickMs        = 300

	EventPlaying PlayEvent = iota
	EventClosed
)

func main() {
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
				startAmp(ampDev)
			case EventClosed:
				stopAmp(ampDev)
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

func stopAmp(f *os.File) {
	i, err := f.Write([]byte{'0'})
	if err != nil {
		log.Fatal(err)
	}
	if i != 1 {
		log.Println("failed to write 0 to device")
		return
	}
	log.Println("stopped")
}

func startAmp(f *os.File) {
	i, err := f.Write([]byte{'1'})
	if err != nil {
		log.Fatal(err)
	}
	if i != 1 {
		log.Println("failed to write 1 to device")
		return
	}
	log.Println("playing")
}
