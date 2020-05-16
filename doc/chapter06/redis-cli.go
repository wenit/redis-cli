package main

import (
	"flag"
	"fmt"
	"github.com/chzyer/readline"
	"io"
	"log"
	"net"
	"os"
	"path"
	"regexp"
	"strings"
)

var host string
var port int

func init() {
	flag.StringVar(&host, "h", "localhost", "host")
	flag.IntVar(&port, "p", 6379, "port")
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("set",
		readline.PcItem("key value [EX seconds]|[PX milliseconds]"),
	),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
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

func loop(conn *Conn) {

	home, _ := os.UserHomeDir()
	historyFile := path.Join(home, ".rediscli_history")

	prompt := fmt.Sprintf("%s:%d>", host, port)
	l, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     historyFile,
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.SetOutput(l.Stderr())

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		// 这里没有使用strings.Fields方法进行分隔，这个方法存在一个问题，比如：set name "zhang san"，
		// 会被分隔层4个字段，不是我们想要的3个字段
		// fields:=strings.Fields(line)

		re := regexp.MustCompile(`("[^"]+?"\S*|\S+)`)
		fields := re.FindAllString(line, -1)

		var rsp interface{}
		if len(fields) == 0 {
			continue
		}

		cmd := fields[0]
		if len(fields) > 1 {
			args := fields[1:]
			rsp, err = conn.Do(cmd, args...)
		} else {
			rsp, err = conn.Do(cmd)
		}

		if err != nil {
			log.Println(err)
			continue
		}

		switch rsp.(type) {
		case []interface{}:
			arr, _ := rsp.([]interface{})
			strArr := make([]string, 0)
			for i := 0; i < len(arr); i++ {
				v, _ := arr[i].([]byte)
				strArr = append(strArr, string(v))
			}
			fmt.Printf("[%v]\n", strings.Join(strArr, ","))
		case int:
			fmt.Printf("%d\n", rsp)
		case nil:
			fmt.Printf("(nil)\n")
		default:
			fmt.Printf("%s\n", rsp)
		}

		if line == "quit" {
			break
		}

	}
}

// 创建连接
func NewConn() *Conn {

	addr := &net.TCPAddr{IP: net.ParseIP(host), Port: port}
	conn, err := net.DialTCP("tcp", nil, addr)

	if err != nil {
		log.Fatal("初始化链接失败：", err)
	}

	return NewRedisConn(conn)
}
