package main

import (
	"net"
)

func initClient(serverAddr string, uuid string) {
	go connectToRemoteServer(serverAddr, uuid) //use goroutine to connect to server
}

func connectToRemoteServer(serverAddr string, uuid string) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		AddEventDisconnect(uuid)
		LOGE(uuid, " proxy fail to connect to remote-server ", serverAddr, " ", err)
	} else {
		AddEventConnect(uuid, conn)
		LOGI(uuid, " proxy connect to remote-server ", serverAddr)
	}
}

func handleClientRcv(conn net.Conn, uuid string) {
	buf := make([]byte, 10240)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			LOGE(uuid, " proxy<---server, read, ", err)
			AddEventDisconnect(uuid)
			return
		} else {
			LOGI(uuid, "proxy<---server, read: ", n)
			AddEventMsg(uuid, buf[:n], n)
		}
	}
}

func handleClientSnd(conn net.Conn, messageChannel chan Message) {
	for {
		select {
		case message := <-messageChannel:
			n, err := conn.Write(message.Data)
			if err != nil {
				LOGE(message.UUID, " proxy--->server, send, fail, ", err)
				return
			} else {
				LOGI(message.UUID, " proxy--->server, send, need: ", message.Length, "send: ", n)
			}
		}
	}
}
