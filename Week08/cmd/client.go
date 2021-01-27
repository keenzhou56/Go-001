package main

import (
	"Go-000/Week09/common"
	"Go-000/Week09/config"
	"Go-000/Week09/json"
	"Go-000/Week09/protocal"
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var quitSemaphore chan bool
var debug = true
func main() {

	remote := "127.0.0.1:8081"
	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", remote)
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	defer conn.Close()
	fmt.Println("connected!")

	go onMessageRecived(conn)

	// 控制台聊天功能加入
	for {
		var msg string

		msgReader := bufio.NewReader(os.Stdin)
		msg, _ = msgReader.ReadString('\n')
		msg = strings.TrimSuffix(msg, "\n")

		if msg == "quit" {
			// logout
			break
		}

		// 分隔消息
		receiveMessageSplit := strings.SplitN(string(msg), "|", 5)
		if len(receiveMessageSplit) < 2 {
			fmt.Println("输入错误，不足2个参数，请重新输入")
			continue
		}

		// 生成包头，2个字节
		imType, _ := strconv.Atoi(receiveMessageSplit[0])
		messageType := uint16(imType)
		// 生成包体
		messageBody := make(map[string]interface{})

		messageBody["msg"] = string(msg)

		common.Vd(debug, messageBody)
		protocal.Send(conn, messageType, config.IM_FROM_TYPE_USER, messageBody)
	}
}

func onMessageRecived(conn *net.TCPConn) {
	for {
		imPacket, err := protocal.ReadPacket(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Disconnected")
				os.Exit(0)
			} else {
				fmt.Println("ReadPacket Error:", err)
			}
			return
		}

		// 调试输出，显示收到的二进制消息
		// 相关参数读取
		// 消息类型
		messageType := imPacket.GetType()

		common.Vd(debug, "message type:", messageType)
		// 包体 map[string]interface{}
		messageBody, err := json.JsonDecode(string(imPacket.GetBody()))
		if err != nil {
			fmt.Println("imPacket JsonDecode error.")
			return
		}
		common.Vd(debug, "message body:", messageBody)
		fmt.Println("cannot supported messageType:", messageType)

	}
	quitSemaphore <- true
}