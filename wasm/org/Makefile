build:
	# go 编译 wasm
	GOARCH=wasm GOOS=js go build -o ../out/index_go.wasm wasm_main.go
	# tinygo 编译 wasm, 限制reflect使用
	#GOARCH=wasm GOOS=js tinygo build -o ../out/index_tinygo.wasm wasm_main.go