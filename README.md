# 简介
cpf 简单的文件传输协议

支持PUT上传文件和GET下载文件

GET支持多目录

# 例子
```go
func main() {
	ser := cpf.NewServer(nil, nil)
	ser.ListenAndServe("localhost:8200")
}
```
# 客户端测试
可以在终端上使用telnet测试
```
telnet localhost 8200
PUT 123.txt auto
GET 123.txt
```