package main

import (
	"io"
    "fmt"
	"encoding/binary"
    "math"
    "net"
)

const PlayerStateBytes = 100

type ServerData struct {
    id          uint32
    address     *net.TCPAddr
}

const MaxWorldPacketSize = 4 * 1024

// ---------------------------------------------------------

func WriteBool(data []byte, index *int, value bool) {
    if value {
        data[*index] = byte(1)
    } else {
        data[*index] = byte(0)
    }

    *index += 1
}

func WriteUint8(data []byte, index *int, value uint8) {
    data[*index] = byte(value)
    *index += 1
}

func WriteUint16(data []byte, index *int, value uint16) {
    binary.LittleEndian.PutUint16(data[*index:], value)
    *index += 2
}

func WriteUint32(data []byte, index *int, value uint32) {
    binary.LittleEndian.PutUint32(data[*index:], value)
    *index += 4
}

func WriteUint64(data []byte, index *int, value uint64) {
    binary.LittleEndian.PutUint64(data[*index:], value)
    *index += 8
}

func WriteInt64(data []byte, index *int, value int64) {
    binary.LittleEndian.PutUint64(data[*index:], uint64(value))
    *index += 8
}

func WriteInt(data []byte, index *int, value int) {
    binary.LittleEndian.PutUint64(data[*index:], uint64(value))
    *index += 8
}

func WriteFloat32(data []byte, index *int, value float32) {
    uintValue := math.Float32bits(value)
    WriteUint32(data, index, uintValue)
}

func WriteFloat64(data []byte, index *int, value float64) {
    uintValue := math.Float64bits(value)
    WriteUint64(data, index, uintValue)
}

func WriteString(data []byte, index *int, value string, maxStringLength uint32) {
    stringLength := uint32(len(value))
    if stringLength > maxStringLength {
        panic("string is too long!\n")
    }
    binary.LittleEndian.PutUint32(data[*index:], stringLength)
    *index += 4
    for i := 0; i < int(stringLength); i++ {
        data[*index] = value[i]
        *index++
    }
}

func WriteBytes(data []byte, index *int, value []byte, numBytes int) {
    for i := 0; i < numBytes; i++ {
        data[*index] = value[i]
        *index++
    }
}

func WriteAddress(index *int, buffer []byte, address *net.TCPAddr) {
    ipv4 := address.IP.To4()
    port := address.Port
    buffer[*index+0] = ipv4[0]
    buffer[*index+1] = ipv4[1]
    buffer[*index+2] = ipv4[2]
    buffer[*index+3] = ipv4[3]
    buffer[*index+4] = (byte)(port & 0xFF)
    buffer[*index+5] = (byte)(port >> 8)
    *index += 6
}

// ------------------------------------------------------------------------

func ReadBool(data []byte, index *int, value *bool) bool {

    if *index+1 > len(data) {
        return false
    }

    if data[*index] > 0 {
        *value = true
    } else {
        *value = false
    }

    *index += 1
    return true
}

func ReadUint8(data []byte, index *int, value *uint8) bool {
    if *index+1 > len(data) {
        return false
    }
    *value = data[*index]
    *index += 1
    return true
}

func ReadUint16(data []byte, index *int, value *uint16) bool {
    if *index+2 > len(data) {
        return false
    }
    *value = binary.LittleEndian.Uint16(data[*index:])
    *index += 2
    return true
}

func ReadUint32(data []byte, index *int, value *uint32) bool {
    if *index+4 > len(data) {
        return false
    }
    *value = binary.LittleEndian.Uint32(data[*index:])
    *index += 4
    return true
}

func ReadUint64(data []byte, index *int, value *uint64) bool {
    if *index+8 > len(data) {
        return false
    }
    *value = binary.LittleEndian.Uint64(data[*index:])
    *index += 8
    return true
}

func ReadInt64(data []byte, index *int, value *int64) bool {
    if *index+8 > len(data) {
        return false
    }
    *value = int64(binary.LittleEndian.Uint64(data[*index:]))
    *index += 8
    return true
}

