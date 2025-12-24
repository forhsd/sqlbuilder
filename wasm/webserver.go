package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const dir = "./out"

func main() {
	// 注意: 当前目录命令行执行[go run .], 则需要将当前目录作为工作目录;
	// 而直接点击 IDE 的 main函数绿色三角, 则默认是以项目根目录作为工作目录, 导致无法找到文件, 因此需要手动指定工作目录
	fs := http.FileServer(http.Dir(dir))
	log.Print("Serving " + dir + " on http://localhost:8080")
	err := http.ListenAndServe(":8080", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(resp, req)
	}))
	if err != nil {
		fmt.Println(err)
	}
}
