package cpf

import (
	"bufio"
	"errors"
	"net"
	"time"
)

const (
	// ErrCodeOK 成功
	ErrCodeOK = 0
	// ErrCodeAuth 验证失败
	ErrCodeAuth = 33
	// ErrCodeNotFile 没有该文件
	ErrCodeNotFile = 42
	// ErrCodeExisted 已经有了
	ErrCodeExisted = 43
	// ErrCodeNotCmd 没有该命令
	ErrCodeNotCmd = 44
	// ErrCodeNotID 没有该 id
	ErrCodeNotID = 45
	// ErrCodeOffset offset 错误
	ErrCodeOffset = 46
	// ErrCodeTimeOut 超时
	ErrCodeTimeOut = 53
	// ErrCodeCmdArgs 命令错误
	ErrCodeCmdArgs = 54
	// ErrCodeServer  服务器错误
	ErrCodeServer = 55
)

var def = &defaultInterface{}

// 日志 接口
type logger interface {
	Debug(v ...interface{})
	Error(v ...interface{})
}

// 配置 接口
type configer interface {
	QueryPutPath(authArg string) (string, error)
	QueryGetPath(authArg string, pathID string) (string, error)
}

// ErrNotPathID 是QueryGetPath的特指
var ErrNotPathID = errors.New("cpf: not pathID")

// Server is a cpf server
type Server struct {
	// 欢迎 语句
	Welcome  string
	listener net.Listener
	cfg      configer
	log      logger

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	auth         func(string) bool
}

// NewServer create a server
func NewServer(cfg configer, log logger) *Server {
	s := &Server{
		Welcome: "Welcome to cpf",

		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  time.Minute * 5,
	}

	if cfg == nil {
		s.cfg = def
	} else {
		s.cfg = cfg
	}

	if log == nil {
		s.log = def
	} else {
		s.log = log
	}

	return s
}

// SetFileGos 设置 操作 文件io 的 并发数
func (s *Server) SetFileGos(count int) {
	fgg = make(ghan, count)
}

// SetAuthFunc 设置 验证接口， 不设置则为 不启用 验证
func (s *Server) SetAuthFunc(fn func(string) bool) {
	s.auth = fn
}

// ListenAndServe the addr
func (s *Server) ListenAndServe(addr string) error {
	if fgg == nil {
		fgg = make(ghan, 100)
	}

	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		s.log.Error("net.Listen:", err)
		return err
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.log.Error("listener.Accept:", err)
			break
		}

		c := s.newClientHandler(conn)
		go c.HandleCommand()
	}

	return err
}

func (s *Server) newClientHandler(conn net.Conn) *clientHandler {
	c := &clientHandler{
		server:     s,
		conn:       conn,
		reader:     bufio.NewReader(conn),
		remoteAddr: conn.RemoteAddr().String(),
	}

	return c
}
