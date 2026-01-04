package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
	"text/template"

	"github.com/forhsd/sqlbuilder/common/facade"
)

const TestText = `
		{{.Val1}}
		{{AFunc    .VB2  }}
		{{    .VB3    }}
		{{    .VB4_AA}} {{/* - 作为变量名会报错, 约束变量名: 英文字母下划线 */}}
		{{.VB4_AA1 }}
		{{BFunc (AFunc .VB1A .VB2B )}}
		.Val1
		{{range .Items}}{{.Name}}{{end}}  // .Items.Name
		{{with .Person}}{{.Name}}{{end}}
		{{if .PAYMENT_TYPE}} {{.NameC}}{{end}}
	`

// 通过正则提取, 存在问题: 1. 误伤 {{xxx "www.baidu.com"}}
func TestTemplateRegex(t *testing.T) {

	// 或者 \{\{(?:.*?){0,}(?<NAME>\.[a-z0-9]+)(?:.*?){0,}\}\} , 都会将嵌套自定义函数非首个参数丢掉,不清楚原因
	re := regexp.MustCompile(`{{[^}]*?(\.\w+)[^}]*}}`)
	matches := re.FindAllStringSubmatch(TestText, -1)

	for _, match := range matches {
		for i := 1; i < len(match); i++ {
			fmt.Println(match[i])
		}
	}
}

func TestExtraNodeVariable(test *testing.T) {

	data, err := os.ReadFile(filepath.Join("../assets/trans2.tmpl"))
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	content := string(data)

	vars, _ := facade.ExtraValFromTemplate(content)
	sort.Strings(vars)

	for _, v := range vars {
		fmt.Println(v)
	}
}

type Specifications struct {
	SourceFrom string `json:"source"`
	Schema     string `json:"schema"`
	Table      string `json:"table"`
	Database   string `json:"db"`
}

// 定义模板
const ExtraTableText = `
		{{/* - 局部变量名称可以改变, dictCvt 元素的key不能改变 */}}
		{{- $元素1 := dictCvt "table" "t1" "schema" "schema1" "db" "ods" "source" "1" -}}
		{{- $元素2 := dictCvt "table" "t2" "schema" "schema2" "db" "ods" "source" "2" -}}
		{{- $数组变量 := arrCvt (jsonCvt $元素1) (jsonCvt $元素2) -}}
		{{/* - 下面一行除了局部变量名称可以改变,逻辑不能改变 */}}
		[{{- range $i, $e := $数组变量 }}{{ if ne $i 0 }}, {{ end }}{{ $e }}{{ end }}]
	`

func InjectExtractTablesFunc() template.FuncMap {
	return template.FuncMap{
		"arrCvt":  ArrConvert,
		"dictCvt": DictConvert,
		"jsonCvt": JsonConvert,
	}
}

func ArrConvert(values ...string) []string {
	return values
}

func DictConvert(keysAndValues ...any) map[string]any {
	if len(keysAndValues)%2 != 0 {
		return nil
	}
	m := make(map[string]any)
	for i := 0; i < len(keysAndValues); i += 2 {
		m[keysAndValues[i].(string)] = keysAndValues[i+1]
	}
	return m
}

func JsonConvert(data any) string {
	b, _ := json.Marshal(data)
	return string(b)
}

func TestExtractTables(test *testing.T) {
	tmpl, err := template.New("example").Funcs(InjectExtractTablesFunc()).Parse(ExtraTableText)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// 创建一个 bytes.Buffer 来作为 io.Writer
	var buf bytes.Buffer

	// 执行模板并将结果输出到 bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	// 输出模板渲染的原始结果
	fmt.Println("Rendered JSON Array:")
	fmt.Println(buf.String())

	// 假设渲染出的结果是 JSON 数组，我们可以反序列化它并填充结构体
	var specifications []Specifications
	err = json.Unmarshal(buf.Bytes(), &specifications)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// 遍历并打印每个 Specifications 结构体的内容
	fmt.Println("\nSpecifications Structs:")
	for _, spec := range specifications {
		fmt.Printf("Table: %s, Schema: %s, Database: %s, Source: %s\n", spec.Table, spec.Schema, spec.Database, spec.SourceFrom)
	}
}
