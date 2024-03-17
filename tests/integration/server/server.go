package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type server struct {
	Ip         string
	StartPort  int
	NumServers int
}

func main() {
	var (
		ip         string
		startPort  int
		numServers int
	)

	flag.StringVar(&ip, "ip", "127.0.0.1", "IP address for the server")
	flag.IntVar(&startPort, "start_port", 8081, "Start port for the server")
	flag.IntVar(&numServers, "num_servers", 1, "Number of servers to run")
	flag.Parse()

	server := server{
		Ip:         ip,
		StartPort:  startPort,
		NumServers: numServers,
	}

	server.StartListening()
}

func (s *server) StartListening() {
	var wg sync.WaitGroup

	wg.Add(s.NumServers)

	for i := 0; i < s.NumServers; i++ {
		port := s.StartPort + i
		addr := fmt.Sprintf("%s:%d", s.Ip, port)

		go func(serverID int) {
			defer wg.Done()

			s.listen(serverID, addr)
		}(i)
	}

	wg.Wait()
}

func (s *server) listen(serverID int, addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("ServerID: %d, Error listening: %s\n", serverID, err.Error())

		return
	}
	defer ln.Close()

	fmt.Printf("ServerID: %d, TCP Server Listening on :%s\n", serverID, addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("ServerID: %d, Error accepting: %s\n", serverID, err.Error())

			continue
		}

		s.handleConnection(serverID, conn)
	}
}

func (s *server) handleConnection(serverID int, conn net.Conn) {
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	remoteAddr := conn.RemoteAddr().String()

	fmt.Printf("ServerID: %d, Local Address: %s, Remote Address: %s\n", serverID, localAddr, remoteAddr)

	// Set a deadline for writing to the connection
	writeDeadline := time.Now().Add(5 * time.Second)
	if err := conn.SetWriteDeadline(writeDeadline); err != nil {

		fmt.Printf("ServerID: %d, Failed to set write deadline: %v\n", serverID, err)

		return
	}

	// Example read operation
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)

	if err != nil {
		if os.IsTimeout(err) {
			fmt.Printf("ServerID: %d, Read timed out\n", serverID)
		} else {
			fmt.Printf("ServerID: %d, Error reading: %s\n", serverID, err.Error())
		}

		return
	}

	response := "Hello, client! Your request has been processed.\n"
	_, err = conn.Write([]byte(response))

	if err != nil {
		if os.IsTimeout(err) {
			fmt.Printf("ServerID: %d, Write timed out\n", serverID)
		} else {
			fmt.Printf("ServerID: %d, Failed to send response to client: %v\n", serverID, err)
		}

		return
	}

	time.Sleep(1 * time.Millisecond)
}
