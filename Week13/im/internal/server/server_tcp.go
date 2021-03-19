package server

import (
	"context"
	"fmt"
	"im/internal/server/conf"
	"im/pkg/common"
	"im/pkg/config"
	"im/pkg/json"
	"im/pkg/protocal"
	"net"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	glog "github.com/golang/glog"
)

const (
	maxInt = 1<<31 - 1
)

var receivedAiMsgCount uint64
var sendedAiMsgCount uint64

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP(s *Server, bind string, accept int) (err error) {
	var (
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
		glog.Errorf("net.ResolveTCPAddr(tcp, %s) error(%v)", bind, err)
		return
	}
	if listener, err = net.ListenTCP("tcp", addr); err != nil {
		glog.Errorf("net.ListenTCP(tcp, %s) error(%v)", bind, err)
		return
	}
	glog.Infof("start tcp listen: %s", bind)
	common.Println("start tcp listen:", bind)
	// split N core accept
	for i := 0; i < accept; i++ {
		go acceptTCP(s, listener)
	}
	go s.broadcaster()
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptTCP(s *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		// r    int
	)

	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			glog.Errorf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		go s.dispatchTCP(conn)
		// if r++; r == maxInt {
		// 	r = 0
		// }
	}
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (s *Server) dispatchTCP(conn *net.TCPConn) {
	// 当前连接的用户id
	user := NewUser()
	defer func() {
		// 捕获异常
		if err := recover(); err != nil {
			common.Println("dispatchTCP defer recover error:", err)
		}
		// 清除用户数据
		if user.UserID > 0 {
			s.removeUser(user.UserID, conn)
			common.Println("dispatchTCP defer conn.close, clientIP:"+conn.RemoteAddr().String(), "userID:", user.UserID)
		}
		conn.Close()
		// runtime.Goexit()
	}()

	go user.handleLoop(s, conn)
	user.readLoop(conn)

}

// auth for goim handshake with client, use rsa & aes.
func (s *Server) authTCP(ctx context.Context) (mid int64, key, rid string, accepts []int32, hb time.Duration, err error) {
	return
}

// 计算登录token
func (s *Server) getLoginToken(userID int64, time int64) string {
	return common.GetToken(conf.Conf.TCPServer.LoginKey, userID, time)
}

// 创建Api token
func (s *Server) generateToken(userID int64) string {
	return common.GetToken(conf.Conf.TCPServer.ChatKey, userID, common.GetTime())
}

// 计算gmtoken
func (s *Server) getGmToken(userID int64, time int64) string {
	return common.GetToken(conf.Conf.TCPServer.SystemKey, userID, time)
}

// 移除用户，此操作会从mapUser移除用户，并且会从所有Group中移除用户
func (s *Server) removeUser(userID int64, conn *net.TCPConn) {
	user, err := s.bucket.GetUser(userID)
	if err != nil {
		common.Println(err)
		return
	}
	// 如果取得的用户连接，和当前连接不一样，表示已经被重新登录，则直接退出，不处理别的
	if user.Conn != conn {
		return
	}

	// 将用户从所有加入的频道移除
	if len(user.GroupIDs) > 0 {
		s.mapGroup.BatchDelUser(user.GroupIDs, userID)
	}
	// 状态更改
	user.ClosedSig <- true
	user.Closed = true
	// 将用户移除mapUser
	s.bucket.DelUser(userID)

	if conf.Conf.TCPServer.Debug {
		common.Println("removeUser disconnected :", userID)
	}

}

// 用户退出登录
func (s *Server) imLogout(userID int) {
	if conf.Conf.TCPServer.Debug {
		common.Println("user logout:", userID)
	}
}

// 用户加入频道
func (s *Server) joinGroup(userID int64, body map[string]interface{}) (int, error) {
	groupID, _ := protocal.GetBodyString(body, "groupID")

	if err := CheckGroupIDValid(groupID); err != nil {
		return config.ImErrorCodeGroupID, err
	}

	// 将频道id写入用户数据
	s.bucket.JoinUserGroupID(userID, groupID)

	// 将用户数据写入group
	s.mapGroup.JoinGroup(groupID, userID)

	if conf.Conf.TCPServer.Debug {
		common.Println("User: ", userID, " joined group:", groupID)
	}

	return config.ImResponseCodeSuccess, nil
}

