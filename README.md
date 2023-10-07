# relay
[![Go](https://github.com/pjlt/relay/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/pjlt/relay/actions/workflows/go.yml)

`Lanthing`的中继服务器。

## 编译
```bash
go build ./cmd/relay
```

## 运行
```bash
relay -c /path/to/relay.xml
```
其中配置`-c /path/to/relay.xml`是可选，如果不提供配置文件，则使用默认配置。配置文件的格式参考`cfg/relay.xml`。

## 验证
向`relay`申请中继需要验证，验证使用的`username/password`存储在`user.db`（可以通过配置文件指定路径）。第一次启动`relay`会自动创建该文件。后续可以使用任一支持`sqlite3`的工具添加删除用户，或者使用`relay`提供的HTTP POST接口进行查询、添加、删除。详情可以参考`tests`目录下的`*.http`文件。