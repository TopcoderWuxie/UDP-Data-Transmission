package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

// UDP包的最大传输量：http://www.cnblogs.com/linuxbug/p/4906000.html
var sendData = make([]byte, 65507)
var random int64

func init() {
	// 使用时间种子来获取不同的随机数
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random = r.Int63n(65508)
	f, _ := os.Open("text.txt")
	sendData = make([]byte, random)
	f.Read(sendData)
	fmt.Println("本次发送的UDP数据包的大小为：", len(string(sendData)))
}

func main() {

	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 8080,
	})
	if err != nil {
		fmt.Println("连接失败!", err)
		return
	}
	defer socket.Close()

	// 发送数据
	_, err = socket.Write(sendData)
	if err != nil {
		fmt.Println("发送数据失败!", err)
		return
	}
	// 接收数据
	data := make([]byte, 65507)
	readNumbers, _, err := socket.ReadFromUDP(data)
	if err != nil {
		fmt.Println("读取数据失败!", err)
		return
	}
	fmt.Printf("%s\n", data[:readNumbers])
}
