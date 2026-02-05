package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
)

type Request struct {
	Method string      `json:"method"`
	Number json.Number `json:"number"`
}

// func isPrime[N int | int64](inputNumber N) bool {
// 	for i := N(2); i <= (inputNumber / 2); i++ {
// 		if inputNumber%i == 0 {
// 			return false
// 		}
// 	}
// 	return true
// }

func isPrime[N int | int64](inputNumber N) bool {
	if inputNumber < 2 {
		return false
	}
	if inputNumber == 2 {
		return true
	}
	if inputNumber%2 == 0 {
		return false
	}
	for i := N(3); i*i <= inputNumber; i += 2 {
		if inputNumber%i == 0 {
			return false
		}
	}
	return true
}

func makeMalformedResponse(resp map[string]any) []byte {
	resp["method"] = "goaway"

	responseBytes := bytes.NewBuffer([]byte{})
	json.NewEncoder(responseBytes).Encode(resp)
	return responseBytes.Bytes()
}

func makeValidResponse(isPrime bool) []byte {
	resp := map[string]any{
		"method": "isPrime",
		"prime":  isPrime,
	}

	responseBytes := bytes.NewBuffer([]byte{})
	json.NewEncoder(responseBytes).Encode(resp)
	return responseBytes.Bytes()
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Handling new connection")

	for {
		buf := make([]byte, 128)
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			conn.Write(makeMalformedResponse(map[string]any{}))
			return
		}
		jsonBytes := bytes.NewReader(buf)

		var request Request
		d := json.NewDecoder(jsonBytes)
		d.UseNumber()
		decodeErr := d.Decode(&request)

		// Default response
		response := map[string]any{
			"method": "isPrime",
			"prime":  "false",
		}

		if decodeErr == nil && request.Method == "isPrime" {
			fmt.Print("Received isPrime request for ", request.Number, "\n")
			if numInt, err := request.Number.Int64(); err == nil {
				response["prime"] = isPrime(numInt)
			} else {
				conn.Write(makeMalformedResponse(response))
				return
			}
		} else {
			response["method"] = "goaway"
			conn.Write(makeMalformedResponse(response))
			return
		}

		conn.Write(makeValidResponse(response["prime"].(bool)))
	}
}

func main() {
	l, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer l.Close()

	fmt.Println("Listening on localhost:8080")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			return
		}
		go handleRequest(conn)
	}
}
