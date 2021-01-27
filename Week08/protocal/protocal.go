// 协议包处理类
// 包的格式(LTV) length type from value
package protocal

import (
	"Go-000/Week09/config"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"Go-000/Week09/json"
	"net"
	"strconv"
)

// 包长度定义
const (
	LENGTH_SIZE     = 2    // 消息包长度位所占字节数
	HEADER_SIZE     = 4    // 消息包头所占字节数
	TYPE_SIZE       = 2    // 消息协议类型所占字节数
	FROM_SIZE       = 2    // 来源类型所占字节数
	BODY_MAX_LENGTH = 2048 // 消息最大长度
)

/*
type ImPacketBody map[string]interface{}

func (body ImPacketBody) Get(key string, valueType string) (interface{}, bool) {
	val, exists := ImPacketBody[key]
	if !exists {
		return nil, false
	}

	return val.(valueType), true
}

func NewImPacketBody() ImPacketBody {
	imPacketBody := make(map[string]interface{})
	return imPacketBody
}
*/

type ImPacket struct {
	buff []byte
}

func (this *ImPacket) Serialize() []byte {
	return this.buff
}

// 获取消息长度，前2字节
func (this *ImPacket) GetLength() uint16 {
	return binary.BigEndian.Uint16(this.buff[0:LENGTH_SIZE])
}

// 获取消息类型，第5-6位置
func (this *ImPacket) GetType() uint16 {
	from := LENGTH_SIZE
	to := from + TYPE_SIZE
	return binary.BigEndian.Uint16(this.buff[from:to])
}

// 获取来源类型
func (this *ImPacket) GetFrom() uint16 {
	from := LENGTH_SIZE + TYPE_SIZE
	to := from + FROM_SIZE
	return binary.BigEndian.Uint16(this.buff[from:to])
}

// 获取包头，包头包括消息类型、发送者id、接受者id
func (this *ImPacket) GetHeader() []byte {
	from := LENGTH_SIZE
	to := LENGTH_SIZE + HEADER_SIZE
	return this.buff[from:to]
}

// 获取包内容
func (this *ImPacket) GetBody() []byte {
	from := LENGTH_SIZE + HEADER_SIZE
	return this.buff[from:]
}

// 生成一个包头
func NewHeader(imType uint16, fromType uint16) []byte {
	headerBytes := make([]byte, HEADER_SIZE)
	// 消息类型
	binary.BigEndian.PutUint16(headerBytes[0:TYPE_SIZE], imType)
	// 来源类型
	binary.BigEndian.PutUint16(headerBytes[TYPE_SIZE:], fromType)

	return headerBytes
}

// 生成一条消息
func NewImPacket(header []byte, body []byte) *ImPacket {
	p := &ImPacket{}

	p.buff = make([]byte, LENGTH_SIZE+HEADER_SIZE+len(body))
	binary.BigEndian.PutUint16(p.buff[0:LENGTH_SIZE], HEADER_SIZE+uint16(len(body))) // 包头长度 + 协议内容长度

	copy(p.buff[LENGTH_SIZE:LENGTH_SIZE+HEADER_SIZE], header)
	copy(p.buff[LENGTH_SIZE+HEADER_SIZE:], body)

	return p
}

