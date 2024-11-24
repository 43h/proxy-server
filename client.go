// 链接代理服务器
package main

import (
	"net"
)

func initClient(serverAddr string, uuid string) net.Conn {
	// Establish a non-TLS connection to the server
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		LOGE("failed to connect to server ", serverAddr, " ", err)
		return nil
	} else {
		LOGI("connected to server at", serverAddr)
	}
	go clientHandleRcv(conn, uuid)

	return conn
}

func clientHandleRcv(conn net.Conn, uuid string) {
	buf := make([]byte, 10240)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			LOGE("failed to read data from server", err)
			serverAddEventDisconnect(uuid)
			return
		} else {
			LOGI("received message from server:", n)
			serverAddEventMsg(uuid, buf[:n], n)
		}
	}
}
