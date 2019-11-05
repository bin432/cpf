package cpf

import (
	"bufio"
	"net"
	"time"
)

var def = &defaultInterface{}

// 日志 接口
type logger interface {
	Debug(v ...interface{})
	Error(v ...interface{})
}

// 配置 接口
type configer interface {
	getPutPath(authArg string) string
	getGetPath(authArg string, pathID string) (string, bool)
}

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
