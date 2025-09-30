package main

import (
	"bufio"
	"fmt"
	"go-toy/tcp-server/frame"
	"go-toy/tcp-server/packet"
	"net"
)

// handlePacket 处理消息
// 从帧中解码消息，根据消息类型处理消息，编码为确认帧
func handlePacket(framePayload []byte) (ackFramePayload []byte, err error) {
	var p packet.Packet
	p, err = packet.Decode(framePayload)
	if err != nil {
		fmt.Println("handleConn: packet decode error:", err)
		return
	}

	switch t := p.(type) {
	case *packet.Submit:
		submit := t
		fmt.Printf("recv submit: id = %s, payload=%s\n", submit.ID, string(submit.Payload))
		submitAck := &packet.SubmitAck{
			ID:     submit.ID,
			Result: 0,
		}
		packet.SubmitPool.Put(submit) // 将submit对象归还给Pool池
		ackFramePayload, err = packet.Encode(submitAck)
		if err != nil {
			fmt.Println("handleConn: packet encode error:", err)
			return nil, err
		}
		return ackFramePayload, nil
	default:
		return nil, fmt.Errorf("unknown packet type")
	}
}

// handleConn 处理连接
// 从连接中读取数据，解码为帧，处理帧中的消息，编码为确认帧，写入连接
func handleConn(c net.Conn) {
	defer c.Close()
	frameCodec := frame.NewMyFrameCodec()
	// 读写缓存
	rbuf := bufio.NewReader(c)
	wbuf := bufio.NewWriter(c)

	defer wbuf.Flush()// 确保写入缓存中的数据被发送到连接

	for {
		// read from the connection

		// decode the frame to get the payload
		// the payload is undecoded packet
		framePayload, err := frameCodec.Decode(rbuf)
		if err != nil {
			fmt.Println("handleConn: frame decode error:", err)
			return
		}

		// do something with the packet
		ackFramePayload, err := handlePacket(framePayload)
		if err != nil {
			fmt.Println("handleConn: handle packet error:", err)
			return
		}

		// write ack frame to the connection
		err = frameCodec.Encode(wbuf, ackFramePayload)
		if err != nil {
			fmt.Println("handleConn: frame encode error:", err)
			return
		}
	}
}

func main() {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("listen error:", err)
		return
	}

	fmt.Println("server start ok(on *.8888)")

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			break
		}
		// start a new goroutine to handle
		// the new connection.
		go handleConn(c)
	}
}