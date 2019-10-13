// LFile 用来限制 并发 操作文件，基本上 一个系统调用要 创建一个线程，
// 同时系统调用就会 创建多个线程，
//
// 这个模块 添加了 对文件操作的 goroutine限制，同时 只能有100个goroutine能操作文件
//

package cpf

import (
	"os"
)

// 限制 同时操作文件 的 gos，高并发限制
var fgg ghan

type ghan chan bool

func (g ghan) enter() { g <- true }
func (g ghan) leave() { <-g }

// LFile is a
type LFile struct {
	f *os.File
}

// OpenFile f
func OpenFile(name string, flag int) (*LFile, error) {
	fgg.enter()
	defer fgg.leave()
	f, err := os.OpenFile(name, flag, 0666)
	if err != nil {
		return nil, err
	}
	return &LFile{f: f}, nil
}

// Close f
func (f *LFile) Close() error {
	fgg.enter()
	defer fgg.leave()
	return f.f.Close()
}

// Seek go
func (f *LFile) Seek(offset int64, whence int) (int64, error) {
	fgg.enter()
	defer fgg.leave()
	return f.f.Seek(offset, whence)
}

func (f *LFile) Read(p []byte) (n int, err error) {
	fgg.enter()
	defer fgg.leave()
	return f.f.Read(p)
}

func (f *LFile) Write(p []byte) (n int, err error) {
	fgg.enter()
	defer fgg.leave()
	return f.f.Write(p)
}
