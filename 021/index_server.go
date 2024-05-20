package main

import (
    "fmt"
    "os"

    "github.com/maurice2k/tcpserver"
)

const Port = 60000

func main() {

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

            fmt.Printf("ping -> pong\n")

            SendIndexServerPacket_Pong(conn)

        case IndexServerPacket_PlayerServerConnect:

            fmt.Printf("player server %s connected\n", conn.GetClientAddr().String())

        case IndexServerPacket_PlayerServerDisconnect:

            fmt.Printf("player server %s disconnected\n", conn.GetClientAddr().String())
        }
    }
}
