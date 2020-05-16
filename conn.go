package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
)

// Redis连接对象
type Conn struct {
	c          net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	readIndex  int
	bufSize    int
	bufFull    bool
	buf        []byte
	maxBufLoop int
}

func NewRedisConn(netCon net.Conn) *Conn {
	return &Conn{
		c:          netCon,
		reader:     bufio.NewReader(netCon),
		writer:     bufio.NewWriter(netCon),
		bufSize:    4096,
		maxBufLoop: 10,
		bufFull:    false,
	}
}

// 写入数据
func (c Conn) Write(data []byte) {
	c.writer.Write(data)
}

// 写入数据并立即刷新输出缓冲
func (c *Conn) WriteFlush(data []byte) {
	c.Write(data)
	c.writer.Flush()
}

func (c *Conn) Do(cmd string, args ...string) (interface{}, error) {

	cmds := make([]string, 0)

	cmds = append(cmds, cmd)
	cmds = append(cmds, args...)

	req := c.MultiBulkMarshal(cmds...)

	c.WriteFlush([]byte(req))

	rsp, err := c.ReadReply()

	c.reset()

	return rsp, err
}

func (c *Conn) MultiBulkMarshal(args ...string) string {
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

func (c *Conn) readByte() (byte, error) {
	return c.reader.ReadByte()
}

func (c *Conn) reset() {
	c.readIndex = 0
	c.buf = c.buf[0:0]
}

func (c *Conn) subReadLine() ([]byte, bool) {
	data := c.buf[c.readIndex:]

	i := bytes.IndexByte(data, '\r')

	var b bool
	if i > 0 && (i+1) < len(data) {
		if data[i+1] == '\n' {
			c.readIndex += i + 2
			b = true
			return data[0:i], true
		}
	}
	return []byte{}, b
}

func (c *Conn) readLine() ([]byte, error) {
	for i := 0; i < c.maxBufLoop; i++ {
		r, b := c.subReadLine()
		if b {
			return r, nil
		}
		c.readToBuf()
	}
	errInfo := fmt.Sprintf("over max buffer size [%d]", c.maxBufLoop*c.bufSize)
	return []byte{}, errors.New(errInfo)
}

func (c *Conn) parseLine(data []byte) ([]byte, error) {
	n, err := parseLen(data[1:])
	if err != nil {
		return nil, err
	}

	if n == -1 {
		return nil, nil
	}

	numLen := len(string(n))

	index := 1 + numLen
	lastIndex := index + n

	p := data[lastIndex:lastIndex]
	return p, nil
}

func (c *Conn) readToBuf() error {
	buf := make([]byte, c.bufSize)
	n, err := c.reader.Read(buf)

	if err != nil {
		return err
	}
	if n == c.bufSize {
		c.bufFull = true
	}

	c.buf = append(c.buf, buf...)

	return nil
}

// 读取响应
func (c *Conn) ReadReply() (interface{}, error) {
	if c.readIndex == 0 {
		err := c.readToBuf()
		if err != nil {
			return nil, err
		}
	}

	line, err := c.readLine()

	if err != nil {
		return nil, err
	}

	if len(line) == 0 {
		return nil, errors.New("short response line")
	}

	// 获取第一个字节
	fb := line[0]

	var result interface{}
	var retErr error

	switch fb {
	case '+', '-':
		result = string(line[1:])
	case ':':
		result, retErr = parseLen(line[1:])
	case '$': // 读取批量回复
		result, retErr = parseLen(line[1:])
		if result == -1 {
			return nil, nil
		}

		return c.readLine()
	case '*': // 读取多批量回复
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		r := make([]interface{}, n)
		for i := range r {
			r[i], err = c.ReadReply()
			if err != nil {
				return nil, err
			}
		}
		return r, nil
	}
	return result, retErr
}

func parseLen(data []byte) (int, error) {
	str := string(data)
	n, err := strconv.Atoi(str)
	return n, err
}

func (c Conn) Close() {
	c.c.Close()
}
