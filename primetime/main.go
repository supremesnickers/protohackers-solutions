package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net"
)

// slow implementation of isPrime, O(n)
// func isPrime[N int | int64](inputNumber N) bool {
// 	for i := N(2); i <= (inputNumber / 2); i++ {
// 		if inputNumber%i == 0 {
// 			return false
// 		}
// 	}
// 	return true
// }

// faster implementation of isPrime, O(sqrt(n))
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
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("[%s] New connection", remoteAddr)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		log.Printf("[%s] Received: %s", remoteAddr, string(line))

		var request map[string]any
		d := json.NewDecoder(bytes.NewReader(line))
		d.UseNumber()
		if err := d.Decode(&request); err != nil {
			log.Printf("[%s] JSON decode error: %v", remoteAddr, err)
			conn.Write(makeMalformedResponse(map[string]any{}))
			return
		}

		log.Printf("[%s] Parsed request: %+v", remoteAddr, request)

		// Validate method
		if request["method"] != "isPrime" {
			log.Printf("[%s] Invalid method: %v", remoteAddr, request["method"])
			conn.Write(makeMalformedResponse(map[string]any{}))
			return
		}

		// Validate number field exists and is a json.Number
		numVal, ok := request["number"].(json.Number)
		if !ok {
			log.Printf("[%s] Invalid number field: %v (type %T)", remoteAddr, request["number"], request["number"])
			conn.Write(makeMalformedResponse(map[string]any{}))
			return
		}

		// Try to parse as int64 first
		if numInt, err := numVal.Int64(); err == nil {
			result := isPrime(numInt)
			log.Printf("[%s] isPrime(%d) = %v", remoteAddr, numInt, result)
			conn.Write(makeValidResponse(result))
		} else {
			// It's a float (or too large) - floats are never prime
			log.Printf("[%s] Number %s is not an integer, returning false", remoteAddr, numVal.String())
			conn.Write(makeValidResponse(false))
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[%s] Scanner error: %v", remoteAddr, err)
	}
	log.Printf("[%s] Connection closed", remoteAddr)
}

func main() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer l.Close()

	log.Println("Listening on port 8080")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting: %v", err)
			continue
		}
		go handleRequest(conn)
	}
}
