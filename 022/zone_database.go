package main

import (
    "fmt"
    "time"
    "sync"
    "os"
    "net"
    "strconv"
    "encoding/binary"
    "os/signal"
    "syscall"

    "github.com/maurice2k/tcpserver"
)

const Port = 50000

const HistorySize = 1024

type PlayerData struct {
    lastUpdateTime uint64
    t              [HistorySize]uint64
    state          [HistorySize][PlayerStateBytes]byte
}

var playerMap map[uint64]*PlayerData

var playerMapMutex sync.Mutex

var indexServer net.Conn
var indexServerMutex sync.Mutex

var world *World

func connectToIndexServer(zoneId *uint32) {

    // open tcp connection to index server

    fmt.Printf( "connecting to index server\n" )

    var err error

    indexServer, err = net.Dial("tcp", "127.0.0.1:60000")
    if err != nil {
        fmt.Printf("\nerror: could not connect to index server: %v\n\n", err)
        os.Exit(1)
    }

    indexServerMutex.Lock()

    // ping it

    SendIndexServerPacket_Ping(indexServer)

    pong := ReceivePacket(indexServer)
    if pong == nil {
        fmt.Printf("disconnected from index server\n")
        return
    }

    indexServerMutex.Unlock()

    if pong[0] != IndexServerPacket_Pong {
        panic("expected pong packet")
    }

    // connect to index server

    fmt.Printf("connected to index server\n")

    indexServerMutex.Lock()

    SendIndexServerPacket_ZoneDatabaseConnect(indexServer, *zoneId)

    packetData := ReceivePacket(indexServer)

    indexServerMutex.Unlock()

    if packetData == nil {
        fmt.Printf("disconnected from index server\n")
        return
    }

    if packetData[0] != IndexServerPacket_ZoneDatabaseConnectResponse {
        panic("expected zone database connect response packet")
    }

    *zoneId = binary.LittleEndian.Uint32(packetData[1:])

    // get world from index server

    requestWorld()
}

func requestWorld() {
    
    indexServerMutex.Lock()

    SendIndexServerPacket_WorldRequest(indexServer)

    packetData := ReceivePacket(indexServer)

    indexServerMutex.Unlock()

    if packetData == nil {
        fmt.Printf("error: disconnected from index server\n")
        os.Exit(1)
    }

    if packetData[0] != IndexServerPacket_WorldResponse {
        panic("expected world response packet")
    }

    world = &World{}
    index := 1
    if !world.Read(packetData, &index) {
        panic("could not read world\n")
    }

    world.Print()
}

func cleanShutdown() {

    fmt.Printf("disconnecting\n")

    indexServerMutex.Lock()

    SendIndexServerPacket_ZoneDatabaseDisconnect(indexServer)

    packetData := ReceivePacket(indexServer)

    indexServerMutex.Unlock()

    if packetData == nil {
        fmt.Printf("disconnected from index server\n")
        return
    }

    if packetData[0] != IndexServerPacket_ZoneDatabaseDisconnectResponse {
        panic("expected zone database disconnect response packet")
    }

    indexServer.Close()

    fmt.Printf("disconnected\n")
}

func main() {

    termChan := make(chan os.Signal, 1)

    signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

    zoneId := uint32(0)
    if len(os.Args) == 2 {
        value, err := strconv.ParseInt(os.Args[1], 10, 32)
        if err != nil {
            panic(err)
        }
        zoneId = uint32(value)
    }

    connectToIndexServer(&zoneId)

    fmt.Printf("zone id is 0x%08x\n", zoneId)

    playerMap = make(map[uint64]*PlayerData)

    server, err := tcpserver.NewServer(fmt.Sprintf("127.0.0.1:%d", Port))

    if err != nil {
        fmt.Printf("error: could not start tcp server: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("zone database started on port %d\n", Port)

    server.SetRequestHandler(requestHandler)
    server.Listen()
    go server.Serve()

    <- termChan

    cleanShutdown()
}

func requestHandler(conn tcpserver.Connection) {

    for {

        packetData := ReceivePacket(conn)

        if packetData == nil {
            return
        }

        switch packetData[0] {

        case ZoneDatabasePacket_Ping:

            SendZoneDatabasePacket_Pong(conn)

        case ZoneDatabasePacket_PlayerState:

            if len(packetData) != 1 + 8 + 8 + 8 + PlayerStateBytes {
                return
            }
        
            sessionId := binary.LittleEndian.Uint64(packetData[1:1+8])
            frame := binary.LittleEndian.Uint64(packetData[1+8:1+8+8])
            t := binary.LittleEndian.Uint64(packetData[1+8+8:1+8+8+8])
        
            playerMapMutex.Lock()

            player := playerMap[sessionId]
            if player == nil {
                player = &PlayerData{}
                playerMap[sessionId] = player
            }

            index := frame % HistorySize

            player.lastUpdateTime = uint64(time.Now().Unix())
            player.t[index] = t
            copy(player.state[index][:], packetData[1+8+8+8:]) 

            playerMapMutex.Unlock()
        }
    }
}
