package main

import (
	"fmt"
	"net"
	"sync"
	"time"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"strconv"
	"sync/atomic"
	"encoding/binary"
)

const StartPort = 10000
const MaxPacketSize = 1384
const SocketBufferSize = 2*1024*1024

const InputSize = 100
const InputsPerPacket = 10
const InputHistory = 1024

const PlayerDataSize = 1024

const InputPacketSize = 1 + 8 + 8 + 8 + (8 + InputSize) * InputsPerPacket
const JoinRequestPacketSize = 1 + 8 + 8 + PlayerDataSize
const JoinResponsePacketSize = 1 + 8 + 8 + 8
const StatsRequestPacketSize = 1 + 8
const StatsResponsePacketSize = 1 + 8

const JoinRequestPacket = 1
const JoinResponsePacket = 2
const InputPacket = 3
const StatsRequestPacket = 4
const StatsResponsePacket = 5

var numClients int

var quit uint64
var joined uint64
var serverTime uint64
var packetsSent uint64
var packetsReceived uint64
var totalInputsProcessed uint64

type Input struct {
	sequence uint64
	t        uint64
	dt       uint64
	input    []byte	
}

func GetInt(name string, defaultValue int) int {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueString, 10, 64)
	if err != nil {
		return defaultValue
	}
	return int(value)
}

func GetAddress(name string, defaultValue string) net.UDPAddr {
	valueString, ok := os.LookupEnv(name)
	if !ok {
	    valueString = defaultValue
	}
	value, err := net.ResolveUDPAddr("udp", valueString)
	if err != nil {
		panic(fmt.Sprintf("invalid address in envvar %s", name))
	}
	return *value
}

func main() {

	serverAddress := GetAddress("SERVER_ADDRESS", "127.0.0.1:40000")

	numClients = GetInt("NUM_CLIENTS", 1)

	fmt.Printf("starting %d clients\n", numClients)

	fmt.Printf("server address is %s\n", serverAddress.String())

	var wg sync.WaitGroup

	for i := 0; i < numClients; i++ {
		go func(clientIndex int) {
			wg.Add(1)
			runClient(clientIndex, &serverAddress)
			wg.Done()
		}(i)
	}

	termChan := make(chan os.Signal, 1)

	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Second)
 
	prev_sent := uint64(0)
	prev_processed := uint64(0)
	prev_player_states := uint64(0)

 	for {
		select {
		case <-termChan:
			fmt.Printf("\nreceived shutdown signal\n")
			atomic.StoreUint64(&quit, 1)
	 	case <-ticker.C:
	 		sent := atomic.LoadUint64(&packetsSent)
	 		processed := atomic.LoadUint64(&totalInputsProcessed)
	 		player_states := atomic.LoadUint64(&playerStatePacketsReceived)
	 		sent_delta := sent - prev_sent
	 		processed_delta := processed - prev_processed
	 		player_state_delta := player_states - prev_player_states
	 		fmt.Printf("inputs sent delta %d, inputs processed delta %d, player state delta %d\n", sent_delta, processed_delta, player_state_delta)
			prev_sent = sent
			prev_processed = processed
			prev_player_states = player_states
	 	}
		quit := atomic.LoadUint64(&quit)
		if quit != 0 {
			break
		}
 	}

	fmt.Printf("shutting down\n")

	wg.Wait()	

	fmt.Printf("done.\n")
}

func sampleInput(sequence uint64, t uint64, dt uint64) Input {
	// todo: here you would sample player input, eg. keyboard, mouse or controller
	return Input{input: make([]byte, 100), sequence: sequence, t: t, dt: dt}
}

func addInput(sequence uint64, inputBuffer []Input, input Input) {
	index := sequence % InputHistory
	inputBuffer[index] = input
}

func writeJoinRequestPacket(sessionId uint64, sentTime uint64, playerData []byte) []byte {
	packet := make([]byte, InputPacketSize)
	packetIndex := 0
	packet[0] = JoinRequestPacket
	packetIndex++
	binary.LittleEndian.PutUint64(packet[packetIndex:], sessionId)
	packetIndex += 8
	binary.LittleEndian.PutUint64(packet[packetIndex:], sentTime)
	packetIndex += 8
	copy(packet[packetIndex:], playerData)
	packetIndex += PlayerDataSize
	return packet[:packetIndex]
}

