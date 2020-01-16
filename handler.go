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
	isAuthed   bool
	authArg    string
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

	c.sendMessage(ErrCodeOK, c.server.Welcome)

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

		switch cmd {
		case "AUTH":
			c.handleAuth(name)
		case "DEL":
			c.handleDel(name)
		case "PUT":
			c.handlePut(name, arg)
		case "GET":
			c.handleGet(name, arg)
		case "QUIT":
			c.sendMessage(ErrCodeOK, "goodbye")
			break
		default:
			c.server.log.Error("not the command: ", cmd)
			c.sendMessage(ErrCodeNotCmd, "not the command")
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
			c.sendMessage(ErrCodeTimeOut, "cmd timeout")
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

// 处理 身份验证
func (c *clientHandler) handleAuth(auth string) {
	// 为空 表示 通过
	if nil == c.server.auth {
		c.isAuthed = true
	} else {
		c.isAuthed = c.server.auth(auth)
	}
	if c.isAuthed {
		c.sendMessage(ErrCodeOK, "auth success")
	} else {
		c.sendMessage(ErrCodeAuth, "auth faild")
	}
}

// 判断 是否 有效
func (c *clientHandler) isValid() bool {
	if nil == c.server.auth {
		return true
	}

	return c.isAuthed
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
