package main

import (
	"fmt"
	"time"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"os/signal"
	"syscall"
	"encoding/binary"
	"log"
)

const MaxCPUs = 16
const MaxSessions = 250
const PlayerInputChanSize = 100000
const PlayerStateSize = 8 + 1000
const PlayerTimeout = 15
const InputSize = 8 + 8 + 8 + 100

type PlayerData struct {
	lastInputTime uint64
	sessionId     uint64
	inputChan     chan []byte
	state         []byte
}

var cpu int
var playerMap map[uint64]*PlayerData

var inputsProcessed uint64

func processInput(input []byte) {
	sessionId := binary.LittleEndian.Uint64(input[:])
	player := playerMap[sessionId]
	if player == nil {
		// fmt.Printf("player %x create\n", sessionId)
		player = &PlayerData{}
		playerMap[sessionId] = player
		player.sessionId = sessionId
		player.inputChan = make(chan []byte, PlayerInputChanSize)
		player.state = make([]byte, PlayerStateSize)
		go func() {
			for {
				input := <-player.inputChan
				if len(input) != InputSize {
					// fmt.Printf("player %x destroy\n", sessionId)
					return
				}
				player.lastInputTime = uint64(time.Now().Unix())
				t := binary.LittleEndian.Uint64(input[8:16])
				dt := binary.LittleEndian.Uint64(input[16:24])
				for i := range player.state {
					player.state[i] ^= byte(t) + byte(i)
				}
				binary.LittleEndian.PutUint64(player.state[0:8], t+dt)
				inputsProcessed++
				runtime.Gosched()			// IMPORTANT: yield back to the goroutine scheduler at the end of work. Without this the goroutine scheduler will process all inputs in player order. We want them to be distributed fairly instead.
			}
		}()
	}
	player.inputChan <- input
}

func main() {

	termChan := make(chan os.Signal, 1)

	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	if len(os.Args) == 2 {
		var err error
		cpu, err =	strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Printf("error: could not read cpu index\n")
			os.Exit(1)
		}
	}

	log.Printf("started worker on cpu %d\n", cpu)

	playerMap = make(map[uint64]*PlayerData)

	if runtime.GOOS == "linux" {
		pid := os.Getpid()
		cmd := exec.Command("taskset", "-pc", fmt.Sprintf("%d", cpu), fmt.Sprintf("%d", pid))
		if err := cmd.Run(); err != nil {
			fmt.Printf("error: could not pin process to cpu %d: %v\n", cpu, err)
			os.Exit(1)
		}	
	}

	runtime.GOMAXPROCS(1)

	ticker := time.NewTicker(time.Second)

	t := uint64(0)
	dt := uint64(1)

	go func() {
		for {
			<-ticker.C
			for i := 0; i < 100; i++ {
				for sessionId := uint64(0); sessionId < MaxSessions; sessionId++ {
					input := make([]byte, InputSize)
					binary.LittleEndian.PutUint64(input, sessionId)
					binary.LittleEndian.PutUint64(input[8:16], t)
					binary.LittleEndian.PutUint64(input[16:24], dt)
					processInput(input)
					runtime.Gosched()
				}
			}
			log.Printf("update %d: %d inputs processed\n", t, inputsProcessed)
			t += dt
	 	}		
	}()

	<- termChan
}
