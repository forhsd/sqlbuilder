package facade

import (
	"bytes"
	"fmt"
	"testing"
)

func BenchmarkStringOperation1(b *testing.B) {
	b.ResetTimer()
	str := ""
	for i := 0; i < b.N; i++ {
		str += "golang"
	}
}

// BenchmarkStringOperation2-4   	   59440	    114643 ns/op
// 表示使用4个goroutine并发执行测试;	59440：表示测试执行了59440次迭代;	114643 ns/op：每次操作平均耗时114643纳秒
func BenchmarkStringOperation2(b *testing.B) {
	b.ResetTimer()
	str := ""
	for i := 0; i < b.N; i++ {
		str = fmt.Sprintf("%s%s", str, "golang")
	}
}

// 字符串拼接，下面性能最好而且差距最大, 但是 fmt.Sprintf 可读性最高, 视情况取舍
func BenchmarkStringOperation3(b *testing.B) {
	b.ResetTimer()
	strBuf := bytes.NewBufferString("")
	for i := 0; i < b.N; i++ {
		strBuf.WriteString("golang")
	}
}
