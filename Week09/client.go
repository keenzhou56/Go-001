package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	remote := "127.0.0.1:8081"
	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", remote)
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	defer conn.Close()
	fmt.Println("connected!")

	go onMessageRecived(conn)

	inputReader := bufio.NewReader(os.Stdin)
	for {
		// 读取用户输入
		input, _ := inputReader.ReadString('\n')
		inputInfo := strings.Trim(input, "\r\n")
		// 如果输入quit就退出
		if strings.ToUpper(inputInfo) == "quit" {
			return
		}
		// 发送数据
		conn.Write([]byte(inputInfo))
	}

}

func onMessageRecived(conn *net.TCPConn) {
	buf := [1024]byte{}
	n, err := conn.Read(buf[:])
	if err != nil {
		fmt.Println("recv failed, err:", err)
		return
	}
	fmt.Println(string(buf[:n]))
}
