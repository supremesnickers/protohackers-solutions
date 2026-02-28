package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"protohackers/budget-chat/chatroom"
	"regexp"
)

// we need separate goroutines for writing and reading to/from the user connection
func handleChat(conn net.Conn, cr *chatroom.ChatRoom) {
	defer func() {
		conn.Close()
	}()

	_, err := conn.Write([]byte("Welcome to budgetchat! What shall I call you?\n"))
	if err != nil {
		log.Printf("Could not send welcome message: %e", err)
		return
	}

	reader := bufio.NewReader(conn)

	username, err := reader.ReadString(byte('\n'))
	if err != nil {
		log.Printf("Error while reading message: %e", err)
		return
	}
	username = stripNewline(username)

	// validate username, ensure length of at least 1 and only alphanumeric characters
	if len(username) < 1 || !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(username) {
		return
	}

	log.Printf("%s joined the server.", username)
	cr.Join(username, conn)
	defer cr.Leave(username)

	// receive message from user and broadcast to others
	for {
		msg, err := reader.ReadString(byte('\n'))
		if err != nil {
			log.Printf("could not read message from %s: %e", username, err)
			return
		}

		cr.Broadcast(fmt.Sprintf("[%s] %s", username, msg), username)
	}
}

func main() {
	s, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("could not listen on port 8080")
	}

	log.Printf("listening on port 8080")

	cr := chatroom.New()

	for {
		conn, err := s.Accept()
		if err != nil {
			log.Printf("could not establish connection %e", err)
		}

		go handleChat(conn, cr)
	}

}
