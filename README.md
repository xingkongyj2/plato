# plato 
一款纯go编写的，支持亿级通信的IM系统。
https://hardcore.feishu.cn/wiki/wikcnRfpMp8DUAxp8AtKAEF7Hng




编译：go bulid



# 命令行
客户端：./plato client

*配置ip：./plato ipconf

*网关：./plato gateway

perf：./plato perf

*state：./plato state


# 项目启动
./plato gateway --config=./plato.yaml
./plato state --config=./plato.yaml
./plato client --config=./plato.yaml

# 性能测试
./plato perf --config=./plato.yaml

# 接口
api.go::NewChat
