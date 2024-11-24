package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const (
	MessageTypeLocal      = "local"
	MessageTypeUpstream   = "upstream"
	MessageTypeDownstream = "downstream"
)

const (
	MessageClassConnect   = "connect"
	MessageClasDisconnect = "disconnect"
	MessageClasData       = "data"
)

type Message struct {
	MessageType  string `json:"message_type"`
	MessageClass string `json:"message_class"`
	UUID         string `json:"uuid"`
	IPStr        string `json:"ip_str"`
	Length       int    `json:"length"`
	Data         []byte `json:"data"`
}

var listener net.Listener
var conn net.Conn

type ConnectionInfo struct {
	uuid       string
	IPStr      string
	Conn       net.Conn
	Status     int
	MsgChannel chan Message
	Timestamp  int64
}

var messageChannel = make(chan Message, 10000)
var connections = make(map[string]ConnectionInfo)

func initServerTls() bool {
	cert, err := tls.LoadX509KeyPair("test.pem", "test.key")
	if err != nil {
		LOGE(" loading certificate:", err)
		return false
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	tmpListener, err := tls.Listen("tcp", ConfigParam.Listen, config)
	if err != nil {
		LOGE("starting TLS listener:", err)
		return false
	}
	listener = tmpListener
	return true
}

func initServer() bool {
	tmpListener, err := net.Listen("tcp", ConfigParam.Listen)
	if err != nil {
		LOGE("Error starting listener:", err)
		return false
	}
	listener = tmpListener
	return true
}

func closeServer() {
	if listener != nil {
		err := listener.Close()
		if err != nil {
			LOGE("Error closing listener:", err)
		} else {
			fmt.Println("Server closed")
		}
		fmt.Println("Error closing listener:", err)
	} else {
		fmt.Println("Server closed")
	}
}

func startServer() {
	fmt.Println("Server started")
	fmt.Println("Listening on ", ConfigParam.Listen)
	go handleEvents()
	for {

		tmpConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}

		if conn != nil {
			fmt.Println("Only one client is allowed to connect at a time")
			tmpConn.Close()
			continue
		} else {
			conn = tmpConn
		}

		go handleRcv(conn)
	}
}

func handleRcv(conn net.Conn) {
	// 处理连接的逻辑
	fmt.Println("New connection accepted")

	buf := make([]byte, 10240)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			LOGE("Error reading:", err)
			conn.Close()
			LOGI("Connection closed")
			conn = nil
			return
		} else {
			LOGI("Received message from downstream:", n)
			var msg Message
			err = json.Unmarshal(buf[:n], &msg)
			if err != nil {
				fmt.Println("Error unmarshaling:", err)
				continue
			}
			messageChannel <- msg
		}
	}
}

func handleEvents() {
	for {
		select {
		case message := <-messageChannel:
			switch message.MessageType {
			case MessageTypeLocal:
				handleEventLocal(message)
			case MessageTypeUpstream:
				handleEventUpstream(message)
			}
		}
	}
}

func handleEventLocal(msg Message) {
	switch msg.MessageClass {
	case MessageClasDisconnect:
		delete(connections, msg.UUID)
	case MessageClasData:
		msg.MessageType = MessageTypeDownstream
		data, err := json.Marshal(msg)
		if err != nil {
			LOGE("Error marshaling message:", err)
			return
		}
		_, err = conn.Write(data)
		if err != nil {
			LOGE("Error writing:", err)
			return
		} else {
			fmt.Println("Sent message to downstream, length:", len(data))
		}
	}
}

func handleEventUpstream(msg Message) {
	_, err := conn.Write(msg.Data)
	if err != nil {
		fmt.Println("Error writing:", err)
		return
	} else {
		fmt.Println("Sent message to downstream, length:", msg.Length)
	}
}

func handleEventConnection(msg Message) {
	conn := initClient(msg.IPStr, msg.UUID)
	if conn == nil {
		fmt.Println("Error connecting to client")
		return
	}

	connection := ConnectionInfo{
		uuid:      msg.UUID,
		IPStr:     msg.IPStr,
		Conn:      conn,
		Timestamp: time.Now().Unix(),
	}
	connections[msg.UUID] = connection
}

func handleEventDisconnect(msg Message) {

}

func handleEventMsg(msg Message) {
	if msg.MessageType == "upstream" {
		connection, exists := connections[msg.UUID]
		if !exists {
			fmt.Println("Error: connection not found")
			return
		}
		_, err := connection.Conn.Write(msg.Data)
		if err != nil {
			fmt.Println("Error writing:", err)
			return
		} else {
			fmt.Println("Sent message to server, length:", msg.Length)
		}
	} else {

	}
}

func serverAddEventDisconnect(uuid string) {
	message := Message{
		MessageType:  MessageTypeLocal,
		MessageClass: MessageClasDisconnect,
		UUID:         uuid,
		IPStr:        "",
		Length:       0,
		Data:         nil,
	}
	messageChannel <- message
}

func serverAddEventMsg(uuid string, buf []byte, len int) {
	message := Message{
		MessageType:  MessageTypeLocal,
		MessageClass: MessageClasData,
		UUID:         uuid,
		IPStr:        "",
		Length:       len,
		Data:         buf[:len],
	}
	connection, exists := connections[uuid]
	if exists {
		connection.Timestamp = time.Now().Unix()
	}
	messageChannel <- message
}
