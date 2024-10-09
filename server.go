package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Server struct {
	store      *Store
	port       string
	workerPool chan struct{}
}

func NewServer(store *Store, port string, maxWorkers int) *Server {
	return &Server{
		store:      store,
		port:       port,
		workerPool: make(chan struct{}, maxWorkers),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Server listening on port %s\n", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		s.workerPool <- struct{}{}
		go func() {
			s.handleConnection(conn)
			<-s.workerPool
		}()
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		command := scanner.Text()
		response := s.executeCommand(command)
		conn.Write([]byte(response + "\n"))
	}
}

func (s *Server) executeCommand(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "ERROR: Empty command"
	}

	switch strings.ToUpper(parts[0]) {
	case "GET":
		if len(parts) != 2 {
			return "ERROR: GET command requires one key"
		}
		value, exists := s.store.Get(parts[1])
		if !exists {
			return "NOT FOUND"
		}
		return value
	case "SET":
		if len(parts) != 3 {
			return "ERROR: SET command requires key and value"
		}
		s.store.Set(parts[1], parts[2])
		return "OK"
	case "DELETE":
		if len(parts) != 2 {
			return "ERROR: DELETE command requires one key"
		}
		s.store.Delete(parts[1])
		return "OK"
	default:
		return "ERROR: Unknown command"
	}
}
