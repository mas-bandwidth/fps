package main

import (
    "fmt"
    "net"
    "time"
    "sync"
    "encoding/binary"
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

    listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", Port))
    if err != nil {
        fmt.Printf("error: could not listen on tcp socket: %v\n", err)
        return
    }
    defer listener.Close()

    fmt.Printf("world database is listening on port %d\n", Port)

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("error: could not accept client connection: %v\n", err)
            continue
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {

    fmt.Printf("new connection from %s\n", conn.RemoteAddr().String())

    defer conn.Close()

    for {

        packetData := ReceivePacket(conn)

        if packetData == nil {
            return
        }

        switch packetData[0] {

        case PingPacket:

            SendPongPacket(conn)

        case PlayerStatePacket:

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
