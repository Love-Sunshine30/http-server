package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// Server type defines a HTTP server
type Server struct {
	listener net.Listener
	Addr     string
	routes   map[string]string
}

// Request type defines a HTTP request
type Request struct {
	Method  string
	URL     string
	Version string
	Headers map[string]string
	body    string
}

// Response type defines a HTTP response
type Response struct {
	Version    string
	StatusCode int
	StatusText string
	Headers    map[string]string
	body       string
}

func NewServer(addr string) *Server {
	return &Server{
		Addr: addr,
	}
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	s.listener = l
	defer l.Close()

	log.Printf("Server listening on: %s", s.Addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		go s.Serve(conn)
	}
}

func (s *Server) Serve(conn net.Conn) {
	defer conn.Close()

	req, err := s.parseRequest(conn)

	if err != nil {
		log.Println("Parsing Error")
		return
	}

	fmt.Fprintf(os.Stdout, "%+v", req)
	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 17\r\n\r\n"))
	conn.Write([]byte("Hello from server"))
}

func (s *Server) parseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)
	request := &Request{
		Headers: make(map[string]string),
	}

	reqLine, err := reader.ReadString('\n')
	if err != nil {
		return request, err
	}
	reqLine = strings.TrimSuffix(reqLine, "\r\n")
	parts := strings.Split(reqLine, " ")

	if len(parts) != 3 {
		return request, fmt.Errorf("[ERROR] malformed requsest line: %s", reqLine)
	}
	request.Method = parts[0]
	request.URL = parts[1]
	request.Version = parts[2]

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return request, err
		}

		if line == "\r\n" {
			break
		}
		line = strings.TrimSuffix(line, "\r\n")
		key, value, found := strings.Cut(line, ":")

		if !found {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		value = strings.TrimSpace(strings.ToLower(value))
		request.Headers[key] = value
	}

	if contentLength, exists := request.Headers["content-length"]; exists {
		size, err := strconv.Atoi(contentLength)
		if err != nil {
			return request, err
		}

		buf := make([]byte, size)
		_, err = io.ReadFull(reader, buf)

		if err != nil {
			return request, err
		}
		request.body = string(buf)
	}

	return request, nil
}

func main() {
	// Let's create a server
	srv := NewServer("localhost:8080")

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("Can't start the server")
	}

}