// 用户退出频道
func (s *Server) quitGroup(userID int64, body map[string]interface{}) (int, error) {
	groupID, _ := protocal.GetBodyString(body, "groupID")

	if err := CheckGroupIDValid(groupID); err != nil {
		return config.ImErrorCodeGroupID, err
	}

	// 删除用户所在组
	if err := s.bucket.DelUserGroupID(userID, groupID); err != nil {
		return config.ImErrorCodeQuitGroup, err
	}

	// 删除组内用户
	if err := s.mapGroup.DelGroupUserID(groupID, userID); err != nil {
		return config.ImErrorCodeQuitGroup, err
	}

	if conf.Conf.TCPServer.Debug {
		common.Println("User: ", userID, " leaved group:", groupID)
	}

	return config.ImResponseCodeSuccess, nil
}

func (s *Server) broadcaster() {
	for {
		select {
		case imPacket := <-s.globalMq:
			time.Sleep(time.Second * 5)
			dst := make([]*User, 0)
			s.bucket.mapUser.Range(func(key, value interface{}) bool {
				if !value.(*User).GmFlag {
					dst = append(dst, value.(*User))
				}
				return true
			})

			for _, v := range dst {
				v.Mq <- imPacket
			}
		}
	}
}

// 世界聊天
func (s *Server) imChatBoradcast(userID int64, fromType uint16, body map[string]interface{}) (int, error) {
	// 生成包头
	headerBytes := protocal.NewHeader(config.ImChatBoradcast, fromType)
	// 生成包体
	// 若是由用户发起的，需要在包体中注入发送者信息
	if config.ImFromTypeSytem != fromType {
		// 读取发送者信息
		senderInfo, err := s.bucket.GetUser(userID)
		if err != nil {
			return config.ImErrorCodeUserInfo, err
		}
		// 发送者信息
		// body["senderID"] = strconv.Itoa(senderInfo.UserID
		body["senderPlatformID"] = senderInfo.PlatformID
		body["senderPlatformName"] = senderInfo.PlatformName
		body["senderExtInfo"] = senderInfo.ExtInfo
	}
	// 生成完整包数据
	bodyBytes, _ := json.Encode(body)
	imPacket := protocal.NewImPacket(headerBytes, bodyBytes)

	s.globalMq <- imPacket

	if config.ImFromTypeSytem == fromType {
		// stat s统计信息-发送系统消息数
		// SysBoradcastMessageCount++
		atomic.AddUint64(&s.stat.SysBoradcastMessageCount, 1)
		if conf.Conf.TCPServer.Debug {
			common.Println("system boradcastMessage send successed:", body)
		}
	} else {
		// stat 统计信息-发送广播消息数
		// statInfo.BoradcastMessageCount++
		atomic.AddUint64(&s.stat.BoradcastMessageCount, 1)
		if conf.Conf.TCPServer.Debug {
			common.Println("boradcastMessage send successed:", body)
		}
	}

	// 压力测试输出，发送消息数量
	if fromType == config.ImFromTypeAi {
		// sendedAiMsgCount++
		idx := atomic.AddUint64(&sendedAiMsgCount, 1)
		if idx%100 == 0 {
			fmt.Println("["+common.GetTimestamp()+"]:sendedAiMsgCount:", idx)
		}
	}

	return config.ImResponseCodeSuccess, nil
}

