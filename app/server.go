package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		buf := make([]byte, 1024)
		length, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Error reading: %#v\n", err)
			return
		}

		var requestData []string
		lines := strings.Split(string(buf[:length]), "\n")
		requestData = append(requestData, lines...)

		dataFields := strings.Fields(requestData[0])
		endpoint := strings.Split(dataFields[1], "/")

		switch endpoint[1] {
		case "":
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		case "echo":
			body := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(endpoint[2]), endpoint[2])
			conn.Write([]byte(body))
		case "user-agent":
			var userAgent string
			for _, header := range requestData {
				headerData := strings.Split(header, ":")
				if headerData[0] == "User-Agent" {
					userAgent = strings.TrimSpace(headerData[1])
					break
				}
			}
			body := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
			conn.Write([]byte(body))
		default:
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}

		conn.Close()

	}
}