// 读取一条消息
func ReadPacket(conn *net.TCPConn) (*ImPacket, error) {
	var (
		lengthBytes []byte = make([]byte, LENGTH_SIZE)
		headerBytes []byte = make([]byte, HEADER_SIZE)
		length      uint16
	)

	// read length
	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		if err == io.EOF {
			return nil, err
		} else {
			return nil, errors.New(fmt.Sprintf("Error: read packet length: %s", err.Error()))
		}
	}

	// 包内容的长度最长2048
	length = binary.BigEndian.Uint16(lengthBytes)
	lengthErr := 0
	if length > BODY_MAX_LENGTH+HEADER_SIZE {
		// debug
		lengthErr = 1
		// fmt.Println(fmt.Sprintf("Error: the size of packet is exceeded the limit:%d, given:%d", BODY_MAX_LENGTH, length))
	}

	// read header
	if _, err := io.ReadFull(conn, headerBytes); err != nil {
		return nil, errors.New(fmt.Sprintf("Error: read packet header: %s", err.Error()))
	}

	// read body
	// 扣除包头的长度
	bodyBytes := make([]byte, length-HEADER_SIZE)
	if _, err := io.ReadFull(conn, bodyBytes); err != nil {
		return nil, errors.New(fmt.Sprintf("Error: read packet body: %s", err.Error()))
	}

	// debug
	if lengthErr == 1 {
		fmt.Println(bodyBytes)
		fmt.Println(string(bodyBytes))

		return nil, errors.New(fmt.Sprintf("Error: the size of packet is exceeded the limit:%d, given:%d", BODY_MAX_LENGTH, length))
	}

	return NewImPacket(headerBytes, bodyBytes), nil
}

// 给客户端发送一个错误
func SendError(conn *net.TCPConn, errorCode int, errorMsg string) *ImPacket {
	// 生成协议内容
	body := make(map[string]interface{})
	body["code"] = errorCode
	body["msg"] = errorMsg

	// 发送消息
	imPacket := Send(conn, config.IM_ERROR, config.IM_FROM_TYPE_SYSTEM, body)

	return imPacket
}

// 给客户端发送一个成功的response
func SendSuccess(conn *net.TCPConn, imType uint16, token string, responseCode int) *ImPacket {
	// 生成协议内容
	body := make(map[string]interface{})
	body["imType"] = imType
	body["token"] = token
	body["code"] = responseCode

	// 发送消息
	imPacket := Send(conn, config.IM_RESPONSE, config.IM_FROM_TYPE_SYSTEM, body)

	return imPacket
}

// 给客户端发送一个成功的response，不同的是，可以支持客户端扩展参数
func SendSuccessWithExtra(conn *net.TCPConn, imType uint16, token string, responseCode int, extra map[string]interface{}) *ImPacket {
	// 生成协议内容
	body := make(map[string]interface{})
	body["imType"] = imType
	body["token"] = token
	body["code"] = responseCode

	// 附加扩展参数
	if len(extra) > 0 {
		for key, value := range extra {
			body[key] = value
		}
	}

	// 发送消息
	imPacket := Send(conn, config.IM_RESPONSE, config.IM_FROM_TYPE_SYSTEM, body)

	return imPacket
}

// 发送消息封装
func Send(conn *net.TCPConn, imType uint16, fromType uint16, body map[string]interface{}) *ImPacket {
	// 生成协议头
	headerBytes := NewHeader(imType, fromType)
	// 生成协议内容
	bodyBytes, _ := json.JsonEncode(body)

	// 生成完整包数据
	imPacket := NewImPacket(headerBytes, bodyBytes)

	// 发送消息
	conn.Write(imPacket.Serialize())

	return imPacket
}

// 获取body中的int值
func GetBodyUint16(body map[string]interface{}, key string) (uint16, bool) {
	val, exists := body[key]
	if !exists {
		return 0, false
	} else {
		return uint16(val.(float64)), true
	}
}

// 获取body中的int值
func GetBodyInt(body map[string]interface{}, key string) (int, bool) {
	val, exists := body[key]
	if !exists {
		return 0, false
	} else {
		return int(val.(float64)), true
	}
}

// 获取body中的string值
func GetBodyString(body map[string]interface{}, key string) (string, bool) {
	val, exists := body[key]
	if !exists {
		return "", false
	} else {
		return val.(string), true
	}
}

// 获取userId
func GetUserId(body map[string]interface{}, key string) (int, bool) {
	userIdStr, exists := GetBodyString(body, key)
	if !exists {
		return 0, false
	} else {
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			return 0, false
		} else {
			return userId, true
		}
	}
}