// 频道聊天
func (s *Server) imChatGroup(userID int64, fromType uint16, body map[string]interface{}) (int, error) {
	// 获取接受消息的频道信息
	groupID, _ := protocal.GetBodyString(body, "groupID")
	group, err := s.mapGroup.Get(groupID)
	if err != nil {
		return config.ImErrorCodeGroupInfo, err
	}

	// 生成包头
	headerBytes := protocal.NewHeader(config.ImChatGroup, fromType)
	// 生成包体
	// 若是由用户发起的，需要在包体中注入发送者信息
	if config.ImFromTypeSytem != fromType {
		// 读取发送者信息
		senderInfo, err := s.bucket.GetUser(userID)
		if err != nil {
			return config.ImErrorCodeUserInfo, err
		}
		// 发送者信息
		// body["senderID"] = strconv.Itoa(senderInfo.UserID)
		body["senderPlatformID"] = senderInfo.PlatformID
		body["senderPlatformName"] = senderInfo.PlatformName
		body["senderExtInfo"] = senderInfo.ExtInfo
	}
	// 生成完整包数据
	bodyBytes, _ := json.Encode(body)
	imPacket := protocal.NewImPacket(headerBytes, bodyBytes)

	// 遍历频道所有用户，发送消息
	for _, receivedUserID := range group.UserIDs {
		receiverInfo, err := s.bucket.GetUser(receivedUserID)
		if err != nil {
			if conf.Conf.TCPServer.Debug {
				common.Println("group message Error: userID not found", receivedUserID, ". continued.")
			}
			continue
		}
		receiverInfo.Mq <- imPacket

		if conf.Conf.TCPServer.Debug {
			common.Println("sendMsg groupMessage to:[", receivedUserID, "]", body)
		}
	}

	if config.ImFromTypeSytem == fromType {
		// stat 统计信息-频道系统消息数
		s.stat.SysGroupMessageCount++

		if conf.Conf.TCPServer.Debug {
			common.Println("system group message send successed:", "groupID:", groupID, "msg:", body)
		}
	} else {
		// 统计信息-发送广播消息数
		s.stat.GroupMessageCount++
		if conf.Conf.TCPServer.Debug {
			common.Println("group message send successed:", "groupID:", groupID, "msg:", body)
		}
	}

	return config.ImResponseCodeSuccess, nil
}

// 私聊
func (s *Server) imChatPrivate(conn *net.TCPConn, userID int64, fromType uint16, body map[string]interface{}) (int, error) {
	// 读取接收者信息
	receivedUserID, _ := protocal.GetUserID(body, "receiverId")
	receiverInfo, err := s.bucket.GetUser(receivedUserID)
	if err != nil {
		// 对方不在线，给发送方发送对方不在线的notice
		return config.ImResponseCodeReceiverOffline, nil
	}

	// 生成包头
	headerBytes := protocal.NewHeader(config.ImChatPrivate, fromType)
	// 生成包体
	body["receiverID"] = strconv.FormatInt(receiverInfo.UserID, 10)
	body["receiverPlatformID"] = receiverInfo.PlatformID
	body["receiverPlatformName"] = receiverInfo.PlatformName
	body["receiverExtInfo"] = receiverInfo.ExtInfo

	// 若是由用户发起的，需要在包体中注入发送者信息
	if config.ImFromTypeSytem != fromType {
		// 读取发送者信息
		senderInfo, err := s.bucket.GetUser(userID)
		if err != nil {
			return config.ImErrorCodeUserInfo, err
		}
		// 发送者信息
		// body["senderID"] = strconv.Itoa(senderInfo.UserID)
		body["senderPlatformID"] = senderInfo.PlatformID
		body["senderPlatformName"] = senderInfo.PlatformName
		body["senderExtInfo"] = receiverInfo.ExtInfo
	}
	// 生成完整包数据
	bodyBytes, _ := json.Encode(body)
	imPacket := protocal.NewImPacket(headerBytes, bodyBytes)

	// 给接收者发送消息
	// receiverInfo.Conn.Write(imPacket.Serialize())
	receiverInfo.Mq <- imPacket

	if conf.Conf.TCPServer.Debug {
		common.Println("send private msg to [", receiverInfo.UserID, "], msg:", body)
	}

	// 若发送者为用户，需要给发送者也发送一个私聊信息
	if config.ImFromTypeUser == fromType {
		senderInfo, _ := s.bucket.GetUser(userID)
		senderInfo.Mq <- imPacket

		if conf.Conf.TCPServer.Debug {
			common.Println("send private msg to [", senderInfo.UserID, "], msg:", body)
		}
	}

	if config.ImFromTypeSytem == fromType {
		// stat 统计信息-系统私聊消息数
		s.stat.SysPrivateMessageCount++
	} else {
		// stat 统计信息-发送私聊消息数
		s.stat.PrivateMessageCount++
	}

	return config.ImResponseCodeSuccess, nil
}

