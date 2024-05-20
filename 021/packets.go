package main

import (
	"io"
	"encoding/binary"
    "net"
)

const PingPacket = 0
const PongPacket = 1
const PlayerStatePacket = 2

const PlayerStateBytes = 100

func SendPingPacket(conn net.Conn) {
    ping := [5]byte{}
    binary.LittleEndian.PutUint32(ping[:4], 1)
    ping[4] = PingPacket
    conn.Write(ping[:])
}

func SendPongPacket(conn net.Conn) {
    pong := [5]byte{}
    binary.LittleEndian.PutUint32(pong[:4], 1)
    pong[4] = PongPacket
    conn.Write(pong[:])
}

func SendPlayerStatePacket(conn net.Conn, sessionId uint64, frame uint64, t uint64, state []byte) {
    packet := [4+1+8+8+8+PlayerStateBytes]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1+8+8+8+PlayerStateBytes)
    packet[4] = PlayerStatePacket
    binary.LittleEndian.PutUint64(packet[5:], sessionId)
    binary.LittleEndian.PutUint64(packet[13:], frame)
    binary.LittleEndian.PutUint64(packet[21:], t)
    copy(packet[29:], state)
    conn.Write(packet[:])
}

func ReceivePacket(conn net.Conn) []byte {
    
    var buffer [4]byte
    index := 0
    for {
        n, err := conn.Read(buffer[index:4])
        if err == io.EOF {
            return nil
        }
        index += n
        if index == 4 {
            break
        }
    }

    length := binary.LittleEndian.Uint32(buffer[:])

    if length == 0 {
        return nil
    }

    packetData := make([]byte, length)
    index = 0
    for {
        n, err := conn.Read(packetData[index:length])
        if err == io.EOF {
            return nil	
        }
        index += n
        if index == int(length) {
            break
        }
    }

    return packetData
}
