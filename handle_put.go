package cpf

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// handlePut return  false that close the conn
func (c *clientHandler) handlePut(name string, arg string) bool {
	if name == "" {
		c.sendMessage(54, "command error")
		return true
	}
	realPath := filepath.Join(c.server.cfg.getPutPath(), name)

	var fileFlag = os.O_WRONLY
	if arg == "appd" {

	} else if arg == "new" {
		fileFlag |= os.O_CREATE
		fileFlag |= os.O_EXCL
	} else if arg == "auto" {
		fileFlag |= os.O_CREATE
	} else {
		c.sendMessage(55, "bad arg")
		return true
	}

	// 读写文件 请使用 handle_file 里的 文件io 做了 协程数限制
	fl, err := OpenFile(realPath, fileFlag)
	if err != nil {
		c.server.log.Error("OpenFile: ", err)
		if os.IsNotExist(err) {
			c.sendMessage(20, "file does not exist")
			return true
		}
		if os.IsExist(err) {
			c.sendMessage(51, "file already exists")
			return true
		}

		c.sendMessage(55, "openfile err")
		return true
	}
	// 移动到 文件尾
	size, _ := fl.Seek(0, io.SeekEnd)

	// 返回 文件 指针 offset 用于 断点续传
	c.sendMessage(0, strconv.FormatInt(size, 10))

	// 开始 接收 数据段
	// 先解析 16进制的 的数据大小
	//buf := make([]byte, 16*1024)
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
		if dataSize == 0 { // 传输 完成
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
			c.sendMessage(0, strconv.FormatInt(hadSize, 10))
		} else {
			c.sendMessage(55, "file save faided")
		}

		return true
	}
	// 在 传输 数据段时，出现了任何错误都将断开连接
	return false
}