// 处理服务器统计信息
func (s *Server) imStat(conn *net.TCPConn) {
	var statInfo = s.stat.Get()
	// 生成包体内容
	messageBody := make(map[string]interface{})
	messageBody["startTime"] = statInfo.StartTime
	messageBody["runTime"] = statInfo.RunTime
	connectCount := s.bucket.LenUser()
	messageBody["connectCount"] = connectCount
	messageBody["maxConnectCount"] = s.bucket.MaxOnLine
	messageBody["groupCount"] = s.mapGroup.GetOnline()
	messageBody["maxGroupCount"] = s.mapGroup.MaxOnLine
	messageBody["privateMessageCount"] = statInfo.PrivateMessageCount
	messageBody["boradcastMessageCount"] = statInfo.BoradcastMessageCount
	messageBody["groupMessageCount"] = statInfo.GroupMessageCount
	messageBody["sysBoradcastMessageCount"] = statInfo.SysBoradcastMessageCount
	messageBody["sysPrivateMessageCount"] = statInfo.SysPrivateMessageCount
	messageBody["sysGroupMessageCount"] = statInfo.SysGroupMessageCount
	if connectCount > 0 && connectCount < 1000 {
		for _, user := range s.bucket.GetMapUser() {
			common.Println("userId:", user.UserID)
		}
	}
	messageBody["svrGoroutineCount"] = runtime.NumGoroutine()
	protocal.Send(conn, config.ImStat, config.ImFromTypeSytem, messageBody)
}

// 处理判断用户是否在线
func (s *Server) imCheckOnline(conn *net.TCPConn, body map[string]interface{}) {
	userID, _ := protocal.GetUserID(body, "userID")
	// 生成包体内容
	messageBody := make(map[string]interface{})
	if s.bucket.ExistsUser(userID) {
		messageBody["onLine"] = 1
	} else {
		messageBody["onLine"] = 0
	}

	protocal.Send(conn, config.ImCheckOnline, config.ImFromTypeSytem, messageBody)
}

// 踢用户下线
func (s *Server) imKickUser(conn *net.TCPConn, body map[string]interface{}) {
	userID, _ := protocal.GetUserID(body, "userID")
	msg, _ := protocal.GetBodyString(body, "msg")
	userInfo, err := s.bucket.GetUser(userID)
	if err == nil {
		// 生成包体内容
		messageBody := make(map[string]interface{})
		messageBody["msg"] = msg
		_, err = protocal.Send(userInfo.Conn, config.ImKickUser, config.ImFromTypeSytem, messageBody)
		if err != nil {
			s.bucket.DelUser(userID)
		}
	}
}

// 踢所有用户下线
func (s *Server) imKickAll(conn *net.TCPConn, body map[string]interface{}) {
	msg, _ := protocal.GetBodyString(body, "msg")
	// 生成包体内容
	messageBody := make(map[string]interface{})
	messageBody["msg"] = msg

	s.bucket.RLockRangeUser(func(user *User) {
		protocal.Send(user.Conn, config.ImKickUser, config.ImFromTypeSytem, messageBody)
	})
}

// 获取频道人员列表
func (s *Server) imGroupUserList(conn *net.TCPConn, body map[string]interface{}) {
	groupID, _ := protocal.GetBodyString(body, "groupID")

	// 频道成员列表
	groupUserList := NewUserList()
	if group, err := s.mapGroup.Get(groupID); err == nil {
		if len(group.UserIDs) > 0 {
			for _, userID := range group.UserIDs {
				if user, err := s.bucket.GetUser(userID); err == nil {
					userInfo := make(map[string]interface{})
					// userInfo["userID"] = strconv.Itoa(userID)
					userInfo["platformId"] = user.PlatformID
					userInfo["platformName"] = user.PlatformName
					userInfo["extInfo"] = user.ExtInfo
					groupUserList = append(groupUserList, userInfo)
				}
			}
		}
	}
	if conf.Conf.TCPServer.Debug {
		common.Println(groupUserList)
	}

	// 生成包体内容
	messageBody := make(map[string]interface{})
	messageBody["userList"] = groupUserList

	// 发送协议
	protocal.Send(conn, config.ImGroupUserList, config.ImFromTypeSytem, messageBody)
}
