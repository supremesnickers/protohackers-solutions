package chatroom

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
)

type ChatRoom struct {
	mu      sync.Mutex
	members map[string]net.Conn
}

// Join adds a new user identified by username and TCP connection to the ChatRoom
func (c *ChatRoom) Join(username string, conn net.Conn) error {
	if _, ok := c.members[username]; ok {
		return fmt.Errorf("username %s already exists", username)
	}

	// broadcast new username to all other users
	c.Broadcast(fmt.Sprintf("* %s has joined the room\n", username), username)

	// list all present users in the room
	var memberNames []string
	for key := range c.members {
		memberNames = append(memberNames, key)
	}
	memberNamesJoined := strings.Join(memberNames, ", ")
	fmt.Fprintf(conn, "* The room contains: %s\n", memberNamesJoined)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.members[username] = conn

	return nil
}

func (c *ChatRoom) Leave(username string) error {
	if _, ok := c.members[username]; !ok {
		return fmt.Errorf("username %s does not exist", username)
	}

	delete(c.members, username)

	err := c.Broadcast(fmt.Sprintf("* %s has left the room", username), username)

	return err
}

// Broadcast a message msg to all members in the ChatRoom, except sender
func (c *ChatRoom) Broadcast(msg string, sender string) error {
	var err error

	for username, conn := range c.members {
		if username == sender {
			continue
		}
		_, e := conn.Write([]byte(msg))
		if e != nil {
			err = errors.Join(err, e)
		}
	}

	return err
}

// Creates a new ChatRoom
func New() *ChatRoom {
	return &ChatRoom{
		members: make(map[string]net.Conn),
	}
}
