# 项目编译构建

## 安装 buf

> 注意buf和grpc版本, 可能导致生成proto代码版本不兼容
> go install github.com/bufbuild/buf/cmd/buf@latest

## 更新grpc依赖

> buf dep update

## 编译 proto
> buf generate
> 生成 gen/proto/*.go

## golang依赖编译

> go mod tidy

# 打包

> go build

## 备注

> buf 使用参见

```
https://github.com/forhsd/mbpb/blob/main/entrypoint.proto
```

## 运行cmd二进制

1. cmd/ui, 运行需要gcc, 安装如下:
```shell
sudo apt-get install libxxf86vm-dev
```


## tinygo 安装

> tinygo 编译 WASM 对编码极其不友好, 疑似使用case-type编码为反射, tinygo又不支持, 放弃使用tinygo
> 参考以下文档:
> 
> https://tinygo.org/docs/guides/webassembly/wasm/
> https://go.dev/wiki/WebAssembly#
> https://pkg.go.dev/syscall/js

```
# 下载 tinygo0.34.0.linux-amd64.tar.gz 并解压
https://github.com/tinygo-org/tinygo

# /etc/profile 配置全局环境变量
export TINYGO_HOME=/home/wch/devlop/tinygo
export PATH=$PATH:$TINYGO_HOME/bin

# source 命令只能终端生效, 因此重启OS, 检查是否生效
source /etc/profile
tinygo version
go env
tinygo env

# os系统用户根目录执行, 生成前端wasm解释器. 注意: 目录和文件名不能变
# cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js .

# Makefile 所在目录执行, 编译时需要指定 GOARCH=wasm GOOS=js
make

# app 所在目录执行
go run .

# 打开浏览器, 控制台执行暴露函数
```