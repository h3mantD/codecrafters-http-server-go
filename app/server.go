package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type RequestHeader struct {
	header string
	value  string
}

type RequestData struct {
	method   string
	endpoint string
	headers  map[string]string
	body     string
}

type ResponseData struct {
	statusCode string
	headers    map[string]string
	body       string
}

var directory *string

func main() {
	directory = flag.String("directory", "/tmp/", "Directory from file has to be fetched.")
	flag.Parse()
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

		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	length, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Error reading: %#v\n", err)
		return
	}

	var requestData []string
	lines := strings.Split(string(buf[:length]), "\n")
	requestData = append(requestData, lines...)

	request := parseRequestData(requestData)

	response := ResponseData{
		statusCode: "200 OK",
		headers:    make(map[string]string),
	}

	if request.endpoint == "/" {
		sendResponse(conn, request, response)
	} else if strings.HasPrefix(request.endpoint, "/echo") {
		endpointFields := strings.Split(request.endpoint, "/")
		response.headers["Content-Type"] = "text/plain"
		response.headers["Content-Length"] = strconv.Itoa(len(endpointFields[2]))
		response.body = endpointFields[2]
		sendResponse(conn, request, response)
	} else if strings.HasPrefix(request.endpoint, "/user-agent") {
		response.headers["Content-Type"] = "text/plain"
		response.headers["Content-Length"] = strconv.Itoa(len(request.headers["User-Agent"]))
		response.body = request.headers["User-Agent"]
		sendResponse(conn, request, response)
	} else if strings.HasPrefix(request.endpoint, "/files") {
		endpointFields := strings.Split(request.endpoint, "/")
		if request.method == "GET" {
			f, err := os.ReadFile(*directory + endpointFields[2])
			if err != nil {
				respondWithNotFound(conn, request, response)
				return
			}

			response.headers["Content-Type"] = "application/octet-stream"
			response.headers["Content-Length"] = strconv.Itoa(len(string(f)))
			response.body = string(f)
			sendResponse(conn, request, response)
			return
		}

		// if method is POST
		body := request.body
		file, err := os.OpenFile(*directory+endpointFields[2], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			respondWithNotFound(conn, request, response)
			return
		}
		defer file.Close()

		cleanedBody := strings.ReplaceAll(body, "\x00", "")
		cleanedBody = strings.ReplaceAll(cleanedBody, "\r", "")
		if _, err := file.WriteString(cleanedBody); err != nil {
			respondWithNotFound(conn, request, response)
			return
		}

		response.statusCode = "201 Created"
		sendResponse(conn, request, response)
	} else {
		respondWithNotFound(conn, request, response)
	}

	defer conn.Close()
}

func parseRequestData(reqData []string) RequestData {
	requestData := RequestData{
		headers: make(map[string]string),
	}
	var updateBody bool
	for i, line := range reqData {
		if line == "\r" {
			updateBody = true
		}

		if i == 0 {
			endpointFields := strings.Fields(line)
			requestData.method = strings.TrimSpace(endpointFields[0])
			requestData.endpoint = strings.TrimSpace(endpointFields[1])
			continue
		}

		if updateBody || !strings.Contains(line, ":") {

			requestData.body = requestData.body + line
		} else if strings.Contains(line, ":") && strings.HasSuffix(line, "\r") && len(strings.Split(line, ":")) == 2 {
			headerDetails := strings.Split(line, ":")
			requestData.headers[strings.TrimSpace(headerDetails[0])] = strings.TrimSpace(headerDetails[1])
		}
	}

	return requestData
}

func sendResponse(conn net.Conn, request RequestData, respData ResponseData) {

	if encodings, exists := request.headers["Accept-Encoding"]; exists {
		supportsGzip := false
		if strings.Contains(encodings, ",") {
			for _, encoding := range strings.Split(encodings, ",") {
				if strings.TrimSpace(encoding) == "gzip" {
					supportsGzip = true
					break
				}
			}
		} else {
			supportsGzip = request.headers["Accept-Encoding"] == "gzip"
		}

		if supportsGzip {
			respData.headers["Content-Encoding"] = "gzip"

			var buff bytes.Buffer
			gz := gzip.NewWriter(&buff)
			if _, err := gz.Write([]byte(respData.body)); err != nil {
				respData.body = ""
			}
			gz.Close()

			respData.headers["Content-Length"] = strconv.Itoa(len(buff.Bytes()))
			respData.body = buff.String()
		}
	}

	response := "HTTP/1.1 " + respData.statusCode + "\r\n"
	for key, value := range respData.headers {
		response += key + ": " + value + "\r\n"
	}
	response += "\r\n" + respData.body

	conn.Write([]byte(response))
}

func respondWithNotFound(conn net.Conn, request RequestData, response ResponseData) {
	response.statusCode = "404 Not Found"
	sendResponse(conn, request, response)
}
