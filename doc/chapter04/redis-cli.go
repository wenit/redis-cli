package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
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

	// 创建连接
	conn := NewConn()

	// 循环获取输入
	loop(conn)

	// 关闭连接
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
		//fmt.Printf("%s\n", input)
		line := strings.TrimSpace(string(input))

		// 这里没有使用strings.Fields方法进行分隔，这个方法存在一个问题，比如：set name "zhang san"，会被分隔层4个字段，不是我们想要的3个字段
		// fields:=strings.Fields(line)

		re := regexp.MustCompile(`("[^"]+?"\S*|\S+)`)
		fields := re.FindAllString(line, -1)

		rsp := Do(conn, fields[0], fields[1:]...)

		v, ok := rsp.([]string)
		if ok {
			fmt.Printf("[%s]\n", strings.Join(v, ","))
		} else {
			fmt.Printf("%s\n", rsp)
		}
	}
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
func Do(conn net.Conn, cmd string, args ...string) interface{} {
	cmds := make([]string, 0)

	cmds = append(cmds, cmd)
	cmds = append(cmds, args...)

	req := MultiBulkMarshal(cmds...)

	Write(conn, []byte(req))

	rspBytes := Read(conn)

	if rspBytes[0] == '*' {
		multiData, _ := MultiUnMarsh(rspBytes)
		rsp := make([]string, 0, len(multiData))
		for i := 0; i < len(multiData); i++ {
			rsp = append(rsp, string(multiData[i]))
		}
		return rsp
	} else {
		rpsData, _, _ := SingleUnMarshal(rspBytes)
		rsp := string(rpsData)
		return rsp
	}
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

// 第一种状态回复：
//状态回复是一段以 "+" 开始， "\r\n" 结尾的单行字符串。如 SET 命令成功的返回值："+OK\r\n"
//所以我们判断第一个字符是否等于 '+' 如果相等，则读取到\r\n
func SingleUnMarshal(data []byte) ([]byte, int, error) {
	var result []byte
	var err error
	var readLen int

	switch data[0] {
	case '+', '-', ':':
		result, err = ReadLine(data[1:])
		readLen = len(result) + 3
	case '$':
		n, err := ReadLine(data[1:])
		if err != nil {
			return []byte{}, 0, err
		}
		dateLen, err := strconv.Atoi(string(n))
		if err != nil {
			return []byte{}, 0, err
		}
		if dateLen == -1 {
			return []byte{}, 0, err
		}
		// +3 的原因 $ \r \n 三个字符
		result = data[len(n)+3 : len(n)+3+dateLen]
		readLen = len(n) + 5 + dateLen

	}

	return result, readLen, err
}

func ReadLine(data []byte) ([]byte, error) {
	for i := 0; i < len(data); i++ {
		if data[i] == '\r' {
			if data[i+1] != '\n' {
				return []byte{}, errors.New("format error")
			}

			return data[0:i], nil
		}
	}
	return []byte{}, errors.New("format error")
}

func MultiUnMarsh(data []byte) ([][]byte, error) {
	if data[0] != '*' {
		return [][]byte{}, errors.New("format error")
	}
	n, err := ReadLine(data[1:])
	if err != nil {
		return [][]byte{}, err
	}
	loopSize, err := strconv.Atoi(string(n))
	if err != nil {
		return [][]byte{}, err
	}
	// 多条批量回复也可以是空白的（empty)
	if loopSize == 0 {
		return [][]byte{}, nil
	}

	// 无内容的多条批量回复（null multi bulk reply）也是存在的,
	// 客户端库应该返回一个 null 对象, 而不是一个空数组。
	if loopSize == -1 {
		return nil, nil
	}

	result := make([][]byte, loopSize)

	readLen := len(n) + 3

	for i := 0; i < loopSize; i++ {
		ret, length, err := SingleUnMarshal(data[readLen:])
		if err != nil {
			return [][]byte{}, errors.New("format error")
		}
		result[i] = ret
		readLen += length
	}
	return result, nil
}
