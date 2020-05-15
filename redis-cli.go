package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
)

var host string
var port int

func init() {
	flag.StringVar(&host, "h", "localhost", "host")
	flag.IntVar(&port, "p", 6379, "port")
}

// 程序入口
func main() {
	// 解析命令行参数
	flag.Parse()

	conn := NewConn()

	loop(conn)

	// 关闭连接
	defer conn.Close()
}

// 命令行
func loop(conn net.Conn) {

	inputStr := "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n"

	// fmt.Printf("%s\n", input)

	// inputStr:=string(input)

	// fields:=strings.Fields(inputStr)

	// req:=MultiBulkMarshal(fields...);

	//Write(conn,[]byte(req))
	Write(conn, []byte(inputStr))

	data := Read(conn)

	fmt.Printf("%s\n", data)

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