func writeInputPacket(sessionId uint64, sequence uint64, inputBuffer []Input) []byte {
	index := sequence % InputHistory
	input := inputBuffer[index]
	packet := make([]byte, InputPacketSize)
	packetIndex := 0
	packet[0] = InputPacket
	packetIndex++
	binary.LittleEndian.PutUint64(packet[packetIndex:], sessionId)
	packetIndex += 8
	binary.LittleEndian.PutUint64(packet[packetIndex:], sequence)
	packetIndex += 8
	binary.LittleEndian.PutUint64(packet[packetIndex:], input.t)
	packetIndex += 8
	for i := 0; i < InputsPerPacket; i++ {
		binary.LittleEndian.PutUint64(packet[packetIndex:], input.dt)
		packetIndex += 8
		copy(packet[packetIndex:], input.input)
		packetIndex += InputSize
		sequence --
		index = sequence % InputHistory
		input = inputBuffer[index]
		if input.sequence != sequence {
			break
		}
	}
	return packet[:packetIndex]
}

func writeStatsRequestPacket() []byte {
	packet := make([]byte, StatsRequestPacketSize)
	packet[0] = StatsRequestPacket
	return packet
}

func runClient(clientIndex int, serverAddress *net.UDPAddr) {

	addr := net.UDPAddr{
	    Port: 0,
	    IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic( "could not create client")
	}
	defer conn.Close()

	if err := conn.SetReadBuffer(SocketBufferSize); err != nil {
		panic(fmt.Sprintf("could not set socket read buffer size: %v", err))
	}

	if err := conn.SetWriteBuffer(SocketBufferSize); err != nil {
		panic(fmt.Sprintf("could not set socket write buffer size: %v", err))
	}

	buffer := make([]byte, MaxPacketSize)

	go func() {
		for {
	
			packetBytes, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				break
			}
	
			if packetBytes < 1 {
				continue
			}

			packetData := buffer[:packetBytes]

			packetType := packetData[0]

			if packetType == JoinResponsePacket && packetBytes == JoinResponsePacketSize {

				fmt.Printf("received join response packet\n")

				atomic.AddUint64(&joined, 1)

				sentTime := binary.LittleEndian.Uint64(packetData[1+8:])

				joinServerTime := binary.LittleEndian.Uint64(packetData[1+8+8:])

				rtt := uint64(time.Now().UnixNano()) - sentTime

				safety := uint64(100 * time.Millisecond)

				offset := rtt/2 + safety

				fmt.Printf("time offset is %d milliseconds\n", offset/1000000)

				startTime := joinServerTime + offset

				atomic.StoreUint64(&serverTime, startTime)

			} else if packetType == StatsResponsePacket && packetBytes == StatsResponsePacketSize {

				packetInputsProcessed := binary.LittleEndian.Uint64(packetData[1:])

				atomic.StoreUint64(&totalInputsProcessed, packetInputsProcessed)

			}

			atomic.AddUint64(&packetsReceived, 1)
		}
	}()

	// join

	rand.Seed(time.Now().UnixNano())

	sessionId := rand.Uint64()
	{
		fmt.Printf("joining server as session %016x\n", sessionId)

		playerData := make([]byte, PlayerDataSize)

		ticker := time.NewTicker(time.Millisecond * 10)

	 	for {
	 		stop := false

			select {

		 	case <-ticker.C:

				joined := atomic.LoadUint64(&joined)
				if joined > 0 {
					stop = true
				}

		 		fmt.Printf("sent join request packet\n")

		 		sentTime := uint64(time.Now().UnixNano())

				joinRequestPacket := writeJoinRequestPacket(sessionId, sentTime, playerData)

				conn.WriteToUDP(joinRequestPacket, serverAddress)
		 	}

			if stop {
				break
			}

			quit := atomic.LoadUint64(&quit)
			if quit != 0 {
				return
			}
	 	}
	}

 	// stats loop

 	go func() {

		ticker := time.NewTicker(time.Second)

	 	for {
			select {

		 	case <-ticker.C:

				statsRequestPacket := writeStatsRequestPacket()

				conn.WriteToUDP(statsRequestPacket, serverAddress)
		 	}

			quit := atomic.LoadUint64(&quit)
			if quit != 0 {
				return
			}
	 	}
 	}()

	// main loop

	fmt.Printf("server time is %d\n", serverTime)

	t := uint64(0)					// nanoseconds
	dt := uint64(1000000000)/100 	// 100ms in nanoseconds

	sequence := uint64(1000)

	inputBuffer := make([]Input, InputHistory)

	ticker := time.NewTicker(time.Millisecond * 10)

 	for {
		select {

	 	case <-ticker.C:

			input := sampleInput(sequence, t, dt)

			addInput(sequence, inputBuffer, input)

			inputPacket := writeInputPacket(sessionId, sequence, inputBuffer)

			conn.WriteToUDP(inputPacket, serverAddress)

			atomic.AddUint64(&packetsSent, 1)

			t += dt

			sequence++
	 	}

		quit := atomic.LoadUint64(&quit)
		if quit != 0 {
			return
		}
 	}
}
