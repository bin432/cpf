package cpf

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// bufSize			数据快 最大 大小 64kb
const maxBufSize = 64 * 1024

// BufPool buf pool
var BufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, maxBufSize)
		return &b
	},
}

// clientHandler 连接 处理
type clientHandler struct {
	server     *Server
	conn       net.Conn
	reader     *bufio.Reader
	remoteAddr string
}

func (c *clientHandler) HandleCommand() {
	c.server.log.Debug("the conn is begin", c.remoteAddr)
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			c.server.log.Error("HandleCommand: panic", err, buf)
		}
		c.conn.Close()
		c.server.log.Debug("the conn is closed", c.remoteAddr)
	}()

	c.sendMessage(0, c.server.Welcome)

	// 设置 读取 超时
	//c.conn.SetReadDeadline(time.Now().Add(time.Second))
	// 可以 处理 多个 命令
	for {
		// 读取 cmd 头
		c.resetITO()
		line, err := c.readLine()
		if err != nil {
			if err != io.EOF {
				c.server.log.Error("readLine err readcmd")
			}
			break
		}
		c.server.log.Debug(line)
		cmd, name, arg := parseLine(line)

		var ok = false
		switch cmd {
		case "PUT":
			ok = c.handlePut(name, arg)
		case "GET":
			ok = c.handleGet(name, arg)
		case "QUIT":
			c.sendMessage(0, "goodbye")
			ok = false
		default:
			c.server.log.Error("not the command: ", cmd)
			c.sendMessage(52, "not the command")
			ok = false
		}

		if !ok {
			break
		}
	}

}

// sendMessage send result msg to client
// errCode is zero, that is success
// if nonzero is err
func (c *clientHandler) sendMessage(errCode int, msg string) {
	c.resetWTO()

	var cmd string
	if 0 == errCode {
		cmd = fmt.Sprintf("+OK %s\r\n", msg)
	} else {
		cmd = fmt.Sprintf("-%d %s\r\n", errCode, msg)
	}
	snd, err := c.conn.Write([]byte(cmd))
	if err != nil {
		c.server.log.Error("conn.Write:", err)
		panic(err)
	}
	if snd < len(cmd) {
		c.server.log.Error("conn.Write less msg:")
		panic(errors.New("write less"))
	}
}

func (c *clientHandler) readLine() (string, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		if io.EOF == err {
			// 断开连接了
			return line, err
		}
		if os.IsTimeout(err) {
			c.server.log.Error("readLine timeout:", err)
			c.sendMessage(53, "cmd timeout")
		} else {
			c.server.log.Error("readLine other:", err)
		}
		return line, err
	}

	return strings.Trim(line, "\r\n"), nil
}

func (c *clientHandler) resetRTO() {
	if c.server.ReadTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.server.ReadTimeout))
	}
}

func (c *clientHandler) resetWTO() {
	if c.server.WriteTimeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.server.WriteTimeout))
	}
}

func (c *clientHandler) resetITO() {
	if c.server.IdleTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.server.IdleTimeout))
	}
}

func parseLine(line string) (string, string, string) {
	params := strings.SplitN(line, " ", 3)
	ls := len(params)
	if ls == 1 {
		return params[0], "", ""
	} else if ls == 2 {
		return params[0], params[1], ""
	}

	return params[0], params[1], params[2]
}
