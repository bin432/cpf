package cpf

import (
	"fmt"
	"io"
)

// Dataer 模拟 数据
type Dataer struct {
	count  int
	offset int
}

// NewDataer is
func NewDataer(count int) *Dataer {
	d := &Dataer{
		count: count,
	}
	return d
}

// Read buf
func (r *Dataer) Read(p []byte) (n int, err error) {
	if r.offset >= r.count {
		return 0, io.EOF
	}

	r.offset++

	data := fmt.Sprintf("[%05d]这是个测试数据啊，随便输入一些数据，比如-m¹mⁿ.txt-\n", r.offset)

	rs := copy(p, []byte(data))

	return rs, nil
}

// Seek data
func (r *Dataer) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

// PutFile demo
func PutFile(addr string, name string) (string, int64) {
	cp, err := Dial(addr)
	if err != nil {
		return "", -1
	}
	defer cp.Close()

	rser := &Dataer{
		count: 4096,
	}

	putSize := cp.Put(name, rser, nil)
	return name, putSize
}

// GetFile is a demo
func GetFile(addr string) int64 {
	return 121
}