func ReadInt(data []byte, index *int, value *int) bool {
    if *index+8 > len(data) {
        return false
    }
    *value = int(binary.LittleEndian.Uint64(data[*index:]))
    *index += 8
    return true
}

func ReadFloat32(data []byte, index *int, value *float32) bool {
    var intValue uint32
    if !ReadUint32(data, index, &intValue) {
        return false
    }
    *value = math.Float32frombits(intValue)
    return true
}

func ReadFloat64(data []byte, index *int, value *float64) bool {
    var uintValue uint64
    if !ReadUint64(data, index, &uintValue) {
        return false
    }
    *value = math.Float64frombits(uintValue)
    return true
}

func ReadString(data []byte, index *int, value *string, maxStringLength uint32) bool {
    var stringLength uint32
    if !ReadUint32(data, index, &stringLength) {
        return false
    }
    if stringLength > maxStringLength {
        return false
    }
    if *index+int(stringLength) > len(data) {
        return false
    }
    stringData := make([]byte, stringLength)
    for i := uint32(0); i < stringLength; i++ {
        stringData[i] = data[*index]
        *index++
    }
    *value = string(stringData)
    return true
}

func ReadBytes(data []byte, index *int, value []byte, bytes uint32) bool {
    if *index+int(bytes) > len(data) {
        return false
    }
    for i := uint32(0); i < bytes; i++ {
        value[i] = data[*index]
        *index++
    }
    return true
}

func ReadAddress(index *int, buffer []byte) *net.TCPAddr {
    address := net.TCPAddr{IP: net.IPv4(buffer[*index+0], buffer[*index+1], buffer[*index+2], buffer[*index+3]), Port: ((int)(binary.LittleEndian.Uint16(buffer[*index+4:])))}
    *index += 6
    return &address
}

// ---------------------------------------------------------

const ZoneDatabasePacket_Ping = 0
const ZoneDatabasePacket_Pong = 1
const ZoneDatabasePacket_PlayerState = 2

func SendZoneDatabasePacket_Ping(conn net.Conn) {
    ping := [5]byte{}
    binary.LittleEndian.PutUint32(ping[:4], 1)
    ping[4] = PlayerServerPacket_Ping
    conn.Write(ping[:])
}

func SendZoneDatabasePacket_Pong(conn net.Conn) {
    pong := [5]byte{}
    binary.LittleEndian.PutUint32(pong[:4], 1)
    pong[4] = PlayerServerPacket_Pong
    conn.Write(pong[:])
}

func SendZoneDatabasePacket_PlayerState(conn net.Conn, sessionId uint64, frame uint64, t uint64, state []byte) {
    packet := [4+1+8+8+8+PlayerStateBytes]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1+8+8+8+PlayerStateBytes)
    packet[4] = ZoneDatabasePacket_PlayerState
    binary.LittleEndian.PutUint64(packet[5:], sessionId)
    binary.LittleEndian.PutUint64(packet[13:], frame)
    binary.LittleEndian.PutUint64(packet[21:], t)
    copy(packet[29:], state)
    conn.Write(packet[:])
}

// ---------------------------------------------------------

const WorldServerPacket_Ping = 0
const WorldServerPacket_Pong = 1
const WorldServerPacket_PlayerServerConnect = 2
const WorldServerPacket_PlayerServerConnectResponse = 3
const WorldServerPacket_PlayerServerDisconnect = 4
const WorldServerPacket_PlayerServerDisconnectResponse = 5
const WorldServerPacket_PlayerServerUpdate = 6
const WorldServerPacket_PlayerServerUpdateResponse = 7
const WorldServerPacket_WorldRequest = 8
const WorldServerPacket_WorldResponse = 9
const WorldServerPacket_ZoneDatabaseConnect = 10
const WorldServerPacket_ZoneDatabaseConnectResponse = 11
const WorldServerPacket_ZoneDatabaseDisconnect = 12
const WorldServerPacket_ZoneDatabaseDisconnectResponse = 13

