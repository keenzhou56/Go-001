package config

// 消息头预定义常量
const (
	// 自定义信息长度限制
	IM_REGISTER_EXT_INFO_LENGTH_LIMIT = 1048

	// 来源类型
	IM_FROM_TYPE_USER   = uint16(0) // 用户
	IM_FROM_TYPE_SYSTEM = uint16(1) // 系统
	IM_FROM_TYPE_AI     = uint16(2) // 机器人

	IM_ERROR    = uint16(1) // 给client发送一条错误消息，一般来说，client收到IM_ERROR，都需要断开当前连接并重新连接，如果是重复登录，则只断开、不重连
	IM_RESPONSE = uint16(2) // 返回消息给client

	// 协议类型
	IM_STAT            = uint16(101) // 统计服务器状态
	IM_CHECK_ONLINE    = uint16(102) // 判断用户是否在线
	IM_KICK_USER       = uint16(103) // 踢某用户下线
	IM_KICK_ALL        = uint16(104) // 踢所有用户下线
	IM_GROUP_USER_LIST = uint16(105) // 获取频道用户列表

	IM_LOGIN             = uint16(201) // 登录
	IM_LOGOUT            = uint16(202) // 退出
	IM_REGISTER_EXT_INFO = uint16(203) // 注册附加信息
	IM_JOIN_GROUP        = uint16(301) // 加入频道
	IM_QUIT_GROUP        = uint16(302) // 退出频道
	IM_CHAT_BORADCAST    = uint16(401) // 世界聊天
	IM_CHAT_GROUP        = uint16(402) // 频道聊天
	IM_CHAT_PRIVATE      = uint16(403) // 私聊

	// 通知类型
	IM_NOTICE_CHAT = 0 // 聊天

	// 错误消息
	IM_ERROR_CODE_RELOGIN                 = 1  // 重复登录
	IM_ERROR_CODE_NO_LOGIN                = 2  // 未登录
	IM_ERROR_CODE_PACKET_READ             = 3  // 读取协议包错误
	IM_ERROR_CODE_PACKET_BODY             = 4  // 解析协议包内容错误
	IM_ERROR_CODE_NOT_ALLOWED_IMTYPE      = 5  // 没有权限发送协议
	IM_ERROR_CODE_PRIVATE_KEY_NOT_MATCHED = 6  // 私钥不匹配
	IM_ERROR_CODE_LOGIN_TOKEN_NOT_MATCHED = 7  // 登录token不匹配
	IM_ERROR_CODE_TOKEN_NOT_MATCHED       = 8  // 消息token不匹配
	IM_ERROR_CODE_MSG_EMPTY               = 9  // 聊天内容为空
	IM_ERROR_CODE_USERID                  = 10 // 用户id错误，小于=0
	IM_ERROR_CODE_PLATFORMID              = 11 // 平台id错误
	IM_ERROR_CODE_PLATFORMNAME            = 12 // 平台名称错误
	IM_ERROR_CODE_GROUPID                 = 13 // 频道id错误，小于=0
	IM_ERROR_CODE_USER_INFO               = 14 // 读取用户登录信息错误
	IM_ERROR_CODE_GROUP_INFO              = 15 // 读取频道消息错误
	IM_ERROR_CODE_EXT_INFO_LENGTH         = 16 // 附加信息长度超出限制

	// 系统返回类型、默认是0
	IM_RESPONSE_CODE_SUCCESS          = 0 // 默认code值
	IM_RESPONSE_CODE_RECEIVER_OFFLINE = 1 // 私聊对象不在线

	GROUP_ID_LENGTH_MIN = 1   // 组队id长度最小值
	GROUP_ID_LENGTH_MAX = 100 // 组队id长度最大值
)
