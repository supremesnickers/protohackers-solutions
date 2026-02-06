package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

type TimestampedPrice struct {
	timestamp int32
	price     int32
}

const PORT = 8080

var prices = make(map[net.Conn][]*TimestampedPrice)

func readNums(buf []byte) (int32, int32, error) {
	num1, err := binary.ReadVarint(bytes.NewReader(buf[1:5]))
	if err != nil {
		return 0, 0, err
	}
	num2, err := binary.ReadVarint(bytes.NewReader(buf[5:9]))
	if err != nil {
		return 0, 0, err
	}

	return int32(num1), int32(num2), nil
}

func handleRequest(conn net.Conn) {
	log.Println("handling new connection")
	defer conn.Close()
	for {
		buf := make([]byte, 9) // size of message is always 9 bytes
		_, err := conn.Read(buf)
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}

		typeChar := buf[0]

		if typeChar == byte('I') {
			timestamp, price, err := readNums(buf)
			if err != nil {
				log.Printf("error while decoding numbers: %s", err)
				return
			}

			tp := &TimestampedPrice{timestamp: timestamp, price: price}

			prices[conn] = append(prices[conn], tp)
		} else if typeChar == byte('Q') {
			mintime, maxtime, err := readNums(buf)
			if err != nil {
				log.Printf("error while decoding numbers: %s", err)
				return
			}

			var mean int32
			n := 0

			if mintime > maxtime {
				mean = 0
			} else {
				for i := range len(prices[conn]) {
					tp := prices[conn][i]

					if tp.timestamp >= mintime && tp.timestamp <= maxtime {
						mean += tp.price
						n++
					}
				}

				mean /= int32(n)
			}

			response := make([]byte, 4)
			binary.Encode(response, binary.BigEndian, mean)
			conn.Write(response)
		} else {
			return
		}
	}
}

func main() {
	s, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	log.Println("listening on port", PORT)
	defer s.Close()

	for {
		conn, err := s.Accept()

		if err != nil {
			log.Printf("Error accepting: %s", err)
		}

		go handleRequest(conn)
	}
}
