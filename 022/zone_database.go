package main

import (
    "fmt"
    "time"
    "sync"
    "os"
    "encoding/binary"

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

var playerMutex sync.Mutex

func main() {

    playerMap = make(map[uint64]*PlayerData)

    server, err := tcpserver.NewServer(fmt.Sprintf("127.0.0.1:%d", Port))

    if err != nil {
        fmt.Printf("error: could not start tcp server: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("zone database started on port %d\n", Port)

    server.SetRequestHandler(requestHandler)
    server.Listen()
    server.Serve()
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
        
            playerMutex.Lock()

            player := playerMap[sessionId]
            if player == nil {
                player = &PlayerData{}
                playerMap[sessionId] = player
            }

            index := frame % HistorySize

            player.lastUpdateTime = uint64(time.Now().Unix())
            player.t[index] = t
            copy(player.state[index][:], packetData[1+8+8+8:]) 

            playerMutex.Unlock()
        }
    }
}
