package cpf

import "fmt"

type defaultInterface struct{}

func (d *defaultInterface) Debug(v ...interface{}) {
	fmt.Println("Debug:", v)
}

func (d *defaultInterface) Error(v ...interface{}) {
	fmt.Println("Error:", v)
}

func (d *defaultInterface) getPutPath(authArg string) string {
	return ""
}
func (d *defaultInterface) getGetPath(authArg string, pathID string) (string, bool) {
	return "", true
}
