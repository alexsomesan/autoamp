package main

import (
	"log"
	"os"
	"strings"
	"sync"
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
	var wg sync.WaitGroup

	ampDev, err := os.OpenFile(DefaultAmpControlDev, os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	log.Println("Amp control device opened")
	defer func() {
		ampDev.Close()
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
	wg.Wait()
}

func readPlayState(ch PlayChannel) {
	as, err := os.ReadFile(DefaultALSAStatusDev)
	if err != nil {
		panic(err)
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
		panic(err)
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
		panic(err)
	}
	if i != 1 {
		log.Println("failed to write 1 to device")
		return
	}
	log.Println("playing")
}
