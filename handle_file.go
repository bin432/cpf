// LFile 用来限制 并发 操作文件，基本上 一个系统调用要 创建一个线程，
// 同时系统调用就会 创建多个线程，
//
// 这个模块 添加了 对文件操作的 goroutine限制，同时 只能有100个goroutine能操作文件
//

package cpf

import (
	"os"
)

// 控制 只能100个 同时操作文件，高并发限制
var fgg = make(ghan, 100)

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

	f, err := os.OpenFile(name, flag, 0666)
	if err != nil {
		fgg.leave()
		return nil, err
	}
	fgg.leave()
	return &LFile{f: f}, nil
}

// Close f
func (f *LFile) Close() (err error) {
	fgg.enter()
	err = f.f.Close()
	fgg.leave()
	return
}

// Seek go
func (f *LFile) Seek(offset int64, whence int) (ret int64, err error) {
	fgg.enter()
	ret, err = f.f.Seek(offset, whence)
	fgg.leave()
	return
}

func (f *LFile) Read(p []byte) (n int, err error) {
	fgg.enter()
	n, err = f.f.Read(p)
	fgg.leave()
	return
}

func (f *LFile) Write(p []byte) (n int, err error) {
	fgg.enter()
	n, err = f.f.Write(p)
	fgg.leave()
	return
}
