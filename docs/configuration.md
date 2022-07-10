## 配置说明
```ini
;实例恢复时的回调地址
on_recover_http_url = ""
;实例异常时的回调地址
on_fail_http_url = ""
;http监听地址，端口必须为raft.addr+1
http_addr="127.0.0.1:7001"

;raft配置
[raft]
;raft节点序列号，每一个节要不一样
id=1
;raft集群
nodes = "1/127.0.0.1:7000,2/127.0.0.1:8000,3/127.0.0.1:9000"
;raft节点通信地址
addr="127.0.0.1:7000"
;raft日志、数据目录
dir="/tmp/node/raft_"

;日志配置
[log]
output = console
path = ./logs
filename = health-check
level = Trace
service = health-check
format = plain
```