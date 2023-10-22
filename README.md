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
其中配置`-c /path/to/relay.xml`是可选，如果不提供配置文件，则使用默认配置。配置文件的格式参考`cfg/relay-example.xml`。

## 验证
向`relay`申请中继需要验证，验证使用的`username/password`有两种配置方式，默认通过配置文件配置，请参考`cfg/relay-example.xml`。

另一种使用`sqlite3`存储，所使用的数据库文件通过配置文件指定，是否启用数据库存储验证信息也是通过配置文件配置。第一次启动`relay`会自动创建该数据库文件。后续可以使用任一支持`sqlite3`的工具添加删除用户。

## 管理
计划添加一个HTTP的管理页面，可以添加、删除账户，显示各种统计信息，比如每条中继连接的速度、使用时间。

因为作者不熟前端，该计划暂时搁置，只实现了几个查询、添加、删除用户的HTTP POST接口。详情可以参考`tests`目录下的`*.http`文件，或者查看源码`internal/mgr/mgr.go`。