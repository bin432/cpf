package cpf

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

// 日志 接口
type logger interface {
	Debug(v ...interface{})
	Error(v ...interface{})
}

// 默认的 日志 记录
type defLogger struct{}

func (l *defLogger) Debug(v ...interface{}) {
	fmt.Println("Debug: ", v)
}
func (l *defLogger) Error(v ...interface{}) {
	fmt.Println("Error: ", v)
}

// Server is a cpf server
type Server struct {
	// 欢迎 语句
	welcome  string
	listener net.Listener
	logger   logger

	putPath  string            // 上传 路径
	getPaths map[string]string // 下载 路径 支持 多个路径 id-path

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// NewServer create a server
func NewServer(welcome string, put string, gets map[string]string, log logger) *Server {
	s := &Server{
		welcome:  welcome,
		putPath:  put,
		getPaths: gets,

		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  time.Minute * 5,
	}

	if log == nil {
		s.logger = &defLogger{}
	} else {
		s.logger = log
	}

	return s
}

// ListenAndServe the addr
func (s *Server) ListenAndServe(addr string) error {
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		s.logger.Error("net.Listen:", err)
		return err
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Error("listener.Accept:", err)
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
