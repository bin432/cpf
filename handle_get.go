package cpf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (c *clientHandler) handleGet(name string, arg string) {
	if !c.isValid() {
		c.sendMessage(33, "not auth")
		return
	}

	if name == "" {
		c.sendMessage(55, "command error")
		return
	}
	offset, pathID := parseGetArg(arg)
	if -1 == offset {
		c.sendMessage(55, "bad arg")
		return
	}
	getPath, err := c.server.cfg.QueryGetPath(c.authArg, pathID)
	if err != nil {
		c.server.log.Error("QueryGetPath err:", err)
		if os.IsNotExist(err) {
			c.sendMessage(43, "not the id")
		} else {
			c.sendMessage(55, "query path error")
		}
		return
	}
	realPath := filepath.Join(getPath, name)
	fl, err := OpenFile(realPath, os.O_RDONLY)
	if err != nil {
		c.server.log.Error("OpenFile: ", err)
		if os.IsNotExist(err) {
			c.sendMessage(44, "file does not exist")
			return
		}
		c.sendMessage(55, "openfile err")
		return
	}
	defer fl.Close() // 只读 方法 Close 基本上 不会 报错

	// 先 获取 文件 大小
	size, _ := fl.Seek(0, io.SeekEnd)
	// 续传
	if offset > 0 {
		if offset > size {
			c.server.log.Error("offset to large:", offset, size)
			c.sendMessage(55, "offset to large")
			return
		}
		// 移动 文件指针 到 off
		_, err = fl.Seek(offset, io.SeekStart)
		if err != nil {
			c.server.log.Error("seek: ", offset, err)
			c.sendMessage(55, "offset err")
			return
		}
	} else {
		// 再次 将文件指针 移动到 开始处
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

		// 因为 BufPool大小限制 如果一次要读取的 buf 太大，就 返回 BufPool 的大小
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
	// fl.Close 在 上面 defer

	if hadDone {
		// 结束传输，返回这个过程传输的 字节数，注意 不是文件的总大小
		c.sendMessage(0, strconv.FormatInt(hadSize, 10))
		return
	}
	// 走到这里 就说明 出错了
	c.sendMessage(55, "server err")
	panic("server err") // 使用 panic 断开连接
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
