package main

import (
    "fmt"
    "net"
    "io"
    "time"
    "encoding/binary"
)

const Port = 50000
const HistorySize = 1024
const PlayerStateBytes = 100

const PingPacket = 0
const PongPacket = 1
const PlayerStatePacket = 2

type PlayerData struct {
    lastUpdateTime uint64
    t              [HistorySize]uint64
    state          [HistorySize][PlayerStateBytes]byte
}

var playerMap map[uint64]*PlayerData

func main() {

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

        // read packet length

        var buffer [4]byte
        index := 0
        for {
            n, err := conn.Read(buffer[index:4])
            if err == io.EOF {
                return
            }
            index += n
            if index == 4 {
                break
            }
        }

        length := binary.LittleEndian.Uint32(buffer[:])

        // quit on zero length

        if length == 0 {
            return
        }

        // read packet data

        packetData := make([]byte, length)
        index = 0
        for {
            n, err := conn.Read(buffer[index:length])
            if err == io.EOF {
                return
            }
            index += n
            if index == int(length) {
                break
            }
        }

        // handle packet type

        switch packetData[0] {

        case PingPacket:

            response := [5]byte{}
            binary.LittleEndian.PutUint32(response[:4], 1)
            response[4] = PongPacket
            conn.Write(response[:])

        case PlayerStatePacket:

            if len(packetData) < 1 + 8 + 8 + 8 {
                return
            }
        
            sessionId := binary.LittleEndian.Uint64(packetData[1:1+8])
            frame := binary.LittleEndian.Uint64(packetData[1+8:1+8+8])
            t := binary.LittleEndian.Uint64(packetData[1+8+8:1+8+8+8])
        
            player := playerMap[sessionId]
            if player == nil {
                player = &PlayerData{}
                playerMap[sessionId] = player
            }

            index := frame % HistorySize

            player.lastUpdateTime = uint64(time.Now().Unix())
            player.t[index] = t
            copy(player.state[index][:], packetData[1+8+8+8:]) 
        }
    }
}
