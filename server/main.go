package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	connW, err := net.Dial("udp", "255.255.255.255:8086")
	if err != nil {
		fmt.Println("链接错误:", err.Error())
		os.Exit(1)
	}
	connR, err2 := net.ListenPacket("udp", ":8085")
	if err2 != nil {
		fmt.Println("链接错误:", err2.Error())
		os.Exit(1)
	}
	defer connW.Close()
	defer connR.Close()
	wg.Add(2)
	go handleRequest(connR, wg)
	go handleSend(connW, wg)
	wg.Wait()
}
func handleRequest(conn net.PacketConn, wg sync.WaitGroup) {
	defer wg.Done()
	for {
		buffer := make([]byte, 1024)
		conn.ReadFrom(buffer)
		buffer = bytes.TrimRight(buffer, "\x00")
		fmt.Println("受到消息:", string(buffer))
	}
}

func handleSend(conn net.Conn, wg sync.WaitGroup) {
	defer wg.Done()
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		fmt.Fprintf(conn, input+"\n")
	}
}
