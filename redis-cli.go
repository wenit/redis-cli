package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var host string
var port int

func init() {
	flag.StringVar(&host, "h", "localhost", "host")
	flag.IntVar(&port, "p", 6379, "port")
}

func main() {
	conn := NewConn()

	loop(conn)

	defer conn.Close()
}

func loop(conn net.Conn) {
	for {
		fmt.Printf("%s:%d>", host, port)
		bio := bufio.NewReader(os.Stdin)
		input, _, err := bio.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", input)
	}
}

// 创建链接
func NewConn() net.Conn {
	// 解析命令行参数
	flag.Parse()

	addr := &net.TCPAddr{IP: net.ParseIP(host), Port: port}
	conn, err := net.DialTCP("tcp", nil, addr)

	if err != nil {
		log.Fatal("初始化链接失败：", err)
	}

	return conn
}
