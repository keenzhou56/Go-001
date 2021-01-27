package main

import (
	"Go-000/Week09/common"
	"Go-000/Week09/config"
	"Go-000/Week09/json"
	"Go-000/Week09/protocal"
	"io"
	"net"
	"runtime"
)

var debug = true

func main() {
	// 开启多核模式
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 使用TCP协议监听端口
	remote := "127.0.0.1:8081"

	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", remote)
	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)

	defer func() {
		// 捕获异常
		if err := recover(); err != nil {
			common.Println("main defer recover error:", err)
		}
		tcpListener.Close()
		common.Println("===============server closed", "==============")

	}()

	common.Println("===============start listen: "+remote, "=================")

	// 监听消息
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}
		accept := runtime.NumCPU()
		// 连接成功，开始监听消息
		for i := 0; i < accept; i++ {
			go tcpPipe(tcpConn)
		}

		// debug 调试输出
		common.Vd(debug, "New User connected:", tcpConn.RemoteAddr().String())
	}
}

// 处理客户端消息
func tcpPipe(conn *net.TCPConn) {
	// 当前连接的用户id
	defer func() {
		// 捕获异常
		if err := recover(); err != nil {
			common.Println("tcpPipe defer recover error:", err)
		}
		// 关闭连接
		conn.Close()
	}()

	for {
		// 读取包内容
		imPacket, err := protocal.ReadPacket(conn)
		if err != nil {
			if err != io.EOF {
				common.Println("ReadPacket Error:", err)
				// Error: 解析协议错误
				protocal.SendError(conn, config.IM_ERROR_CODE_PACKET_READ, err.Error())
			} else {
				common.Vd(debug, " remote:", conn.RemoteAddr().String())
			}
			return
		}

		// 相关参数读取
		// 消息类型 int
		messageType := imPacket.GetType()
		// 来源类型
		fromType := imPacket.GetFrom()
		// 包体 map[string]interface{}
		messageBody, err := json.JsonDecode(string(imPacket.GetBody()))
		if err != nil {
			// Error: 使用json解析协议包错误
			common.Println(err)
			protocal.SendError(conn, config.IM_ERROR_CODE_PACKET_BODY, err.Error())
			return
		}

		// debug 调试输出
		common.Vd(debug, "messageType:", messageType)
		common.Vd(debug, "fromType:", fromType)
		common.Vd(debug, "messageBody:", messageBody)

		// 发送登录成功的通知
		lastToken := "xxx"
		code := 200
		protocal.SendSuccess(conn, messageType, lastToken, code)

	}
}
