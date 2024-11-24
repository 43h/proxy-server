package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const (
	MessageClassLocal      = "local"
	MessageClassUpstream   = "upstream"
	MessageClassDownstream = "downstream"
)

const (
	MessageTypeConnect    = "connect"
	MessageTypeDisconnect = "disconnect"
	MessageTypeData       = "data"
)

type Message struct {
	MessageClass string `json:"message_class"`
	MessageType  string `json:"message_type"`
	UUID         string `json:"uuid"`
	IPStr        string `json:"ip_str"`
	Length       int    `json:"length"`
	Data         []byte `json:"data"`
}

var listener net.Listener
var conn net.Conn
var status int

type ConnectionInfo struct {
	IPStr      string
	Conn       net.Conn
	Status     int
	MsgChannel chan Message //cache upstream message
	Timestamp  int64
}

const (
	Connected = iota + 1
	Disconnected
)

var messageChannel = make(chan Message, 10000)
var connections = make(map[string]ConnectionInfo)

func initServerTls() bool {
	LOGI("Starting TLS server")
	cert, err := tls.LoadX509KeyPair("test.pem", "test.key")
	if err != nil {
		LOGE("fail to load certificate, ", err)
		return false
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	tmpListener, err := tls.Listen("tcp", ConfigParam.Listen, config)
	if err != nil {
		LOGE("fail to start TLS listener, ", err)
		return false
	}
	listener = tmpListener
	return true
}

func initServer() bool {
	LOGI("Starting server without TLS")
	tmpListener, err := net.Listen("tcp", ConfigParam.Listen)
	if err != nil {
		LOGE("fail to start listener, ", err)
		return false
	}
	listener = tmpListener
	return true
}

func closeServer() {
	if listener != nil {
		err := listener.Close()
		if err != nil {
			LOGE("fail to close listener, ", err)
		} else {
			LOGI("listener closed")
		}
	} else {
		LOGI("Server closed")
	}
}

func startServer() {
	LOGI("Server started, Listening on ", ConfigParam.Listen)
	go handleEvents()
	for {
		tmpConn, err := listener.Accept()
		if err != nil {
			LOGE("fail to accepting, ", err)
			continue
		}

		if conn != nil || status == Connected {
			fmt.Println("Only one client is allowed to connect at a time")
			tmpConn.Close()
			continue
		} else {
			conn = tmpConn
			status = Connected
		}

		go rcvServer(conn)
	}
}

func rcvServer(conn net.Conn) {
	LOGI("downstream connect to upstream")

	buf := make([]byte, 10240)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			LOGE("upstream fail to read data from downstream, ", err)
			conn.Close()
			conn = nil
			status = Disconnected
			LOGI(" downstream connection closed")
			return
		} else {
			LOGI("upstream read message from downstream, length ", n)
			var msg Message
			err = json.Unmarshal(buf[:n], &msg)
			if err != nil {
				LOGE("upstream fail to unmarshaling ", err)
				continue
			} else {
				messageChannel <- msg
			}
		}
	}
}

func handleEvents() {
	for {
		select {
		case message := <-messageChannel:
			switch message.MessageClass {
			case MessageClassLocal:
				handleEventLocal(message)
			case MessageClassUpstream:
				handleEventUpstream(message)
			}
		}
	}
}

func handleEventLocal(msg Message) {
	switch msg.MessageType {
	case MessageTypeConnect: //proxy connect to remote server
		connection, exists := connections[msg.UUID]
		if exists {
			connection.Timestamp = time.Now().Unix()
			go handleClientRcv(connection.Conn, msg.UUID)
			go handleClientSnd(connection.Conn, connection.MsgChannel)
		} else {
			LOGE(msg.UUID, " fail to find connection between proxy and remote-server")
		}
	case MessageTypeDisconnect:
		delete(connections, msg.UUID)
	case MessageTypeData:
		msg.MessageClass = MessageClassDownstream
		data, err := json.Marshal(msg)
		if err != nil {
			LOGE(msg.UUID, " fail to marshaling message, ", err)
			return
		}
		_, err = conn.Write(data)
		if err != nil {
			LOGE(msg.UUID, " upstream fail to send event-data to downstream, ", err)
			return
		} else {
			LOGI(msg.UUID, " upstream sent event-data to downstream, length ", len(data))
		}
	}
}

func handleEventUpstream(msg Message) {
	switch msg.MessageType {
	case MessageTypeConnect: //connect to remote server
		connection := ConnectionInfo{
			IPStr:      msg.IPStr,
			Conn:       nil,
			Status:     Disconnected,
			MsgChannel: make(chan Message, 1000),
			Timestamp:  time.Now().Unix(),
		}
		connections[msg.UUID] = connection
		initClient(msg.IPStr, msg.UUID)
	case MessageTypeData:
		connection, exists := connections[msg.UUID]
		if exists {
			connection.MsgChannel <- msg
		} else {
			LOGE(msg.UUID, " connection not found")
		}
	default:
		LOGE("Unknown message type")
	}
}

func AddEventConnect(uuid string, conn net.Conn) {
	connection, exists := connections[uuid]
	if exists {
		connection.Conn = conn
		connection.Status = Connected
	} else {
		LOGE(uuid, " fail to find the connection")
		return
	}

	message := Message{
		MessageClass: MessageClassLocal,
		MessageType:  MessageTypeConnect,
		UUID:         uuid,
		IPStr:        "",
		Length:       0,
		Data:         nil,
	}
	messageChannel <- message
}

func AddEventDisconnect(uuid string) {
	message := Message{
		MessageClass: MessageClassLocal,
		MessageType:  MessageTypeDisconnect,
		UUID:         uuid,
		IPStr:        "",
		Length:       0,
		Data:         nil,
	}
	messageChannel <- message
}

func AddEventMsg(uuid string, buf []byte, len int) {
	message := Message{
		MessageClass: MessageClassLocal,
		MessageType:  MessageTypeData,
		UUID:         uuid,
		IPStr:        "",
		Length:       len,
		Data:         buf[:len],
	}
	messageChannel <- message
}
