# 简介
cpf 简单的文件传输协议

支持DEL、PUT上传文件
支持GET下载文件，GET支持多目录

# 例子
```go
func main() {
	server := cpf.NewServer(nil, nil)
	server.ListenAndServe("localhost:8200")
}
```
# 客户端测试
可以在终端上使用telnet测试
```
telnet localhost 8200
AUTH arg
+OK auth success

DEL 123.txt
+OK success

PUT 123.txt auto
+OK filesize

GET 123.txt offset
+OK filesize

QUIT
+OK good bye
```

# 
一共5个指令，其中 执行 PUT或GET指令后，程序 进入 数据传输模式。