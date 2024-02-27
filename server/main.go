package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// 创建UDP连接
	udpAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:8888") // 广播地址
	if err != nil {
		fmt.Println("Error resolving UDP address:", err.Error())
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error connecting to UDP address:", err.Error())
		return
	}
	defer conn.Close()

	// 发送广播消息
	message := []byte("Requesting IP Address")
	_, err = conn.Write(message)
	if err != nil {
		fmt.Println("Error sending broadcast message:", err.Error())
		return
	}

	fmt.Println("Broadcast message sent.")

	// 接收回复
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // 设置读取超时时间
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("Error receiving response:", err.Error())
		return
	}

	fmt.Println("Response received:", string(buf[:n]))
}
