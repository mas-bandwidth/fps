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
    "encoding/binary"
    "sync"

    "github.com/maurice2k/tcpserver"
)

const NumPlayers = 250

var playerUpdates uint64

var indexServer net.Conn
var indexServerMutex sync.Mutex

var playerServerMap map[uint32]*ServerData
var playerServerMapMutex sync.Mutex

var world *World

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

            SendPlayerServerPacket_Pong(conn)

        	// ...
        }
    }
}

func connectToWorldServer() {

	// open tcp connection to world server

	fmt.Printf( "connecting to world server\n" )

	var err error

    indexServer, err = net.Dial("tcp", "127.0.0.1:60000")
    if err != nil {
        fmt.Printf("\nerror: could not connect to world server: %v\n\n", err)
        os.Exit(1)
    }

	indexServerMutex.Lock()

	// ping it

 	SendWorldServerPacket_Ping(indexServer)

    pong := ReceivePacket(indexServer)
    if pong == nil {
    	fmt.Printf("disconnected from world server\n")
    	return
    }

	indexServerMutex.Unlock()

   	if pong[0] != WorldServerPacket_Pong {
    	panic("expected pong packet")
    }

    // connect to world server

	fmt.Printf("connected to world server\n")

	indexServerMutex.Lock()

 	SendWorldServerPacket_PlayerServerConnect(indexServer)

    packetData := ReceivePacket(indexServer)

	indexServerMutex.Unlock()

    if packetData == nil {
    	fmt.Printf("disconnected from world server\n")
        return
    }

    if packetData[0] != WorldServerPacket_PlayerServerConnectResponse {
    	panic("expected player server connect response packet")
    }

    tag := binary.LittleEndian.Uint32(packetData[1:])

    fmt.Printf("player server tag is 0x%08x\n", tag)

    // update player servers from world server

    updatePlayerServers()

    go func() {
	    ticker := time.NewTicker(time.Second)
    	for {
			<-ticker.C
			updatePlayerServers()
    	}
    }()

    // get world from world server

    requestWorld()
}

func updatePlayerServers() {
    
    // update player servers

	indexServerMutex.Lock()

 	SendWorldServerPacket_PlayerServerUpdate(indexServer)

    packetData := ReceivePacket(indexServer)

	indexServerMutex.Unlock()

	if packetData == nil {
		fmt.Printf("error: disconnected from world server\n")
		os.Exit(1)
	}

    if packetData[0] != WorldServerPacket_PlayerServerUpdateResponse {
    	panic("expected player server update response packet")
    }

    numPlayerServers := binary.LittleEndian.Uint32(packetData[1:])

    fmt.Printf("----------------------------------------\n")
    index := 1 + 4
    for i := 0; i < int(numPlayerServers); i++ {
        tag := binary.LittleEndian.Uint32(packetData[index:])
        index += 4
        address := ReadAddress(&index ,packetData)
        fmt.Printf("[0x%08x] %s\n", tag, address.String())
        // todo: store tag -> address mapping etc.
    }	
    fmt.Printf("----------------------------------------\n")
}

func requestWorld() {
    
	indexServerMutex.Lock()

 	SendWorldServerPacket_WorldRequest(indexServer)

    packetData := ReceivePacket(indexServer)

	indexServerMutex.Unlock()

	if packetData == nil {
		fmt.Printf("error: disconnected from world server\n")
		os.Exit(1)
	}

    if packetData[0] != WorldServerPacket_WorldResponse {
    	panic("expected world response packet")
    }

    world = &World{}
    index := 1
    if !world.Read(packetData, &index) {
    	panic("could not read world\n")
    }

    world.Print()
}

func updatePlayers() {

	// update players

	for _ = range NumPlayers {
		
		sessionId := rand.Uint64()

		go func(sessionId uint64) {

	        zone_database, err := net.Dial("tcp", "127.0.0.1:50000")
	        if err != nil {
	            fmt.Printf("\nerror: could not connect to zone database: %v\n\n", err)
	            os.Exit(1)
	        }

	        defer zone_database.Close()

	        ticker := time.NewTicker(time.Millisecond*10)

	        state := make([]byte, PlayerStateBytes)

	        frame := uint64(0)
	        t := uint64(0)
	        dt := uint64(5)

			for {
			 	<-ticker.C

			 	SendZoneDatabasePacket_Ping(zone_database)

		        pong := ReceivePacket(zone_database)
		        if pong == nil {
		        	fmt.Printf("error: disconnected from zone database\n")
		        	os.Exit(1)
		        }

		       	if pong[0] != ZoneDatabasePacket_Pong {
		        	panic("expected pong packet")
		        }

		        SendZoneDatabasePacket_PlayerState(zone_database, sessionId, frame, t, state)
		        
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

func cleanShutdown() {

	fmt.Printf("disconnecting\n")

	indexServerMutex.Lock()

 	SendWorldServerPacket_PlayerServerDisconnect(indexServer)

    packetData := ReceivePacket(indexServer)

	indexServerMutex.Unlock()

    if packetData == nil {
    	fmt.Printf("disconnected from world server\n")
        return
    }

    if packetData[0] != WorldServerPacket_PlayerServerDisconnectResponse {
    	panic("expected player server disconnect response packet")
    }

    indexServer.Close()

    playerServerMapMutex.Lock()
	clear(playerServerMap)
    playerServerMapMutex.Unlock()

	fmt.Printf("disconnected\n")
}

func main() {

	runtime.GOMAXPROCS(1)

	termChan := make(chan os.Signal, 1)

	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	port := 20000

	listenForCommands(port)

	connectToWorldServer()

	go updatePlayers()

	go printStats()

	<- termChan

	cleanShutdown()
}
