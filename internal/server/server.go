package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
	"github.com/saichander17/dashdata/internal/store"
)

type Server struct {
	store      store.Store
	port       string
	workerPool chan struct{}
	maxQueue   int
	timeout    time.Duration
}

func NewServer(store store.Store, port string, maxWorkers, maxQueue int, timeout time.Duration) *Server {
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
    reader := bufio.NewReader(conn)

    for {
        command, err := readCommand(reader)
        if err != nil {
            if err != io.EOF {
                fmt.Printf("Error reading command: %v\n", err)
            }
            return
        }

        response := s.executeCommand(command)
        conn.Write([]byte(response))
    }
}

func readCommand(reader *bufio.Reader) ([]string, error) {
    // Read the first byte to determine the type of input
    b, err := reader.ReadByte()
    if err != nil {
        return nil, err
    }

    if b == '*' {
        // RESP Array
        return readRESPArray(reader)
    } else {
        // Inline command
        reader.UnreadByte()
        line, err := reader.ReadString('\n')
        if err != nil {
            return nil, err
        }
        return strings.Fields(strings.TrimSpace(line)), nil
    }
}

func readRESPArray(reader *bufio.Reader) ([]string, error) {
    countStr, err := reader.ReadString('\n')
    if err != nil {
        return nil, err
    }
    count, err := strconv.Atoi(strings.TrimSpace(countStr))
    if err != nil {
        return nil, err
    }

    command := make([]string, count)
    for i := 0; i < count; i++ {
        b, err := reader.ReadByte()
        if err != nil {
            return nil, err
        }
        if b != '$' {
            return nil, fmt.Errorf("expected '$', got '%c'", b)
        }

        lengthStr, err := reader.ReadString('\n')
        if err != nil {
            return nil, err
        }
        length, err := strconv.Atoi(strings.TrimSpace(lengthStr))
        if err != nil {
            return nil, err
        }

        valueBytes := make([]byte, length+2) // +2 for \r\n
        _, err = io.ReadFull(reader, valueBytes)
        if err != nil {
            return nil, err
        }
        command[i] = string(valueBytes[:length])
    }

    return command, nil
}

func (s *Server) executeCommand(command []string) string {
    if len(command) == 0 {
        return "-ERR empty command\r\n"
    }

    switch strings.ToUpper(command[0]) {
    case "GET":
        if len(command) != 2 {
            return "-ERR wrong number of arguments for 'get' command\r\n"
        }
        value, exists := s.store.Get(command[1])
        if !exists {
            return "$-1\r\n"
        }
        return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
    case "SET":
        if len(command) != 3 {
            return "-ERR wrong number of arguments for 'set' command\r\n"
        }
        s.store.Set(command[1], command[2])
        return "+OK\r\n"
    case "DEL":
        if len(command) != 2 {
            return "-ERR wrong number of arguments for 'del' command\r\n"
        }
        s.store.Delete(command[1])
        return ":1\r\n"
    default:
        return "-ERR unknown command\r\n"
    }
}
