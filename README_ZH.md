<div >
<h1 align="center">Begonia</h1>  
<center>

[![Go Report Card](https://goreportcard.com/badge/github.com/begonia-org/begonia)](https://goreportcard.com/report/github.com/begonia-org/begonia)
[![codecov](https://codecov.io/github/begonia-org/begonia/graph/badge.svg?token=VGGAA5A87B)](https://codecov.io/github/begonia-org/begonia)

</center>

<center>

[English](README.md) | [中文](README_ZH.md)

</center>
<p>
Begonia 是一个 HTTP 到 gRPC 的反向代理服务器，它基于 protoc 生成的 descriptor_set_out 注册由`gRPC-gateway`定义的服务路由到网关，从而实现反向代理功能。HTTP 服务遵循 RESTful 标准来处理 HTTP 请求，并将 RESTful 请求转发到 gRPC 服务。
</p>
</div>

# 特性

- 兼容所有的`gRPC-gateway`功能特性

- 支持所有 HTTP 请求方式和参数格式转换为 gRPC 请求和数据格式

- 支持将 gRPC 的双向流式传输转换为 websocket 协议进行请求处理

- 支持将服务端流式传输转换为 SSE(Server-Side-Event)协议进行请求处理

- 基于自定义的`application/begonia-client-stream` 请求类型转发 gRPC 的客户端流式请求

- 支持`application/x-www-form-urlencoded`和`multipart/form-data`参数格式的请求

- 丰富的内置中间件，例如 APIKEY 校验、AKSK 校验，`go-playground/validator`参数校验中间件

# 开始

### 安装

```bash
git clone https://github.com/begonia-org/begonia.git
```

```bash
cd begonia && make install
```

### 定义 proto

参考[example/example.proto](example/example.proto)

### 生成 Descriptor Set

```shell
protoc --descriptor_set_out=example.pb --include_imports --proto_path=./ example.proto
```

### 启动网关服务

#### 1、构建运行环境

```bash
docker compose up -d
```

#### 2、初始化数据库

```bash
begonia init -e dev
```

#### 3、启动服务

```bash
begonia start -e dev
```

#### 4、注册服务

```bash
go run . endpoint add  -n "example" -d /data/work/begonia-org/begonia/example/example.pb -p 127.0.0.1:1949  -p 127.0.0.1:2024
```

#### 5、测试请求服务

```
curl -vvv http://127.0.0.1:12138/api/v1/example/hello
```
