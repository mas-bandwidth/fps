package main

import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"runtime"
	"syscall"
    "net"
)

const NumPlayers = 1000

var playerUpdates uint64

func updatePlayers() {

	for i := range NumPlayers {
		
		go func(sessionId uint64) {

	        conn, err := net.Dial("tcp", "127.0.0.1:50000")
	        if err != nil {
	            fmt.Printf("\nerror: could not connect to world database: %v\n\n", err)
	            os.Exit(1)
	        }

	        defer conn.Close()

	        ticker := time.NewTicker(time.Millisecond*10)

	        state := make([]byte, PlayerStateBytes)

	        frame := uint64(0)
	        t := uint64(0)
	        dt := uint64(5)

			for {
			 	<-ticker.C

			 	SendPingPacket(conn)

		        pong := ReceivePacket(conn)
		        if pong == nil {
		        	fmt.Printf("disconnected\n")
		        	return
		        }

		       	if pong[0] != PongPacket {
		        	panic("expected pong packet")
		        }

		        SendPlayerStatePacket(conn, sessionId, frame, t, state)
		        
		        t += dt
		        frame++

		        playerUpdates++
		 	}

		}(uint64(i))

		time.Sleep(time.Millisecond)
	}
}

func printStats() {
	ticker := time.NewTicker(time.Second)
	previousPlayerUpdates := uint64(0)
	for {
 		<-ticker.C
 		currentPlayerUpdates := playerUpdates
 		playerUpdateDelta := currentPlayerUpdates - previousPlayerUpdates
 		fmt.Printf("player update delta = %d\n", playerUpdateDelta)
 		previousPlayerUpdates = currentPlayerUpdates
 	}
}

func main() {

	runtime.GOMAXPROCS(1)

	termChan := make(chan os.Signal, 1)

	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	go updatePlayers()

	go printStats()

	<- termChan
}
