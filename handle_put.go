package cpf

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// handleDel 只能 删除 putPath 里的文件
func (c *clientHandler) handleDel(name string) {
	if !c.isValid() {
		c.sendMessage(ErrCodeAuth, "not auth")
		return
	}

	if name == "" {
		c.sendMessage(ErrCodeCmdArgs, "command error")
		return
	}
	putPath, err := c.server.cfg.QueryPutPath(c.authArg)
	if err != nil {
		c.server.log.Error("QueryPutPath err:", err)
		c.sendMessage(ErrCodeServer, "query path error")
		return
	}
	realPath := filepath.Join(putPath, name)
	err = os.Remove(realPath)
	if err != nil {
		if !os.IsNotExist(err) {
			c.server.log.Error("Del err:", err)
			c.sendMessage(ErrCodeServer, "del fail")
			return
		}
	}
	c.sendMessage(ErrCodeOK, "del success")
}

// handlePut
func (c *clientHandler) handlePut(name string, arg string) {
	if !c.isValid() {
		c.sendMessage(ErrCodeAuth, "not auth")
		return
	}

	if name == "" {
		c.sendMessage(ErrCodeCmdArgs, "command error")
		return
	}
	putPath, err := c.server.cfg.QueryPutPath(c.authArg)
	if err != nil {
		c.server.log.Error("QueryPutPath err:", err)
		c.sendMessage(ErrCodeServer, "query path error")
		return
	}
	realPath := filepath.Join(putPath, name)

	var fileFlag = os.O_WRONLY
	if arg == "appd" {
		// 后面 会 移到 文件尾
	} else if arg == "new" {
		fileFlag |= os.O_CREATE
		fileFlag |= os.O_EXCL
	} else if arg == "auto" {
		fileFlag |= os.O_CREATE
	} else {
		c.server.log.Error("bad arg:", arg)
		c.sendMessage(ErrCodeCmdArgs, "bad arg")
		return
	}

	// 读写文件 请使用 handle_file 里的 文件io 做了 协程数限制
	fl, err := OpenFile(realPath, fileFlag)
	if err != nil {
		c.server.log.Error("OpenFile: ", err)
		if os.IsNotExist(err) {
			c.sendMessage(ErrCodeNotFile, "file does not exist")
			return
		}
		if os.IsExist(err) {
			c.sendMessage(ErrCodeExisted, "file already exists")
			return
		}

		c.sendMessage(ErrCodeServer, "openfile err")
		return
	}
	// 移动到 文件尾
	size, _ := fl.Seek(0, io.SeekEnd)

	// 返回 文件 指针 offset 用于 断点续传
	c.sendMessage(ErrCodeOK, strconv.FormatInt(size, 10))

	// 开始 接收 数据段
	//buf := make([]byte, 16*1024)
	bufp := BufPool.Get().(*[]byte)
	var hadSize int64
	var hadDone = false
	for {
		c.resetRTO()
		dataLine, errLine := c.readLine()
		if errLine != nil {
			c.server.log.Error("readLine err readdata")
			break
		}
		// 先解析 16进制的 的数据大小
		dataSize, errData := strconv.ParseInt(dataLine, 16, 0)
		if errData != nil {
			c.server.log.Error("data.ParseUint:", dataLine, errData)
			break
		}
		if dataSize == 0 {
			// 传输 完成
			hadDone = true
			break
		}

		// 使用CopyBuffer 避免内存 频繁创建和删除
		cps, errCopy := io.CopyBuffer(fl, io.LimitReader(c.reader, dataSize), *bufp)
		hadSize += cps
		if errCopy != nil {
			c.server.log.Error("CopyBuffer File:", errCopy)
			break
		}

		if cps < dataSize { // the conn EOF disconnect
			c.server.log.Error("CopyBuffer cps < dataSize:", cps, dataSize)
			break
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
			c.sendMessage(ErrCodeOK, strconv.FormatInt(hadSize, 10))
		} else {
			c.sendMessage(ErrCodeServer, "file save faided")
		}
		return
	}
	// 走到这里 就说明 出错了
	c.sendMessage(ErrCodeServer, "server err")
	panic("server err") // 使用 panic 断开连接
}
