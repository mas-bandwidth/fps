package main

import (
	"io"
	"encoding/binary"
    "net"
)

const PlayerStateBytes = 100

const Meter = 1000000
const Kilometer = 1000 * Meter
const Centimeter = Meter / 100
const Millimeter = Meter / 1000
const Micrometer = Meter / 1000000

type ServerData struct {
    tag         uint32
    address     *net.TCPAddr
}

// ---------------------------------------------------------

func WriteAddress(buffer []byte, address *net.TCPAddr) {
    ipv4 := address.IP.To4()
    port := address.Port
    buffer[0] = ipv4[0]
    buffer[1] = ipv4[1]
    buffer[2] = ipv4[2]
    buffer[3] = ipv4[3]
    buffer[4] = (byte)(port & 0xFF)
    buffer[5] = (byte)(port >> 8)
}

func ReadAddress(buffer []byte) *net.TCPAddr {
    return &net.TCPAddr{IP: net.IPv4(buffer[0], buffer[1], buffer[2], buffer[3]), Port: ((int)(binary.LittleEndian.Uint16(buffer[4:])))}
}

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
const IndexServerPacket_PlayerServerUpdate = 6
const IndexServerPacket_PlayerServerUpdateResponse = 7

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
}

func SendIndexServerPacket_PlayerServerConnectResponse(conn net.Conn, tag uint32) {
    packet := make([]byte, 4+1+4+4)
    binary.LittleEndian.PutUint32(packet[:4], uint32(len(packet)-4))
    packet[4] = IndexServerPacket_PlayerServerConnectResponse
    binary.LittleEndian.PutUint32(packet[1+4:], tag)
    conn.Write(packet[:])
}

func SendIndexServerPacket_PlayerServerDisconnect(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = IndexServerPacket_PlayerServerDisconnect
    conn.Write(packet[:])
}

func SendIndexServerPacket_PlayerServerDisconnectResponse(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = IndexServerPacket_PlayerServerDisconnectResponse
    conn.Write(packet[:])
}

func SendIndexServerPacket_PlayerServerUpdate(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = IndexServerPacket_PlayerServerUpdate
    conn.Write(packet[:])
}

func SendIndexServerPacket_PlayerServerUpdateResponse(conn net.Conn, playerServers []*ServerData) {
    packet := make([]byte, 4+1+4+(4+6)*len(playerServers))
    binary.LittleEndian.PutUint32(packet[:4], uint32(len(packet)-4))
    packet[4] = IndexServerPacket_PlayerServerUpdateResponse
    binary.LittleEndian.PutUint32(packet[4+1:], uint32(len(playerServers)))
    index := 4 + 1 + 4
    for i := range playerServers {
        binary.LittleEndian.PutUint32(packet[index:], playerServers[i].tag)
        WriteAddress(packet[index+4:], playerServers[i].address)
        index += 4 + 6
    }
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
