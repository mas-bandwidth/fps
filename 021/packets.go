package main

import (
	"io"
	"encoding/binary"
    "net"
)

const PlayerStateBytes = 100

// ---------------------------------------------------------

const WorldDatabasePacket_Ping = 0
const WorldDatabasePacket_Pong = 1
const WorldDatabasePacket_PlayerState = 2

func SendWorldDatabasePacket_Ping(conn net.Conn) {
    ping := [5]byte{}
    binary.LittleEndian.PutUint32(ping[:4], 1)
    ping[4] = PlayerServerPacket_Ping
    conn.Write(ping[:])
}

func SendWorldDatabasePacket_Pong(conn net.Conn) {
    pong := [5]byte{}
    binary.LittleEndian.PutUint32(pong[:4], 1)
    pong[4] = PlayerServerPacket_Pong
    conn.Write(pong[:])
}

func SendWorldDatabasePacket_PlayerState(conn net.Conn, sessionId uint64, frame uint64, t uint64, state []byte) {
    packet := [4+1+8+8+8+PlayerStateBytes]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1+8+8+8+PlayerStateBytes)
    packet[4] = WorldDatabasePacket_PlayerState
    binary.LittleEndian.PutUint64(packet[5:], sessionId)
    binary.LittleEndian.PutUint64(packet[13:], frame)
    binary.LittleEndian.PutUint64(packet[21:], t)
    copy(packet[29:], state)
    conn.Write(packet[:])
}

// ---------------------------------------------------------

const IndexServerPacket_Ping = 0
const IndexServerPacket_Pong = 1
const IndexServerPacket_PlayerServerConnect = 2
const IndexServerPacket_PlayerServerConnectResponse = 3
const IndexServerPacket_PlayerServerDisconnect = 4
const IndexServerPacket_PlayerServerDisconnectResponse = 5
const IndexServerPacket_PlayerServerConnected = 6
const IndexServerPacket_PlayerServerDisconnected = 7

func SendIndexServerPacket_Ping(conn net.Conn) {
    ping := [5]byte{}
    binary.LittleEndian.PutUint32(ping[:4], 1)
    ping[4] = IndexServerPacket_Ping
    conn.Write(ping[:])
}

func SendIndexServerPacket_Pong(conn net.Conn) {
    pong := [5]byte{}
    binary.LittleEndian.PutUint32(pong[:4], 1)
    pong[4] = IndexServerPacket_Pong
    conn.Write(pong[:])
}

func SendIndexServerPacket_PlayerServerConnect(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = IndexServerPacket_PlayerServerConnect
    conn.Write(packet[:])
    // todo: return tag?
}

func SendIndexServerPacket_PlayerServerDisconnect(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = IndexServerPacket_PlayerServerDisconnect
    conn.Write(packet[:])
}

// ---------------------------------------------------------

const PlayerServerPacket_Ping = 0
const PlayerServerPacket_Pong = 1

func SendPlayerServerPacket_Ping(conn net.Conn) {
    ping := [5]byte{}
    binary.LittleEndian.PutUint32(ping[:4], 1)
    ping[4] = PlayerServerPacket_Ping
    conn.Write(ping[:])
}

func SendPlayerServerPacket_Pong(conn net.Conn) {
    pong := [5]byte{}
    binary.LittleEndian.PutUint32(pong[:4], 1)
    pong[4] = PlayerServerPacket_Pong
    conn.Write(pong[:])
}

// ---------------------------------------------------------

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
