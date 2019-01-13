package runnerserver

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

type RunnerServer struct {
	Routes map[string]RequestHandler
}

type Request struct {
	Headers    map[string]string
	Connection net.Conn
}

type RequestHandler func(r Request)

func (s *RunnerServer) Boot() {
	ln, _ := net.Listen("tcp", ":8080")
	for {
		conn, _ := ln.Accept()

		go s.handleConnection(conn)
	}
}

func (s *RunnerServer) handleConnection(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("FATAL ERROR: %v", r)
		}
	}()

	var headerBytes []byte
	var bodyBytes []byte

	for {
		buf := make([]byte, 12)
		_, err := conn.Read(buf)

		if err != nil {
			panic(err)
		}

		fmt.Printf("Header in progress: %s\n", string(buf))

		headerBytes = append(headerBytes, buf...)

		if slices := bytes.SplitN(headerBytes, []byte("ENDHEADERS\n"), 2); len(slices) >= 2 {
			headerBytes = slices[0]
			bodyBytes = slices[1]
			break
		}
	}

	headers := parseHeaders(headerBytes)

	fmt.Printf("Headers: %v\n", headers)

	request := Request{
		Headers:    headers,
		Connection: conn,
	}

	if handler, ok := s.Routes[request.Headers["Route"]]; ok {
		handler(request)
	} else {
		panic(fmt.Sprintf("No route found for %s", request.Headers["Route"]))
	}

	conn.Close()

	fmt.Println("Closed")
}

func parseHeaders(headerBytes []byte) map[string]string {
	headers := make(map[string]string)

	lines := bytes.Split(headerBytes, []byte("\n"))
	for _, line := range lines {
		parts := bytes.SplitN(line, []byte("="), 2)
		if len(parts) == 2 {
			headers[string(parts[0])] = string(parts[1])
		}
	}

	return headers
}