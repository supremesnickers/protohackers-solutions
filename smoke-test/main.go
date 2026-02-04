package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")

	if err != nil {
		panic(err)
	}

	defer ln.Close()

	fmt.Println("Server is listening on port 8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go func(c net.Conn) {
			defer c.Close()
			b, err := io.ReadAll(c)

			if err != nil {
				return
			}

			fmt.Println(string(b))

			c.Write(b)
		}(conn)
	}
}
