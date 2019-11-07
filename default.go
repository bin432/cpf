package cpf

import "fmt"

type defaultInterface struct{}

func (d *defaultInterface) Debug(v ...interface{}) {
	fmt.Println("Debug:", v)
}

func (d *defaultInterface) Error(v ...interface{}) {
	fmt.Println("Error:", v)
}

func (d *defaultInterface) QueryPutPath(authArg string) (string, error) {
	return "", nil
}
func (d *defaultInterface) QueryGetPath(authArg string, pathID string) (string, error) {
	return "", nil
}
