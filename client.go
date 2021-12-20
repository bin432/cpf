package cpf

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// Client is a cpf connect
type Client struct {
	address string
	welcome string
	conn    net.Conn
	reader  *bufio.Reader
	buf     []byte // 先定义一个buf
}

// Dial new Client and Connect
func Dial(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	c := &Client{
		address: address,
		conn:    conn,
		reader:  bufio.NewReaderSize(conn, 1024),
		buf:     make([]byte, 64*1024),
	}
	c.welcome = c.readLine()
	if c.welcome == "" {
		conn.Close()
		return nil, err
	}
	fmt.Println(c.welcome)
	return c, nil
}

// Put 上传文件
// fn Progress callback
func (c *Client) Put(name string, rser io.ReadSeeker, fn func(int) bool) int64 {
	putCmd := fmt.Sprintf("PUT %s auto\r\n", name)
	_, err := c.conn.Write([]byte(putCmd))
	if err != nil {
		fmt.Println("Put Write err:", err)
		return -1
	}

	line := c.readLine()
	if line == "" {
		return -1
	}

	code, msg := parseCmd(line)
	if code != 0 {
		fmt.Printf("Put %s err:code-%d,msg-%s\r\n", name, code, msg)
		return -1
	}
	seek, _ := strconv.ParseInt(msg, 10, 0)
	if seek > 0 {
		_, _ = rser.Seek(seek, io.SeekStart)
	}

	for {
		rc, err := rser.Read(c.buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Print("rser Read Err:", err)
				break
			}
		}
		c.sendData(c.buf[:rc])
		if fn != nil && !fn(rc) {
			break
		}
	}
	return c.endData()
}

// Get 下载文件
// fn Progress callback
func (c *Client) Get(name string, wser io.WriteSeeker, fn func(int) bool) int64 {
	var err error
	// 移动到文件尾，获取 offset
	offset, err := wser.Seek(0, io.SeekEnd)

	putCmd := fmt.Sprintf("GET %s %d\r\n", name, offset)
	_, err = c.conn.Write([]byte(putCmd))
	if err != nil {
		fmt.Println("Get Write err:", err)
		return -1
	}

	line := c.readLine()
	if line == "" {
		return -1
	}

	code, msg := parseCmd(line)
	if code != 0 {
		fmt.Printf("Get %s err:code-%d,msg-%s\r\n", name, code, msg)
		return -1
	}
	// 获取 服务器上 文件 大小
	fsize, _ := strconv.ParseInt(msg, 10, 0)
	fmt.Println("size:", fsize)

	for {
		rc := c.writeTo(wser)
		if rc == -1 {
			fmt.Print("c recvData -1")
			break
		} else if rc == 0 {
			break
		}
		if fn != nil && !fn(rc) {
			break
		}
	}
	return c.endData()
}

func (c *Client) readLine() string {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		if io.EOF == err {
			// 断开连接了
			fmt.Println("readLine EOF")
			return ""
		}
		if os.IsTimeout(err) {
			fmt.Println("readLine timeout:", err)
		} else {
			fmt.Println("readLine other:", err)
		}
		return ""
	}

	return strings.Trim(line, "\r\n")
}

func (c *Client) sendData(data []byte) {
	cmd := fmt.Sprintf("%x\r\n", len(data))
	_, err := c.conn.Write([]byte(cmd))
	if err != nil {
		fmt.Println("sendData.WriteCmd err:", err)
	}
	_, err = c.conn.Write(data)
	if err != nil {
		fmt.Println("sendData.WriteData err:", err)
	}
}

func (c *Client) writeTo(wser io.WriteSeeker) int {
	// 每次向 服务器 读取数据 都已 buf 度量
	cmd := fmt.Sprintf("%x\r\n", len(c.buf))
	_, err := c.conn.Write([]byte(cmd))
	if err != nil {
		fmt.Println("recvData.WriteCmd err:", err)
		return -1
	}

	line := c.readLine()
	if line == "" {
		return -1
	}

	size, err := strconv.ParseInt(line, 16, 0)
	if err != nil {
		fmt.Println("data.ParseUint:", line, err)
		return -1
	}

	// 服务器 文件 EOF
	if size == 0 {
		return 0
	}

	cps, errCopy := io.CopyBuffer(wser, io.LimitReader(c.reader, size), c.buf)
	if errCopy != nil {
		fmt.Println("CopyBuffer:", errCopy)
		return -1
	}
	if cps < size {
		fmt.Println("cps < size", cps, size)
		return -1
	}

	return int(cps)
}

func (c *Client) endData() int64 {
	_, err := c.conn.Write([]byte("0\r\n"))
	if err != nil {
		fmt.Println("endData.Write err:", err)
	}
	cmd := c.readLine()
	if cmd == "" {
		return -1
	}

	code, msg := parseCmd(cmd)
	if code != 0 {
		fmt.Printf("endData err:code-%d,msg-%s\r\n", code, msg)
		return -1
	}
	s, _ := strconv.ParseInt(msg, 10, 0)
	return s
}

// Close thar close the conn
func (c *Client) Close() {
	c.conn.Write([]byte("QUIT\r\n"))
	cmd := c.readLine()
	fmt.Println(cmd)
	c.conn.Close()
}

func parseCmd(cmd string) (code int, msg string) {
	params := strings.SplitN(cmd, " ", 2)
	if "+OK" == params[0] {
		code = 0
	} else {
		code, _ = strconv.Atoi(msg)
		if code == 0 { // 这里 不可能为0
			code = -1
		}
	}
	if len(params) == 1 {
		msg = ""
	} else {
		msg = params[1]
	}
	return
}
