package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	// Uncomment this block to pass the first stage
	// "net"
	// "os"
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
		req, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		data, err := bufio.NewReader(req).ReadString('\n')
		if err != nil {
			log.Print(err.Error())
			return
		}

		dataFields := strings.Fields(data)
		endpoint := strings.Split(dataFields[1], "/")
		switch endpoint[1] {
		case "":
			req.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		case "echo":
			body := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 3\r\n\r\n" + endpoint[2]
			req.Write([]byte(body))
		default:
			req.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}

	}
}
