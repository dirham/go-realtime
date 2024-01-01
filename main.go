package main

import (
	"log"
	"net"
)

const PORT = ":8080"

// Define Enums
type MessageType int

const (
	NewClient MessageType = iota + 1
	NewMessage
	DeleteMessage
)

// Message type
type Message struct {
	Type MessageType
	Conn net.Conn
	Text string
}

// server: received channel Message
func server(messages chan Message) {
	conns := make(map[string]net.Conn)

	// waiting for incoming message from handler
	for {
		// received messages channel
		msg := <-messages
		switch msg.Type {
		case NewClient:
			log.Printf("New Client connected: %s", msg.Conn.RemoteAddr().String())
			conns[msg.Conn.RemoteAddr().String()] = msg.Conn
		case DeleteMessage:
			delete(conns, msg.Conn.LocalAddr().String())
		case NewMessage:
			for _, conn := range conns {
				if conn.RemoteAddr().String() != msg.Conn.RemoteAddr().String() {
					log.Printf("Recived new message from %s:%s\n", conn.RemoteAddr().String(), msg.Text)
					conn.Write([]byte(msg.Text))
				}
			}
		}
	}
}

// handle incoming connections and message from client
func handler(conn net.Conn, outgoing chan Message) {
	buffer := make([]byte, 512)
	// waiting for incomming connection and message from client
	for {
		// read message from connection and pass it to buffer
		n, err := conn.Read(buffer)
		if err != nil {
			conn.Close()
			outgoing <- Message{
				Type: DeleteMessage,
				Conn: conn,
			}
			return
		}
		outgoing <- Message{
			Type: NewMessage,
			Text: string(buffer[0:n]),
			Conn: conn,
		}
	}
}

func main() {
	c, err := net.Listen("tcp", PORT)

	if err != nil {
		log.Fatalf("Can not create connection %s\n", err.Error())
	}

	log.Printf("Ready to receive to connections at port %s\n", PORT)
	// create chat channel
	chat := make(chan Message)

	// spawn new goroutine for server
	go server(chat)

	// waiting for tcp connection and upgrade to websocket using accept
	for {
		conn, err := c.Accept()
		if err != nil {
			log.Printf("Can't Accept connection %s\n", conn.RemoteAddr())
		}

		// when connection is received register the client to message struct
		chat <- Message{
			Type: NewClient,
			Conn: conn,
		}

		go handler(conn, chat)
	}
}
