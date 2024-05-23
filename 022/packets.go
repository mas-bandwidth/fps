package main

import (
	"io"
    "fmt"
	"encoding/binary"
    "math"
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

const MaxWorldPacketSize = 4 * 1024

// ---------------------------------------------------------

type Vector struct {
    x int64
    y int64
    z int64
}

func (value *Vector) Write(data []byte, index *int) {
    WriteInt64(data, index, value.x)
    WriteInt64(data, index, value.y)
    WriteInt64(data, index, value.z)
}

func (value *Vector) Read(data []byte, index *int) bool {
    if !ReadInt64(data, index, &value.x) {
        return false
    }
    if !ReadInt64(data, index, &value.y) {
        return false
    }
    if !ReadInt64(data, index, &value.z) {
        return false
    }
    return true
}

// ---------------------------------------------------------

type Plane struct {
    normal Vector
    d      int64
}

func (value *Plane) Write(data []byte, index *int) {
    value.normal.Write(data, index)
    WriteInt64(data, index, value.d)    
}

func (value *Plane) Read(data []byte, index *int) bool {
    if !value.normal.Read(data, index) {
        return false
    }
    if !ReadInt64(data, index, &value.d) {
        return false
    }
    return true
}

// ---------------------------------------------------------

type AABB struct {
    min Vector
    max Vector
}

func (value *AABB) Write(data []byte, index *int) {
    value.min.Write(data, index)
    value.max.Write(data, index)
}

func (value *AABB) Read(data []byte, index *int) bool {
    if !value.min.Read(data, index) {
        return false
    }
    if !value.max.Read(data, index) {
        return false
    }
    return true
}

// ---------------------------------------------------------

type Volume struct {
    bounds AABB
    planes []Plane
}

func (value *Volume) Write(data []byte, index *int) {
    value.bounds.Write(data, index)
    numPlanes := len(value.planes)
    WriteInt(data, index, numPlanes)
    for i := 0; i < numPlanes; i++ {
        value.planes[i].Write(data, index)
    }
}

func (value *Volume) Read(data []byte, index *int) bool {
    if !value.bounds.Read(data, index) {
        return false
    }
    var numPlanes int
    if !ReadInt(data, index, &numPlanes) {
        return false
    }
    value.planes = make([]Plane, numPlanes)
    for i := 0; i < numPlanes; i++ {
        if !value.planes[i].Read(data, index) {
            return false
        }
    }
    return true
}

// ---------------------------------------------------------

type Zone struct {
    id      uint32
    origin  Vector
    bounds  AABB
    volumes []Volume
}

func (value *Zone) Write(data []byte, index *int) {
    WriteUint32(data, index, value.id)
    value.origin.Write(data, index)
    value.bounds.Write(data, index)
    numVolumes := len(value.volumes)
    WriteInt(data, index, numVolumes)
    for i := 0; i < numVolumes; i++ {
        value.volumes[i].Write(data, index)
    }
}

func (value *Zone) Read(data []byte, index *int) bool {
    if !ReadUint32(data, index, &value.id) {
        return false
    }
    if !value.origin.Read(data, index) {
        return false
    }
    if !value.bounds.Read(data, index) {
        return false
    }
    var numVolumes int
    if !ReadInt(data, index, &numVolumes) {
        return false
    }
    value.volumes = make([]Volume, numVolumes)
    for i := 0; i < numVolumes; i++ {
        if !value.volumes[i].Read(data, index) {
            return false
        }
    }
    return true
}

// ---------------------------------------------------------

type World struct {
    bounds  AABB
    zones   []Zone
    zoneMap map[uint32]*Zone
}

func (world *World) Fixup() {
    world.zoneMap = make(map[uint32]*Zone, len(world.zones))
    for i := range world.zones {
        world.zoneMap[world.zones[i].id] = &world.zones[i]
    }
}

func (world *World) Print() {

    fmt.Printf("world bounds are (%d,%d,%d) -> (%d,%d,%d)\n", 
        world.bounds.min.x,
        world.bounds.min.y,
        world.bounds.min.z,
        world.bounds.max.x,
        world.bounds.max.y,
        world.bounds.max.z,
    )

    fmt.Printf("world has %d zones:\n", len(world.zones))

    for i := range world.zones {
        fmt.Printf(" + 0x%08x: (%d,%d,%d) -> (%d,%d,%d)\n",
            world.zones[i].id,
            world.zones[i].bounds.min.x,
            world.zones[i].bounds.min.y,
            world.zones[i].bounds.min.z,
            world.zones[i].bounds.max.x,
            world.zones[i].bounds.max.y,
            world.zones[i].bounds.max.z,
        )
    }
}

func (value *World) Write(data []byte, index *int) {
    value.bounds.Write(data, index)
    numZones := len(value.zones)
    WriteInt(data, index, numZones)
    for i := 0; i < numZones; i++ {
        value.zones[i].Write(data, index)
    }
}

func (value *World) Read(data []byte, index *int) bool {
    if !value.bounds.Read(data, index) {
        return false
    }
    var numZones int
    if !ReadInt(data, index, &numZones) {
        return false
    }
    value.zones = make([]Zone, numZones)
    for i := 0; i < numZones; i++ {
        if !value.zones[i].Read(data, index) {
            return false
        }
    }
    value.Fixup()
    return true
}

// ---------------------------------------------------------

func generateGridWorld(i int64, j int64, k int64, cellSize uint64) *World {

    fmt.Printf("generating grid world: %dx%dx%d\n", i, j, k)
    
    world := World{}

    world.bounds.max.x = i * int64(cellSize)
    world.bounds.max.y = j * int64(cellSize)
    world.bounds.max.z = k * int64(cellSize)

    numZones := i * j * k

    world.zones = make([]Zone, numZones)

    index := 0

    for y := int64(0); y < j; y++ {

        for z := int64(0); z < k; z++ {

            for x := int64(0); x < i; x++ {

                world.zones[index].id = uint32(index) + 1

                world.zones[index].bounds.min.x = x * int64(cellSize)
                world.zones[index].bounds.min.y = y * int64(cellSize)
                world.zones[index].bounds.min.z = z * int64(cellSize)

                world.zones[index].bounds.max.x = (x+1) * int64(cellSize)
                world.zones[index].bounds.max.y = (y+1) * int64(cellSize)
                world.zones[index].bounds.max.z = (z+1) * int64(cellSize)

                world.zones[index].origin.x = ( world.zones[index].bounds.min.x + world.zones[index].bounds.max.x ) / 2
                world.zones[index].origin.y = ( world.zones[index].bounds.min.y + world.zones[index].bounds.max.y ) / 2
                world.zones[index].origin.z = ( world.zones[index].bounds.min.z + world.zones[index].bounds.max.z ) / 2

                world.zones[index].volumes = make([]Volume, 1)

                world.zones[index].volumes[0].bounds = world.zones[index].bounds

                world.zones[index].volumes[0].planes = make([]Plane, 6)

                // left plane

                world.zones[index].volumes[0].planes[0].normal.x = Meter
                world.zones[index].volumes[0].planes[0].normal.y = 0
                world.zones[index].volumes[0].planes[0].normal.z = 0
                world.zones[index].volumes[0].planes[0].d = x * Meter

                // right plane

                world.zones[index].volumes[0].planes[1].normal.x = -Meter
                world.zones[index].volumes[0].planes[1].normal.y = 0
                world.zones[index].volumes[0].planes[1].normal.z = 0
                world.zones[index].volumes[0].planes[1].d = x * Meter + int64(cellSize)

                // bottom plane

                world.zones[index].volumes[0].planes[2].normal.x = 0
                world.zones[index].volumes[0].planes[2].normal.y = Meter
                world.zones[index].volumes[0].planes[2].normal.z = 0
                world.zones[index].volumes[0].planes[2].d = y * Meter

                // top plane

                world.zones[index].volumes[0].planes[3].normal.x = 0
                world.zones[index].volumes[0].planes[3].normal.y = -Meter
                world.zones[index].volumes[0].planes[3].normal.z = 0
                world.zones[index].volumes[0].planes[3].d = y * Meter + int64(cellSize)

                // front plane

                world.zones[index].volumes[0].planes[4].normal.x = 0
                world.zones[index].volumes[0].planes[4].normal.y = 0
                world.zones[index].volumes[0].planes[4].normal.z = Meter
                world.zones[index].volumes[0].planes[4].d = z * Meter

                // back plane

                world.zones[index].volumes[0].planes[5].normal.x = 0
                world.zones[index].volumes[0].planes[5].normal.y = 0
                world.zones[index].volumes[0].planes[5].normal.z = -Meter
                world.zones[index].volumes[0].planes[5].d = z * Meter + int64(cellSize)

                index++
            }

        }

    }
    
    return &world
}

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

const IndexServerPacket_Ping = 0
const IndexServerPacket_Pong = 1
const IndexServerPacket_PlayerServerConnect = 2
const IndexServerPacket_PlayerServerConnectResponse = 3
const IndexServerPacket_PlayerServerDisconnect = 4
const IndexServerPacket_PlayerServerDisconnectResponse = 5
const IndexServerPacket_PlayerServerUpdate = 6
const IndexServerPacket_PlayerServerUpdateResponse = 7
const IndexServerPacket_WorldRequest = 8
const IndexServerPacket_WorldResponse = 9

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
        index += 4
        WriteAddress(&index, packet, playerServers[i].address)
    }
    conn.Write(packet[:])
}

func SendIndexServerPacket_WorldRequest(conn net.Conn) {
    packet := [5]byte{}
    binary.LittleEndian.PutUint32(packet[:4], 1)
    packet[4] = IndexServerPacket_WorldRequest
    conn.Write(packet[:])
}

func SendIndexServerPacket_WorldResponse(conn net.Conn, world *World) {
    packet := make([]byte, MaxWorldPacketSize)
    packet[4] = IndexServerPacket_WorldResponse
    index := 5
    world.Write(packet, &index)
    packet = packet[:index]
    binary.LittleEndian.PutUint32(packet[:4], uint32(len(packet)-4))
    conn.Write(packet)
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
