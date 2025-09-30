package packet

import (
	"bytes"
	"fmt"
	"sync"
)

/*
### packet header
1 byte: commandID

### submit packet

8字节 ID 字符串
任意字节 payload

### submit ack packet

8字节 ID 字符串
1字节 result
*/

// commandID 类型
const (
	CommandConn   = iota + 0x01 // 0x01, 连接请求
	CommandSubmit               // 0x02, 消息请求
)

const (
	CommandConnAck   = iota + 0x81 // 0x81, 连接响应
	CommandSubmitAck               //0x82, 消息响应
)

// Packet 接口定义了消息的编码和解码方法
type Packet interface {
	Decode([]byte) error     // []byte -> struct
	Encode() ([]byte, error) //  struct -> []byte
}

// Submit 消息结构体
// 包含消息ID和负载数据
type Submit struct {
	ID      string
	Payload []byte
}

// Decode 解码提交消息
// 将输入的字节数组解码为Submit结构体
func (s *Submit) Decode(pktBody []byte) error {
	s.ID = string(pktBody[:8])
	s.Payload = pktBody[8:]
	return nil
}

// Encode 编码提交消息
// 将Submit结构体编码为字节数组
func (s *Submit) Encode() ([]byte, error) {
	return bytes.Join([][]byte{[]byte(s.ID[:8]), s.Payload}, nil), nil
}

// SubmitAck 消息结构体
// 包含消息ID和响应结果
type SubmitAck struct {
	ID     string
	Result uint8
}

// Decode 解码提交确认消息
// 将输入的字节数组解码为SubmitAck结构体
func (s *SubmitAck) Decode(pktBody []byte) error {
	s.ID = string(pktBody[:8])
	s.Result = uint8(pktBody[8])
	return nil
}

// Encode 编码提交确认消息
// 将SubmitAck结构体编码为字节数组
func (s *SubmitAck) Encode() ([]byte, error) {
	return bytes.Join([][]byte{[]byte(s.ID[:8]), {s.Result}}, nil), nil
}

// SubmitPool 提交消息池
// 用于复用Submit结构体，避免频繁分配内存
var SubmitPool = sync.Pool{
	New: func() interface{} {
		return &Submit{}
	},
}

// Decode 解码消息
// 根据commandID将输入的字节数组解码为对应的消息结构体
func Decode(packet []byte) (Packet, error) {
	commandID := packet[0]
	pktBody := packet[1:]

	switch commandID {
	case CommandConn:
		return nil, nil
	case CommandConnAck:
		return nil, nil
	case CommandSubmit:
		s := SubmitPool.Get().(*Submit) // 从SubmitPool池中获取一个Submit内存对象
		err := s.Decode(pktBody)
		if err != nil {
			return nil, err
		}
		return s, nil
	case CommandSubmitAck:
		s := SubmitAck{}
		err := s.Decode(pktBody)
		if err != nil {
			return nil, err
		}
		return &s, nil
	default:
		return nil, fmt.Errorf("unknown commandID [%d]", commandID)
	}
}

// Encode 编码消息
// 根据消息类型将消息结构体编码为字节数组
func Encode(p Packet) ([]byte, error) {
	var commandID uint8
	var pktBody []byte
	var err error

	switch t := p.(type) {
	case *Submit:
		commandID = CommandSubmit
		pktBody, err = p.Encode()
		if err != nil {
			return nil, err
		}
	case *SubmitAck:
		commandID = CommandSubmitAck
		pktBody, err = p.Encode()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown type [%s]", t)
	}
	return bytes.Join([][]byte{{commandID}, pktBody}, nil), nil
}