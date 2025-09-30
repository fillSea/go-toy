package frame

import (
	"encoding/binary"
	"errors"
	"io"
)

/*
Frame定义

frameHeader + framePayload(packet)

frameHeader
	4 bytes: length 整型，帧总长度(含头及payload)

framePayload
	Packet
*/

// FramePayload packet部分
type FramePayload []byte
// StreamFrameCodec 流帧编码解码器
type StreamFrameCodec interface {
	Encode(io.Writer, FramePayload) error   // data -> frame，并写入io.Writer
	Decode(io.Reader) (FramePayload, error) // 从io.Reader中提取frame payload，并返回给上层
}

// 错误定义
var (
	ErrShortWrite = errors.New("short write")
	ErrShortRead = errors.New("short read")
)

type myFrameCodec struct{}

func NewMyFrameCodec() StreamFrameCodec {
	return &myFrameCodec{}
}

// Encode 将输入的Framepayload编码为一个Frame
func (p *myFrameCodec) Encode(w io.Writer, framePayload FramePayload) error {
	var f = framePayload
	// 计算总长度，4字节为消息长度
	var totalLen int32 = int32(len(framePayload)) + 4

	// 以大端字节序写入总长度
	err := binary.Write(w, binary.BigEndian, &totalLen)
	if err != nil {
		return err
	}

	// 写入 packet
	n, err := w.Write([]byte(f)) // write the frame payload to outbound stream
	if err != nil {
		return err
	}

	if n != len(framePayload) {
		return ErrShortWrite
	}
	return nil
}

// Decode 解码, 从io.Reader中读取一个完整Frame，并将得到的 Frame payload 解析出来并返回
func (p *myFrameCodec) Decode(r io.Reader) (FramePayload, error) {
	var totalLen int32
	// 读取 4 字节的总长度
	err := binary.Read(r, binary.BigEndian, &totalLen)
	if err != nil {
		return nil, err
	}

	// packet 缓冲区
	buf := make([]byte, totalLen - 4)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	if n != int(totalLen - 4) {
		return nil, ErrShortRead
	}

	return FramePayload(buf), nil
}