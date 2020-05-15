package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

var host string
var port int

// 程序入口
func main() {
	host = "localhost"
	port = 6379

	conn := NewConn()

	rsp := Do(conn, "set", "name", "zhangsan")
	fmt.Printf("%s\n", rsp)

	rsp = Do(conn, "get", "name")
	fmt.Printf("%s\n", rsp)

	// 关闭连接
	defer conn.Close()
}

// 创建连接
func NewConn() net.Conn {

	addr := &net.TCPAddr{IP: net.ParseIP(host), Port: port}
	conn, err := net.DialTCP("tcp", nil, addr)

	if err != nil {
		log.Fatal("初始化链接失败：", err)
	}

	return conn
}

// 写入数据
func Write(conn net.Conn, input []byte) {
	conn.Write(input)
}

// 读取响应
func Read(conn net.Conn) []byte {
	buff := make([]byte, 1024)
	n, err := conn.Read(buff)
	if err != nil {
		log.Fatal(err)
	}
	return buff[0:n]
}

// 执行命令
func Do(conn net.Conn, cmd string, args ...string) string {
	cmds := make([]string, 0)

	cmds = append(cmds, cmd)
	cmds = append(cmds, args...)

	req := MultiBulkMarshal(cmds...)

	Write(conn, []byte(req))

	rspBytes := Read(conn)

	rsp := string(rspBytes)
	return rsp
}

func MultiBulkMarshal(args ...string) string {
	var s string
	s = "*"
	s += strconv.Itoa(len(args))
	s += "\r\n"

	// 命令所有参数
	for _, v := range args {
		s += "$"
		s += strconv.Itoa(len(v))
		s += "\r\n"
		s += v
		s += "\r\n"
	}

	return s
}
