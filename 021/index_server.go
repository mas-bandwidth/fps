package main

import (
    "fmt"
    "os"
    "sync"
    "math/rand"

    "github.com/maurice2k/tcpserver"
)

const Port = 60000

var playerServerMapByTag     map[uint32]*ServerData

var playerServerMapByAddress map[string]*ServerData

var serverMutex sync.Mutex

var worldConfig WorldConfig

func main() {

    worldConfig.gridWidth = 2
    worldConfig.gridHeight = 2
    worldConfig.gridSize = Kilometer
    worldConfig.calcDerived()

    fmt.Printf("world is a %dx%d grid across (0,0) -> (%.1f, %.1f) kms\n", worldConfig.gridWidth, worldConfig.gridHeight, float64(worldConfig.width)/float64(Kilometer), float64(worldConfig.height)/float64(Kilometer))

    playerServerMapByTag = make(map[uint32]*ServerData)
    playerServerMapByAddress = make(map[string]*ServerData)

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

        case WorldDatabasePacket_Ping:

            SendIndexServerPacket_Pong(conn)

        case IndexServerPacket_PlayerServerConnect:

            tag := rand.Uint32()
            serverMutex.Lock()
            for {
                if playerServerMapByTag[tag] == nil {
                    break
                }
                tag = rand.Uint32()
            }
            serverMutex.Unlock()

            serverAddress := conn.GetClientAddr()

            fmt.Printf("player server %s connected [0x%08x]\n", serverAddress, tag)

            SendIndexServerPacket_PlayerServerConnectResponse(conn, tag)

            serverData := &ServerData{
                tag:        tag,
                address:    serverAddress,
            }

            addressString := serverAddress.String()

            serverMutex.Lock()
            playerServerMapByTag[tag] = serverData
            playerServerMapByAddress[addressString] = serverData
            serverMutex.Unlock()

        case IndexServerPacket_PlayerServerUpdate:

            serverAddress := conn.GetClientAddr()

            serverMutex.Lock()
            serverData := playerServerMapByAddress[serverAddress.String()]
            serverMutex.Unlock()

            if serverData == nil {
                fmt.Printf("warning: unknown player server %s\n", serverAddress)
                return
            }

            tag := serverData.tag

            fmt.Printf("player server %s update [0x%08x]\n", serverAddress, tag)

            serverMutex.Lock()
            playerServers := make([]*ServerData, len(playerServerMapByTag))
            index := 0
            for _,v := range playerServerMapByTag {
                playerServers[index] = v
                index++
            }
            serverMutex.Unlock()

            SendIndexServerPacket_PlayerServerUpdateResponse(conn, playerServers)

        case IndexServerPacket_PlayerServerDisconnect:

            serverAddress := conn.GetClientAddr()

            addressString := serverAddress.String()

            serverMutex.Lock()
            serverData := playerServerMapByAddress[addressString]
            if serverData != nil {
                delete(playerServerMapByTag, serverData.tag)
                delete(playerServerMapByAddress, addressString)
            }
            serverMutex.Unlock()

            if serverData == nil {
                fmt.Printf("warning: unknown server %s disconnected\n", addressString)
                return
            }

            fmt.Printf("player server %s disconnected [0x%08x]\n", addressString, serverData.tag)

            SendIndexServerPacket_PlayerServerDisconnectResponse(conn)
        }
    }
}
