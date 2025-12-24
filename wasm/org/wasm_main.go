package main

import (
	"fmt"
	"syscall/js"

	"github.com/forhsd/sqlbuilder/wasm/org/tools"
)

// tinygo 编译存在两个问题:
// syscall/js.finalizeRef not implemented => go.importObject.env["syscall/js.finalizeRef"] = () => {} 解决
// panic: unimplemented: (reflect.Type).NumOut()
// 问题2: 即使代码不存在反射, 使用 type-case, wasm就会存在对反射的使用, 无法避免

// 注意: go 编译 WASM 支持反射, tinygo	不支持反射
// https://tinygo.org/docs/guides/webassembly/wasm/
// https://go.dev/wiki/WebAssembly#
// https://pkg.go.dev/syscall/js
func main() {
	done := make(chan struct{})
	global := js.Global()
	global.Set("etv", js.FuncOf(extraValsWrapper))
	<-done
}

// 提取模板中的变量名
func extraValsWrapper(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		fmt.Println("ERROR: number of argument doesn't match.")
		return nil
	}

	result, err := tools.ExtraValFromTemplate(args[0].String())
	if err != nil {
		fmt.Println("ERROR: number of argument doesn't match.")
		return nil
	}

	res := make([]interface{}, len(result))
	for i, v := range result {
		res[i] = v
	}

	fmt.Printf("vals : %+v", res)

	// 由于 js.ValueOf 只支持源码 case 类型, 可不包装 js.ValueOf(result), 不接收 []string 等类型, 需要转换成支持类型
	// js.ValueOf(res)
	return res
}
