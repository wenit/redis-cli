package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
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

	loop()
}

// 命令行
func loop() {
	for {
		fmt.Printf("%s:%d>", host, port)
		bio := bufio.NewReader(os.Stdin)
		// 获取命令行输入
		input, _, err := bio.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", input)
	}
}
