package main

import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"runtime"
	"syscall"
    "net"
    "math/rand"

    "github.com/maurice2k/tcpserver"
)

const NumPlayers = 250

var playerUpdates uint64

func listenForCommands(port int) {

    server, err := tcpserver.NewServer(fmt.Sprintf("127.0.0.1:%d", port))

    if err != nil {
        fmt.Printf("error: could not start tcp server: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("player server started on port %d\n", port)

    server.SetRequestHandler(requestHandler)
    
    server.Listen()

    go server.Serve()
}

func requestHandler(conn tcpserver.Connection) {

    for {

        packetData := ReceivePacket(conn)

        if packetData == nil {
            return
        }

        switch packetData[0] {

        case PlayerServerPacket_Ping:

            fmt.Printf("ping -> pong\n")

            SendPlayerServerPacket_Pong(conn)

        	// ...
        }
    }
}

func connectToIndexServer() {

	// connect to index server

	fmt.Printf( "connecting to index server\n" )

    index_server, err := net.Dial("tcp", "127.0.0.1:60000")
    if err != nil {
        fmt.Printf("\nerror: could not connect to index server: %v\n\n", err)
        os.Exit(1)
    }

    defer index_server.Close()

 	SendIndexServerPacket_Ping(index_server)

    pong := ReceivePacket(index_server)
    if pong == nil {
    	fmt.Printf("disconnected from index server\n")
    	return
    }

   	if pong[0] != IndexServerPacket_Pong {
    	panic("expected pong packet")
    }

	fmt.Printf( "connected to index server\n" )

 	SendIndexServerPacket_PlayerServerConnect(index_server)

    // todo: send player server connect packet

    // todo: receive player server connect response packet

    // todo: goroutine, handle packets sent from the index server
}

func updatePlayers() {

	// update players

	for _ = range NumPlayers {
		
		sessionId := rand.Uint64()

		go func(sessionId uint64) {

	        world_database, err := net.Dial("tcp", "127.0.0.1:50000")
	        if err != nil {
	            fmt.Printf("\nerror: could not connect to world database: %v\n\n", err)
	            os.Exit(1)
	        }

	        defer world_database.Close()

	        ticker := time.NewTicker(time.Millisecond*10)

	        state := make([]byte, PlayerStateBytes)

	        frame := uint64(0)
	        t := uint64(0)
	        dt := uint64(5)

			for {
			 	<-ticker.C

			 	SendWorldDatabasePacket_Ping(world_database)

		        pong := ReceivePacket(world_database)
		        if pong == nil {
		        	fmt.Printf("disconnected from world server\n")
		        	return
		        }

		       	if pong[0] != WorldDatabasePacket_Pong {
		        	panic("expected pong packet")
		        }

		        SendWorldDatabasePacket_PlayerState(world_database, sessionId, frame, t, state)
		        
		        t += dt
		        frame++

		        playerUpdates++
		 	}

		}(sessionId)

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

	port := 20000

	listenForCommands(port)

	connectToIndexServer()

	go updatePlayers()

	go printStats()

	<- termChan
}
