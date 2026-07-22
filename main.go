package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	StatusOK      = 200
	StatusCreated = 201

	StatusBadRequest = 400
	StatusNotFound   = 404

	StatusInternalServerError = 500
)

func StatusText(code int) string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusCreated:
		return "Created"
	case StatusBadRequest:
		return "Bad request"
	case StatusNotFound:
		return "Not found"
	case StatusInternalServerError:
		return "Internal server error"
	default:
		return "Invalid code"
	}
}

// Server type defines a HTTP server
type Server struct {
	listener net.Listener
	Addr     string
	routes   map[string]Handler
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

type Handler interface {
	ServeHTTP(*Request, ResponseWriter)
}

type HandlerFunc func(*Request, ResponseWriter)

func (f HandlerFunc) ServeHTTP(r *Request, w ResponseWriter) {
	f(r, w)
}

type ResponseWriter interface {
	// Write will write a string back to the connection as HTTP response.
	//  This will be mainly used for writing the body of the response.
	WriteBody(string)

	// SendCode() will send HTTP response code to the client
	SendCode(int)

	// WriteHeader() will write response header to the client
	WriteHeader(key, value string)

	// Write() will flush the buffer to the connection.
	Write()
}

type responseWriter struct {
	w *bufio.Writer
}

func (r *responseWriter) Write() {
	r.w.Flush()
}

func (r *responseWriter) WriteBody(s string) {
	r.w.WriteString("\r\n")
	r.w.WriteString(s)

}

func (r *responseWriter) SendCode(code int) {
	resLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", code, StatusText(code))
	r.w.WriteString(resLine)
}

func (r *responseWriter) WriteHeader(key, value string) {
	h := fmt.Sprintf("%s: %s\r\n", key, value)
	r.w.WriteString(h)

}

func NewServer(addr string) *Server {
	return &Server{
		Addr:   addr,
		routes: make(map[string]Handler),
	}
}

func (s *Server) AddRoute(path string, handler Handler) {
	s.routes[path] = handler
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

	resWriter := &responseWriter{w: bufio.NewWriter(conn)}

	url := req.URL
	for path, handler := range s.routes {
		if path == url {
			handler.ServeHTTP(req, resWriter)
		}
	}

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

func hello(r *Request, w ResponseWriter) {
	w.SendCode(StatusOK)
	w.WriteHeader("Content-Length", "18")
	w.WriteBody("Hello from server\n")
	w.Write()
}

func main() {
	// Let's create a server
	srv := NewServer("localhost:8080")
	srv.AddRoute("/hello", HandlerFunc(hello))

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("Can't start the server")
	}

}
