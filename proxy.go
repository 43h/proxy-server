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
		LOGE(uuid, " fail to connect to remote-server ", serverAddr, " ", err)
	} else {
		AddEventConnect(uuid, conn)
		LOGI(uuid, " connect to remote-server ", serverAddr)
	}
}

func handleClientRcv(conn net.Conn, uuid string) {
	buf := make([]byte, 10240)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			LOGE(uuid, " fail to read data from remote-server, ", err)
			AddEventDisconnect(uuid)
			return
		} else {
			LOGI(uuid, " read data from remote-server, length:", n)
			AddEventMsg(uuid, buf[:n], n)
		}
	}
}

func handleClientSnd(conn net.Conn, messageChannel chan Message) {
	for {
		select {
		case message := <-messageChannel:
			_, err := conn.Write(message.Data)
			if err != nil {
				LOGE(message.UUID, " fail to send data to remote-server, ", err)
				return
			} else {
				LOGI(message.UUID, " sent data to remote-server, length:", message.Length)
			}
		}
	}
}