func SendWorldServerPacket_Ping(conn net.Conn) {
    ping := [5]byte{}
    binary.LittleEndian.PutUint32(ping[:4], 1)
    ping[4] = WorldServerPacket_Ping
    conn.Write(ping[:])
}

func SendWorldServerPacket_Pong(conn net.Conn) {
    pong := [5]byte{}
    binary.LittleEndian.PutUint32(pong[:4], 1)
    pong[4] = WorldServerPacket_Pong
    conn.Write(pong[:])
}

func SendWorldServerPacket_PlayerServerConnect(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = WorldServerPacket_PlayerServerConnect
    conn.Write(packet[:])
}

func SendWorldServerPacket_PlayerServerConnectResponse(conn net.Conn, id uint32) {
    packet := make([]byte, 4+1+4+4)
    binary.LittleEndian.PutUint32(packet[:4], uint32(len(packet)-4))
    packet[4] = WorldServerPacket_PlayerServerConnectResponse
    binary.LittleEndian.PutUint32(packet[4+1:], id)
    conn.Write(packet[:])
}

func SendWorldServerPacket_PlayerServerDisconnect(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = WorldServerPacket_PlayerServerDisconnect
    conn.Write(packet[:])
}

func SendWorldServerPacket_PlayerServerDisconnectResponse(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = WorldServerPacket_PlayerServerDisconnectResponse
    conn.Write(packet[:])
}

func SendWorldServerPacket_PlayerServerUpdate(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = WorldServerPacket_PlayerServerUpdate
    conn.Write(packet[:])
}

func SendWorldServerPacket_PlayerServerUpdateResponse(conn net.Conn, playerServers []*ServerData) {
    packet := make([]byte, 4+1+4+(4+6)*len(playerServers))
    binary.LittleEndian.PutUint32(packet[:4], uint32(len(packet)-4))
    packet[4] = WorldServerPacket_PlayerServerUpdateResponse
    binary.LittleEndian.PutUint32(packet[4+1:], uint32(len(playerServers)))
    index := 4 + 1 + 4
    for i := range playerServers {
        binary.LittleEndian.PutUint32(packet[index:], playerServers[i].id)
        index += 4
        WriteAddress(&index, packet, playerServers[i].address)
    }
    conn.Write(packet[:])
}

func SendWorldServerPacket_WorldRequest(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = WorldServerPacket_WorldRequest
    conn.Write(packet[:])
}

func SendWorldServerPacket_WorldResponse(conn net.Conn, world *World) {
    packet := make([]byte, MaxWorldPacketSize)
    packet[4] = WorldServerPacket_WorldResponse
    index := 5
    world.Write(packet, &index)
    packet = packet[:index]
    binary.LittleEndian.PutUint32(packet[:4], uint32(len(packet)-4))
    conn.Write(packet)
}

func SendWorldServerPacket_ZoneDatabaseConnect(conn net.Conn, zoneId uint32) {
    packet := [4+1+4]byte{}
    binary.LittleEndian.PutUint32(packet[:4], uint32(len(packet)-4))
    packet[4] = WorldServerPacket_ZoneDatabaseConnect
    binary.LittleEndian.PutUint32(packet[5:], zoneId)
    conn.Write(packet[:])
}

func SendWorldServerPacket_ZoneDatabaseConnectResponse(conn net.Conn, zoneId uint32) {
    fmt.Printf("write zone id of 0x%08x\n", zoneId)
    packet := make([]byte, 4+1+4)
    binary.LittleEndian.PutUint32(packet[:4], uint32(len(packet)-4))
    packet[4] = WorldServerPacket_ZoneDatabaseConnectResponse
    binary.LittleEndian.PutUint32(packet[4+1:], zoneId)
    conn.Write(packet[:])
}

func SendWorldServerPacket_ZoneDatabaseDisconnect(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = WorldServerPacket_ZoneDatabaseDisconnect
    conn.Write(packet[:])
}

func SendWorldServerPacket_ZoneDatabaseDisconnectResponse(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = WorldServerPacket_ZoneDatabaseDisconnectResponse
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

// ---------------------------------------------------------
