package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// Request struct definition
type Request struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    []byte
}

// Function to parse the HTTP request from a net.Conn
func parseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)

	// Initialize the Request struct
	req := &Request{
		Headers: make(map[string]string),
	}

	// Read the request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading request line: %w", err)
	}
	requestLine = strings.TrimSpace(requestLine)
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", requestLine)
	}
	req.Method = parts[0]
	req.Path = parts[1]
	req.Version = parts[2]

	// Read headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading header line: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			// End of headers
			break
		}
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			headerName := strings.TrimSpace(headerParts[0])
			headerValue := strings.TrimSpace(headerParts[1])
			req.Headers[headerName] = headerValue
		}
	}

	// Read the body if it is a POST request
	if req.Method == "POST" {
		contentLength, err := strconv.Atoi(req.Headers["Content-Length"])
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length: %w", err)
		}
		body := make([]byte, contentLength)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			return nil, fmt.Errorf("error reading body: %w", err)
		}
		req.Body = body
	}

	return req, nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Parse the request
	req, err := parseRequest(conn)
	if err != nil {
		fmt.Println("Error parsing request:", err)
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	// Check if the path starts with "/echo/"
	if strings.HasPrefix(req.Path, "/echo/") {
		// Extract the dynamic part after "/echo/"
		str := req.Path
		lastParam := strings.Split(str, "/")[len(strings.Split(str, "/"))-1]
		contentLength := len(lastParam)

		// Write the headers and the body
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, lastParam)
		conn.Write([]byte(response))
		return
	}

	if strings.HasPrefix(req.Path, "/user-agent") {
		userAgent := req.Headers["User-Agent"]
		contentLength := len(userAgent)

		// Write the headers and the body
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, userAgent)
		conn.Write([]byte(response))
		return
	}

	if strings.HasPrefix(req.Path, "/files/") {

		directory := os.Args[2]
		fileName := strings.TrimPrefix(req.Path, "/files/")

		if req.Method == "GET" {

			data, err := os.ReadFile(directory + fileName)

			if err != nil {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			} else {
				conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + strconv.Itoa(len(data)) + "\r\n\r\n" + string(data) + "\r\n\r\n"))
			}
			return
		} else if req.Method == "POST" {
			// Handle POST request
			err := os.WriteFile(directory+fileName, req.Body, 0644)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			} else {
				conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
			}
			return
		}
	}

	// If the path is exactly "/", return 200 OK with no content
	if req.Path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return
	}

	// For any other path, return 404 Not Found
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func main() {
	// Start a TCP server
	listener, err := net.Listen("tcp", ":4221")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is running on http://localhost:4221")

	// ruuning the server
	for {
		// Accept a new connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}
