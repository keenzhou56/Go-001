package server

import (
	"im/pkg/common"
	"im/pkg/config"
	"im/pkg/json"
	"im/pkg/protocal"
	"io"
	"math/rand"
	"net"
	"reflect"
	"runtime"
	"time"
)

// User 用户信息
type User struct {
	UserID         int64
	PlatformID     string
	PlatformName   string
	GroupIDs       map[string]string
	ExtInfo        string
	LastActionTime int64
	Conn           *net.TCPConn
	inChan         chan *protocal.ImPacket
	Mq             chan *protocal.ImPacket
	ClosedSig      chan bool
	Closed         bool
	LastHbTime     time.Time
	LastToken      string
	GmFlag         bool
	// ctx            context.Context
	// cancel         context.CancelFunc
}

// UserList ...
type UserList []interface{}

// NewUser 生成新用户
func NewUser() *User {
	user := &User{}
	user.GroupIDs = make(map[string]string)
	user.Mq = make(chan *protocal.ImPacket, 1024)
	user.inChan = make(chan *protocal.ImPacket, 1024)
	user.ClosedSig = make(chan bool, 1)
	user.Closed = false
	user.UserID = 0
	user.GmFlag = false
	return user
}

// NewUserList ...
func NewUserList() UserList {
	userList := make([]interface{}, 0)
	return userList
}

// SendMessage 检查用户消息队列，并给用户发消息
func (user *User) SendMessage() {
	user.LastHbTime = time.Now()
	var serverHeartbeat = user.RandServerHearbeat()
	for {

		select {
		case imPacket := <-user.Mq:
			if user.Closed {
				goto user_quit
			}
			_, err := user.Conn.Write(imPacket.Serialize())
			if err != nil {
				common.Println("SendMessage.serverHeartbeat conn.Write userId:", user.UserID)
				goto user_quit
			}
			user.LastHbTime = time.Now()

		case close := <-user.ClosedSig:
			if close {
				goto user_quit
			}

		case <-time.After(time.Second * 300): // 5分钟检测一次心跳
			if user.Closed {
				common.Println("SendMessage.user.Closed userId:", user.UserID)
				goto user_quit
			}
			if time.Now().Sub(user.LastHbTime) > serverHeartbeat {
				common.Println("SendMessage.user.lastHb userId:", user.UserID)
				goto user_quit
			}

		}
	}

user_quit:
	user.Closed = true
	user.inChan <- user.ImQuitUserMqPacket()
	runtime.Goexit()

}

// RandServerHearbeat ...
func (user *User) RandServerHearbeat() time.Duration {
	return (minServerHeartbeat + time.Duration(rand.Int63n(int64(maxServerHeartbeat-minServerHeartbeat))))
}

func (user *User) readLoop(conn *net.TCPConn) error {
	for {
		// 读取包内容
		imPacket, err := protocal.ReadPacket(conn)
		if err != nil {
			if err != io.EOF {
				// Error: 解析协议错误
				protocal.SendError(conn, config.ImErrorCodePacketRead, err.Error())
			}
			common.Println("ReadPacket Error:", err)
			return err
		}
		if user.Closed == true {
			return nil
		}
		user.inChan <- imPacket
		user.LastHbTime = time.Now()
	}
}

func (user *User) handleLoop(s *Server, conn *net.TCPConn) {
	// var (
	// 	autoID int64
	// )

	for {
		select {
		case imPacket := <-user.inChan:
			// 消息类型
			messageType := imPacket.GetType()
			// 来源类型
			fromType := imPacket.GetFrom()

			// 退出处理
			if messageType == config.ImQuitUserMq {
				common.Println("ImQuitUserMq, handleLoop quit")
				goto handleLoopQuit
			}
			// 用户主动退出
			if messageType == config.ImLogout {
				goto handleLoopQuit
			}

			if user.UserID > 0 && messageType == config.ImLogin {
				// 重复登入消息，强制退出
				common.Println("Repeat login, handleLoop quit")
				goto handleLoopQuit
			} else if user.UserID < 1 && messageType != config.ImLogin {
				// 未发登入消息，不能发其他消息
				common.Println("No login, handleLoop quit")
				goto handleLoopQuit
			}

			// 预处理如果是gm协议，必须验证user.GmFlag 或 来源IP
			if messageType != config.ImLogin && messageType > 100 && messageType < 200 && user.GmFlag != true {
				common.Println("No gm user, handleLoop quit")
				goto handleLoopQuit
			}

			// 内容分发
			handlerFuncName := s.FindRouter(messageType)
			if handlerFuncName == "" {
				common.Println("Unknown type, handleLoop quit")
				goto handleLoopQuit
			}
			ctx := NewContext()
			ctx.user = user
			ctx.conn = conn
			ctx.messageType = messageType
			ctx.fromType = fromType
			ctx.body = imPacket.GetBody()
			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf(ctx)
			values := reflect.ValueOf(s).MethodByName(handlerFuncName).Call(in)
			// 返回结果为数组，最后一个元素为error
			len := len(values)
			if values[len-1].Interface() != nil {
				errMsg := values[len-1].Interface().(error).Error()
				common.Println(values[len-1].Interface().(error).Error())
				protocal.SendError(conn, config.ImErrorCodePacketRead, errMsg)
				goto handleLoopQuit
			}

		}
	}
handleLoopQuit:
	user.Closed = true
	runtime.Goexit()
}

func (user *User) ImQuitUserMqPacket() *protocal.ImPacket {
	headerBytes := protocal.NewHeader(config.ImQuitUserMq, config.ImFromTypeSytem)
	body := make(map[string]interface{})
	body["inChan"] = "quit"
	bodyBytes, _ := json.Encode(body)
	imPacket := protocal.NewImPacket(headerBytes, bodyBytes)
	return imPacket
}
