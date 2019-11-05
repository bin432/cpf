package cpf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (c *clientHandler) handleGet(name string, arg string) bool {
	if !c.isValid() {
		c.sendMessage(57, "not auth")
		return false
	}

	if name == "" {
		c.sendMessage(54, "command error")
		return true
	}
	offset, pathID := parseGetArg(arg)
	if -1 == offset {
		c.sendMessage(55, "bad arg")
		return true
	}
	base, ok := c.server.cfg.getGetPath(c.authArg, pathID)
	if !ok {
		c.sendMessage(55, "not the id")
		return true
	}
	realPath := filepath.Join(base, name)

	fl, err := OpenFile(realPath, os.O_RDONLY)
	if err != nil {
		c.server.log.Error("OpenFile: ", err)
		if os.IsNotExist(err) {
			c.sendMessage(20, "file does not exist")
			return true
		}
		c.sendMessage(55, "openfile err")
		return true
	}

	size, _ := fl.Seek(0, io.SeekEnd)
	// 续传
	if offset > 0 {
		if offset > size {
			fl.Close()
			c.sendMessage(54, "offset to large")
			return true
		}
		_, err = fl.Seek(offset, io.SeekStart)
		if err != nil {
			c.server.log.Error("seek: ", offset, err)
			fl.Close()
			c.sendMessage(55, "offset err")
			return true
		}
	} else {
		fl.Seek(0, io.SeekStart)
	}

	// 返回 文件总大小
	c.sendMessage(0, strconv.FormatInt(size, 10))

	bufp := BufPool.Get().(*[]byte)
	var hadSize int64
	var hadDone = false
	for {
		c.resetRTO()
		dataLine, errData := c.readLine()
		if errData != nil {
			c.server.log.Error("readLine err readdata")
			break
		}
		dataSize, errData := strconv.ParseInt(dataLine, 16, 0)
		if errData != nil {
			c.server.log.Error("data.ParseUint:", dataLine, errData)
			break
		}
		if dataSize == 0 { // 结束传输
			hadDone = true
			break
		}

		// 如果一次要读取的 buf 太大，就 返回小点
		if dataSize > maxBufSize {
			dataSize = maxBufSize
		}

		rs, errRead := fl.Read((*bufp)[:dataSize])
		if errRead != nil {
			if errRead != io.EOF { // 如果 不是 EOF 就直接 报错退出
				c.server.log.Error("fl.Read:", rs, errRead)
				break
			}
		}

		// 开始 会写 数据  先 一行 数据段 大小
		c.resetWTO()
		cmd := fmt.Sprintf("%X\r\n", rs)
		snd, err := c.conn.Write([]byte(cmd))
		if err != nil {
			c.server.log.Error("conn.Write dataSize: ", err)
			break
		}
		if snd < len(cmd) {
			c.server.log.Error("conn.Write dataSize less msg:")
			break
		}

		if rs > 0 {
			c.resetWTO()
			cws, errWrite := c.conn.Write((*bufp)[:rs])
			hadSize += int64(cws)
			if errWrite != nil {
				c.server.log.Error("conn.Write:", errWrite)
				break
			}

			if cws < rs { // the conn EOF disconnect
				c.server.log.Error("conn.Write cws < rs:", cws, rs)
				break
			}
		}
	}
	BufPool.Put(bufp)
	errClose := fl.Close()
	if errClose != nil {
		c.server.log.Error("fClose:", name, errClose)
	}

	if hadDone {
		if errClose == nil {
			// 结束传输，返回这个过程传输的 字节数，注意 不是文件的总大小
			c.sendMessage(0, strconv.FormatInt(hadSize, 10))
		} else {
			c.sendMessage(55, "file err")
		}

		return true
	}
	// 在 传输 数据段时，出现了任何错误都将断开连接
	return false
}

func parseGetArg(arg string) (offset int64, id string) {
	var err error
	id = "0"
	params := strings.SplitN(arg, " ", 2)
	if len(params) > 1 {
		id = params[1]
	}

	offset, err = strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		offset = -1
	}
	return
}
