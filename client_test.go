package cpf_test

import (
	"cpfd/cpf"
	"testing"
)

func TestPutFile(t *testing.T) {
	name, putSize := cpf.PutFile("192.168.3.77:8200", "")

	t.Logf("Put %s size %d\r\n", name, putSize)
}

func TestGetFile(t *testing.T) {
	cpf.GetFile("192.168.3.77:8200")
}

func Benchmark_putFile(b *testing.B) {

}
