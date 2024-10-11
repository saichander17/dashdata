package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type Server struct {
	store      *Store
	port       string
	workerPool chan struct{}
	maxQueue   int
	timeout    time.Duration
}

func NewServer(store *Store, port string, maxWorkers, maxQueue int, timeout time.Duration) *Server {
	return &Server{
		store:      store,
		port:       port,
		workerPool: make(chan struct{}, maxWorkers),
		maxQueue:   maxQueue,
		timeout:    timeout,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Server listening on port %s\n", s.port)

	queue := make(chan net.Conn, s.maxQueue)
	go s.processQueue(queue)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		select {
		case queue <- conn:
			// Connection added to queue
		default:
			// Queue is full, reject the connection
			conn.Close()
			fmt.Println("Connection rejected: queue full")
		}
	}
}

func (s *Server) processQueue(queue chan net.Conn) {
	for conn := range queue {
		select {
		case s.workerPool <- struct{}{}:
			go func(c net.Conn) {
				s.handleConnection(c)
				<-s.workerPool
			}(conn)
		case <-time.After(s.timeout):
			conn.Close()
			fmt.Println("Connection timeout: worker unavailable")
		}
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
