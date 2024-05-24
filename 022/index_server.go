package main

import (
    "fmt"
    "os"
    "sync"
    "math/rand"
    "encoding/binary"

    "github.com/maurice2k/tcpserver"
)

const Port = 60000

var playerServerMutex        sync.Mutex
var playerServerMapById      map[uint32]*ServerData
var playerServerMapByAddress map[string]*ServerData

var zoneDatabaseMutex        sync.Mutex
var zoneDatabaseMapById      map[uint32]*ServerData
var zoneDatabaseMapByAddress map[string]*ServerData

var world *World

func main() {

    world = generateGridWorld(2, 1, 2, Kilometer)

    world.Print()

    grid := createGrid(world, 100*Meter)

    _ = grid

    playerServerMapById = make(map[uint32]*ServerData)
    playerServerMapByAddress = make(map[string]*ServerData)

    zoneDatabaseMapById = make(map[uint32]*ServerData)
    zoneDatabaseMapByAddress = make(map[string]*ServerData)

    server, err := tcpserver.NewServer(fmt.Sprintf("127.0.0.1:%d", Port))

    if err != nil {
        fmt.Printf("error: could not start tcp server: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("index server started on port %d\n", Port)

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

            SendIndexServerPacket_Pong(conn)

        case IndexServerPacket_PlayerServerConnect:

            id := rand.Uint32()
            playerServerMutex.Lock()
            for {
                if id == 0 || playerServerMapById[id] == nil {
                    break
                }
                id = rand.Uint32()
            }
            playerServerMutex.Unlock()

            serverAddress := conn.GetClientAddr()

            fmt.Printf("player server %s connected [0x%08x]\n", serverAddress, id)

            SendIndexServerPacket_PlayerServerConnectResponse(conn, id)

            serverData := &ServerData{
                id:         id,
                address:    serverAddress,
            }

            addressString := serverAddress.String()

            playerServerMutex.Lock()
            playerServerMapById[id] = serverData
            playerServerMapByAddress[addressString] = serverData
            playerServerMutex.Unlock()

        case IndexServerPacket_PlayerServerUpdate:

            serverAddress := conn.GetClientAddr()

            playerServerMutex.Lock()
            serverData := playerServerMapByAddress[serverAddress.String()]
            playerServerMutex.Unlock()

            if serverData == nil {
                fmt.Printf("warning: unknown player server %s\n", serverAddress)
                return
            }

            id := serverData.id

            fmt.Printf("player server %s update [0x%08x]\n", serverAddress, id)

            playerServerMutex.Lock()
            playerServers := make([]*ServerData, len(playerServerMapById))
            index := 0
            for _,v := range playerServerMapById {
                playerServers[index] = v
                index++
            }
            playerServerMutex.Unlock()

            SendIndexServerPacket_PlayerServerUpdateResponse(conn, playerServers)

        case IndexServerPacket_PlayerServerDisconnect:

            serverAddress := conn.GetClientAddr()

            addressString := serverAddress.String()

            playerServerMutex.Lock()
            serverData := playerServerMapByAddress[addressString]
            if serverData != nil {
                delete(playerServerMapById, serverData.id)
                delete(playerServerMapByAddress, addressString)
            }
            playerServerMutex.Unlock()

            if serverData == nil {
                fmt.Printf("warning: unknown player server %s disconnected\n", addressString)
                return
            }

            fmt.Printf("player server %s disconnected [0x%08x]\n", addressString, serverData.id)

            SendIndexServerPacket_PlayerServerDisconnectResponse(conn)

        case IndexServerPacket_WorldRequest:

            SendIndexServerPacket_WorldResponse(conn, world)

        case IndexServerPacket_ZoneDatabaseConnect:

            zoneId := binary.LittleEndian.Uint32(packetData[1:])

            serverAddress := conn.GetClientAddr()

            zoneDatabaseMutex.Lock()

            if zoneId == 0 {

                // find a free zone and assign it

                found := false
                for i := range world.zones {
                    _, exists := zoneDatabaseMapById[world.zones[i].id]
                    if !exists {
                        fmt.Printf("found free zone 0x%08x\n", world.zones[i].id)
                        zoneId = world.zones[i].id
                        found = true
                        break
                    }
                }

                // no free zone

                if !found {
                    fmt.Printf("warning: no free zone available\n")
                    return
                }

            } else {

                // assign to a specific zone id

                fmt.Printf("zone database connecting as specific zone id 0x%08x\n", zoneId)

                _, exists := zoneDatabaseMapById[zoneId]
                if exists {
                    fmt.Printf("warning: zone 0x%08x is already allocated\n", zoneId)
                    return
                }
            }

            serverData := &ServerData{
                id:     zoneId,
                address: serverAddress,
            }

            zoneDatabaseMapById[zoneId] = serverData
            zoneDatabaseMapByAddress[serverAddress.String()] = serverData
            
            zoneDatabaseMutex.Unlock()

            fmt.Printf("zone database %s connected [0x%08x]\n", serverAddress.String(), zoneId)

            SendIndexServerPacket_ZoneDatabaseConnectResponse(conn, zoneId)

        case IndexServerPacket_ZoneDatabaseDisconnect:

            serverAddress := conn.GetClientAddr()

            addressString := serverAddress.String()

            zoneDatabaseMutex.Lock()
            serverData := zoneDatabaseMapByAddress[addressString]
            if serverData != nil {
                delete(zoneDatabaseMapById, serverData.id)
                delete(zoneDatabaseMapByAddress, addressString)
            }
            zoneDatabaseMutex.Unlock()

            if serverData == nil {
                fmt.Printf("warning: unknown zone database %s disconnected\n", addressString)
                return
            }

            fmt.Printf("zone database %s disconnected [0x%08x]\n", addressString, serverData.id)

            SendIndexServerPacket_ZoneDatabaseDisconnectResponse(conn)
        }
    }
}
