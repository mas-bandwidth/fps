package main

import (
    "fmt"
    "net"
    "bufio"
)

const Port = 50000

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

    reader := bufio.NewReader(conn)

    for {

        line, err := reader.ReadString('\n')
        if err != nil {
            fmt.Printf("error: could not read line: %v\n", err)
            return
        }

        line = strings.TrimSpace(string(line))

        if line == "ping" {
            fmt.Printf("%s: ping -> pong\n", conn.RemoteAddr().String())            
            conn.Write([]byte(string("pong\n")))
        }

        if line == "quit" {
            fmt.Printf("%s: quits\n", conn.RemoteAddr().String())
            break
        }

    }
}
