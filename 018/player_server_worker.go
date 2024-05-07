package main

import (
	"fmt"
	"time"
	"os"
	"runtime"
	"strconv"
	"os/signal"
	"syscall"
	"encoding/binary"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
)

const MaxCPUs = 32
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
var playerStateMap *ebpf.Map
var inputsProcessed uint64
var inputsProcessedMap *ebpf.Map

func processInput(input []byte) {
	sessionId := binary.LittleEndian.Uint64(input[:])
	player := playerMap[sessionId]
	if player == nil {
		fmt.Printf("player %x create\n", sessionId)
		player = &PlayerData{}
		playerMap[sessionId] = player
		player.sessionId = sessionId
		player.inputChan = make(chan []byte, PlayerInputChanSize)
		player.state = make([]byte, PlayerStateSize)
		go func() {
			for {
				input := <-player.inputChan
				if len(input) == 1 {
					fmt.Printf("player %x destroy\n", sessionId)
					return
				}
				player.lastInputTime = uint64(time.Now().Unix())
				t := binary.LittleEndian.Uint64(input[8:])
				dt := binary.LittleEndian.Uint64(input[16:])
				fmt.Printf("player %x process input: t = %x, dt = %x [cpu #%d]\n", player.sessionId, t, dt, cpu)
				for i := range player.state {
					player.state[i] ^= byte(t) + byte(i)
				}
				binary.LittleEndian.PutUint64(player.state[0:8], t+dt)
				err := playerStateMap.Put(sessionId, player.state)
				if err != nil {
					panic(err)
				}
				inputsProcessed++
				runtime.Gosched()
			}
		}()
	}
	player.inputChan <- input
	runtime.Gosched()
}

func main() {

	if len(os.Args) != 2 {
		fmt.Printf( "\nusage: ./player_server_worker <cpu_index>\n\n")
		os.Exit(0)
	}

	termChan := make(chan os.Signal, 1)

	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	var err error
	cpu, err =	strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("error: could not read cpu index\n")
		os.Exit(1)
	}

	fmt.Printf("player server worker running on cpu #%d\n", cpu)

	runtime.GOMAXPROCS(1)

	// get inputs processed map

	inputsProcessedMap, err = ebpf.LoadPinnedMap("/sys/fs/bpf/inputs_processed_map", nil)
	if err != nil {
		fmt.Printf("error: could not get inputs processed map: %v\n", err)
		os.Exit(1)
	}
	defer inputsProcessedMap.Close()

	// get player state map for our CPU

	player_state_outer, err := ebpf.LoadPinnedMap("/sys/fs/bpf/player_state_map", nil)
	if err != nil {
		fmt.Printf("error: could not get player state map: %v\n", err)
		os.Exit(1)
	}
	defer player_state_outer.Close()

	err = player_state_outer.Lookup(uint32(cpu), &playerStateMap)
	if err != nil {
		fmt.Printf("error: could not lookup player state map for cpu %d: %v\n", cpu, err)
		os.Exit(1)
	}

	// get input buffer map for our CPU

	input_buffer_outer, err := ebpf.LoadPinnedMap("/sys/fs/bpf/input_buffer_map", nil)
	if err != nil {
		fmt.Printf("error: could not get input buffer map: %v\n", err)
		os.Exit(1)
	}
	defer input_buffer_outer.Close()

	var input_buffer_inner *ebpf.Map
	err = input_buffer_outer.Lookup(uint32(cpu), &input_buffer_inner)
	if err != nil {
		fmt.Printf("error: could not lookup input buffer for cpu %d: %v\n", cpu, err)
		os.Exit(1)
	}

	// create input ring buffer

	input_buffer, err := ringbuf.NewReader(input_buffer_inner)

	// create player map

	playerMap = make(map[uint64]*PlayerData)

	// periodically clean up the player map

	go func() {
		ticker := time.NewTicker(time.Second)
	 	for {
		 	<-ticker.C
		 	currentTime := uint64(time.Now().Unix())
		 	for k,v := range playerMap {
			    if v.lastInputTime + PlayerTimeout < currentTime {
			    	v.inputChan <- make([]byte, 1)
			    	delete(playerMap, k)
			    }
			}
	 	}
	}()

	// update inputs processed map once per-second

	go func() {
		ticker := time.NewTicker(time.Second)
	 	for {
		 	<-ticker.C
			inputsProcessedMap.Put(&cpu, inputsProcessed)
	 	}
	}()

	// poll ring buffer to read inputs

	go func() {
		
		for {
			record, err := input_buffer.Read()
			if err != nil {
				fmt.Printf("error: failed to read from ring buffer: %v\n", err)
				os.Exit(1)
			}
			processInput(record.RawSample)
		}
	}()

	<- termChan
}
