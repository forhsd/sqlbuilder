package facade

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"text/template"
)

func TestTemplateString(t *testing.T) {
	strA := `"%s"."%s" as temp`

	// 填充两个占位符, 返回填充后的字符串
	finalStr := fmt.Sprintf(strA, "a", "b")

	expected := `"a"."b" as temp`

	if finalStr != expected {
		t.Errorf("Expected %s, but got %s", expected, finalStr)
	}
}

type Inventory struct {
	Material    string
	Count       uint
	TestStrArrs []string
	TestIntArrs []int
	Empty       string
}

/**
局部变量 $var 可以与多种模板操作一起使用，以增强模板的灵活性和表达力。常见的用法包括：
	{{define}}：定义和使用模板块内的局部变量(解构data, 用作某个属于的变量接收体)。
	{{with}}：在新的上下文中定义和使用局部变量(解构data, 用作某个属于的变量接收体)。
	{{range}}：在循环中定义和使用局部变量(解构 index,item & 解构data)。
	{{if}} 和 {{else}}：在条件判断块中定义和使用局部变量(解构data, 用作某个属于的变量接收体)。
*/

const templateString = `
	{{.Count}} of {{.Material}}
	{{if eq .Count  17 }} T1 {{else}} T0 {{end}}
	{{if gt .Count  17}} T1 {{else if le .Count  17}} T0 {{end}}
	{{"\"output\""}}
	{{len .Material}}
	{{index .Material 0 }} {{/* 返回索引指向值, 字符串返回字符ASCII值 */}}
	{{Greet .Material}}

	{{"测试range, 相当于forr" | printf "%q"}}
	{{range $index, $element := .TestStrArrs}}
	  Element {{$index}}: {{$element}}
	{{else}}
	  No elements
	{{end}}

	{{"测试注入结构体" | printf "%q"}} {{/* %q 字符串字面值，必要时会采用安全的转义表示 */}}
	{{.}} {{/* .代表Execute注入内容 */}}

	{{"测试申明处理局部变量" | printf "%q"}}
	   {{- $total := 0 -}}  
	{{- range .TestIntArrs -}}
	  {{- $total = Add $total . -}} {{/* . 这里代表循环item,在最外层和局部代表的内容是不一样的; 大致理解.代表注入模板或局部的整个结构体 */}}
	{{- end }} {{/* - 代表前后的空白和换行符和制表符 */}}

	Total: {{$total}}

	{{/* $val 1.作为局部变量,常和搭配 2. 作为data和模板内局部变量的桥梁赋值 */}}
	{{- $bridgeVal := .Material -}}
	{{- $partVal := "局部变量!" -}}

	bridgeVal: {{ $bridgeVal }}
	partVal: {{ $partVal }}


	{{"模板套模板定义,开始" | printf "%q"}}
	{{define "header"}} 
	Header: {{.}} {{/* . 代表传入名为header的template的入参值  */}}
	{{end -}}
	
	{{define "body"}} 
	Body: {{.}} 
	{{end -}}

	{{define "footer"}} 
	{{template "header" "嵌套内 header"}}
	{{template "body" "嵌套内 body"}}
	Footer: {{.}} 
	{{end -}}
	{{"模板套模板定义,结束" | printf "%q"}}
	
	{{/* define 代表定义但不执行翻译模板, template代表调用执行翻译模板, 由于[模板无限嵌套内]有执行, 因此环绕代码需要挂载在template上下  */}}
	{{"模板套模板,执行开始" | printf "%q"}}
	{{- template "header" .Material}}
	{{- template "body" .TestStrArrs}}
	{{"模板套模板定义,执行结束" | printf "%q"}}

	{{"模板+无限嵌套开始" | printf "%q"}}
	{{- template "footer" .TestIntArrs}}
	{{"模板+无限嵌套结束" | printf "%q"}}

	{{"存在判断" | printf "%q"}}
	{{with .Empty}} {{.}} {{else}} T0 {{end}}
	{{/* with判断接下来表达式是否是非空值, 非空执行T1,为空执行T0, 同时会对指定变量重新指定临时变量并赋值,方法体内的{{.}}此时代表.Empty */}}

	{{"使用逻辑运算处理零值判断" | printf "%q"}}
	{{if and .Material .Empty}}
		Both conditions are true!
	{{else}}
		At least one condition is false.
	{{end}}

	{{"嵌套自定义方法" | printf "%q"}}

	{{BFunc (AFunc .Material)}}
`

const aaa = "{{`\"output\"`}}"

// 标准库内没有的函数需要自定义,并且在指定模板上FuncMap中注册
func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func add(a, b int) int {
	return a + b
}

func aFunc(str string) string {
	return str + "a"
}

func bFunc(str string) string {
	return str + "b"
}

func TestTemplate(t *testing.T) {
	testStrArrs := []string{"apple", "banana", "cherry"}
	testIntArrs := []int{1, 2, 3}
	sweaters := Inventory{
		Material:    "wool",
		Count:       17,
		TestStrArrs: testStrArrs,
		TestIntArrs: testIntArrs,
		Empty:       "aa",
	}

	// 申明模板, 模板注册函数
	tmpl, err := template.New("测试模板").Funcs(template.FuncMap{
		"Greet": greet,
		"Add":   add,
		"AFunc": aFunc,
		"BFunc": bFunc,
	}).Parse(templateString)
	if err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}

	var buf bytes.Buffer
	// 模板注入参数; ExecuteTemplate
	err = tmpl.Execute(&buf, sweaters)
	if err != nil {
		t.Fatalf("Error executing template: %v", err)
	}

	fmt.Printf("%s\n", buf.String())
}

// 场景: @_start @_end mbql替代方案: 使用模板定义嵌套ExecuteTemplate(*N), 再将其解析后的 &bufNest 作为参数注入Execute(&buf, argsMap)的argsMap
const templateContent = `
	{{define "T1"}}Hello, {{.Name}}!{{end}}
	{{define "T2"}}{{template "T1" .}} Welcome to {{.Place}}!{{end}}
`

func TestExecuteTemplate(t *testing.T) {
	// 解析模板
	tmpl, err := template.New("templates").Parse(templateContent)
	if err != nil {
		panic(err)
	}

	// 创建数据结构
	data := struct {
		Name  string
		Place string
	}{
		Name:  "Alice",
		Place: "GoLand",
	}

	// 执行模板 "T1" 并获取其输出
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "T1", data)
	if err != nil {
		panic(err)
	}

	// 将 "T1" 的输出作为数据传递给 "T2"
	t2Data := struct {
		InnerOutput string
		Place       string
		Name        string
	}{
		Name:        "Alice",      // 同一个 tmpl, T2没用 .Name 但依赖的T1用了, 因此必须注入
		InnerOutput: buf.String(), // 使用 "T1" 的输出
		Place:       data.Place,
	}

	// 解析模板 "T2"
	var buf2 bytes.Buffer
	// 注意!!!!!: 由于使用同一个template对象, 因此需要把所有模板用到的变量参数注入
	err = tmpl.ExecuteTemplate(&buf2, "T2", t2Data)
	if err != nil {
		panic(err)
	}

	fmt.Println(buf2.String()) // 输出: Hello, Alice! Welcome to GoLand!
}

// 同一个外层 tmpl, 测试内层buf追加

const templateBufContext = `
	{{define "index"}}
		<html>
		<head><title>{{.Title}}</title></head>
		<body>
		<h1>{{.Heading}}</h1>
	{{end}}

	{{define "about"}}
		<html>
		<head><title>{{.Title}}</title></head>
		<body>
		<h1>About Us</h1>
		<p>{{.Content}}</p>
	{{end}}

	{{define "context"}}
		{{.AAA}}
		上下文,模板1:
			{{template "index" .}}
		上下文,模板2:
			{{template "about" .}}
	{{end}}
`

func TestBufAppend(t *testing.T) {
	// 解析模板
	tmpl, err := template.New("context").Parse(templateBufContext)
	if err != nil {
		panic(err)
	}

	// 创建数据
	data := struct {
		Title   string
		Heading string
		Content string
		AAA     string
	}{
		Title:   "Welcome",
		Heading: "Hello, World!",
		Content: "This is the about page content.",
		AAA:     "BBB",
	}

	// 执行模板 "context" 并将数据传递给它
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "context", data)
	if err != nil {
		panic(err)
	}

	// 如果像下面这样没有执行解析 context 模板, 那么只有 index 和 about 会被记载出来, 属于 context 模板 的 {{.AAA}}和两行中文不会加载解析

	// 执行模板 "index"
	//var buf bytes.Buffer
	//err = tmpl.ExecuteTemplate(&buf, "index", data)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// 清空缓冲区 => 不清空可以一直追加
	////buf.Reset()
	//
	//// 执行模板 "about"
	//err = tmpl.ExecuteTemplate(&buf, "about", data)
	//if err != nil {
	//	panic(err)
	//}

	fmt.Println(buf.String())
}

// 模拟 @_start @_end

const templateBufImitate = `

{{/* 优先执行块内容, 可以写入文件, 利用ParseFiles 和 ParseGlob 读取文件作为一个append模板解析 */}}

{{define "slave1"}}
	<html>
	<head><title>{{.Title}}</title></head>
	<body>
	<h1>{{.Heading}}</h1>
{{end}}

{{define "slave2"}}
	<html>
	<head><title>{{.Title}}</title></head>
	<body>
	<h1>About Us</h1>
	<p>{{.Content}}</p>
{{end}}

{{define "master"}}
	{{.AAA}}
	上下文,模板1:
		{{.Slave1Content}}
	上下文,模板2:
		{{.Slave2Content}}
{{end}}
`

func TestBufImitate(t *testing.T) {
	// 解析所有模板
	tmpl, err := template.New("all").Parse(templateBufImitate)
	if err != nil {
		panic(err)
	}

	// 创建数据
	data := struct {
		Title   string
		Heading string
		Content string
		AAA     string
	}{
		Title:   "标题 Welcome",
		Heading: "表头 Hello, World!",
		Content: "正文 This is the about page content.",
		AAA:     "附加内容 This is some additional context.",
	}

	// 执行模板 "index" 并获取其内容
	var indexBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&indexBuf, "slave1", data)
	if err != nil {
		panic(err)
	}

	// 执行模板 "about" 并获取其内容
	var aboutBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&aboutBuf, "slave2", data)
	if err != nil {
		panic(err)
	}

	// 创建数据结构，包含从 "index" 和 "about" 模板获取的内容
	contextData := struct {
		Title         string
		Heading       string
		Content       string
		AAA           string
		Slave1Content string
		Slave2Content string
	}{
		Title:         data.Title,
		Heading:       data.Heading,
		Content:       data.Content,
		AAA:           data.AAA,
		Slave1Content: indexBuf.String(),
		Slave2Content: aboutBuf.String(),
	}

	// 执行模板 "context"
	var contextBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&contextBuf, "master", contextData)
	if err != nil {
		panic(err)
	}

	// 去除: {{define "context"}}{{end}} 可以使用以下方式输出最终内容; 但是会导致上面一大块申明内容被视为空白行
	// 如果希望减少根据模板生成SQL无效行,最好在最终外层sql加上模板定义申明
	//err = tmpl.Execute(&contextBuf, contextData)
	//if err != nil {
	//	panic(err)
	//}

	fmt.Println(contextBuf.String())
}

// extractDefinitions 从模板字符串中提取定义
func extractDefinitions(templateStr string) (map[string]string, map[string]string, error) {
	definitions := make(map[string]string)
	// master 模板需要向 data 添加额外提前执行模板的数据
	masterDefinitions := make(map[string]string)

	// 提前执行模板的数据
	subTmplData := map[string]any{
		"Title":   "标题 Welcome",
		"Heading": "表头 Hello, World!",
		"Content": "正文 This is the about page content.",
		"AAA":     "附加内容 This is some additional context.",
	}

	// 创建一个新的模板对象
	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		return nil, nil, err
	}

	// 遍历模板定义
	for _, t := range tmpl.Templates() {
		// 获取模板内容
		buf := new(strings.Builder)
		// 只保留模板名称slave开始的模板
		if !strings.HasPrefix(strings.ToLower(t.Name()), "slave") {
			continue
		}

		err := t.Execute(buf, subTmplData)
		if err != nil {
			return nil, nil, err
		}
		definitions[t.Name()] = buf.String()
	}

	// 下面两组值相当于执行 @_start @_end 解析得到的sql
	subTmplData["Slave1Content"] = `
			<html>
			<head><title>标题 Welcome</title></head>
			<body>
			<h1>表头 Hello, World!</h1>
	`
	subTmplData["Slave2Content"] = `
			<html>
			<head><title>标题 Welcome</title></head>
			<body>
			<h1>About Us</h1>
			<p>正文 This is the about page content.</p>
	`

	for _, t := range tmpl.Templates() {
		// 获取模板内容
		buf := new(strings.Builder)
		// 只保留模板名称slave开始的模板
		if !strings.HasPrefix(strings.ToLower(t.Name()), "master") {
			continue
		}

		err := t.Execute(buf, subTmplData)
		if err != nil {
			return nil, nil, err
		}
		masterDefinitions[t.Name()] = buf.String()
	}

	return definitions, masterDefinitions, nil
}

func TestExtraFromTemplateString(t *testing.T) {
	// 提取模板定义
	definitions, masterDefinitions, err := extractDefinitions(templateBufImitate)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 输出模板定义
	for name, content := range definitions {
		fmt.Println("slave 打印内容--模板名称:", name)
		fmt.Println("slave 打印内容--模板内容:", content)
		fmt.Println()
	}

	for name, content := range masterDefinitions {
		fmt.Println("master 打印内容--模板名称:", name)
		fmt.Println("master 打印内容--模板内容:", content)
		fmt.Println()
	}
}

const sqlTemplate = `
	SELECT 
		team.username 姓名,
		userdetail.user_name, 
		team.user_identification account, 
		team.tiname 部门,
		team.triname 岗位, 
		userdetail.project_name 项目, 
		userdetail.types 类型,
		CASE 
			WHEN userdetail.severities = '-' THEN '' 
			ELSE userdetail.severities 
		END 严重程度, 
		userdetail.num 数量
	FROM (
		SELECT
			id,
			account,
			user_name,
			project_name,
			types,
			severities,
			num 
		FROM "sonar".tb_user_detailquality detail
		WHERE 
			severities IN ('LOW', 'MEDIUM', 'HIGH')
			{{- if .ProjectName }}
			AND project_name = ANY({{.ProjectName}}::text[])
			{{- end }}
			{{- if .TypeCode }}
			AND types = ANY({{.TypeCode}}::text[])
			{{- end }}
			{{- if .SeveritiesType }}
			AND severities = ANY({{.SeveritiesType}}::text[])
			{{- end }}
			{{- if .StartTime }}
			AND detail.create_time = {{.StartTime}}
			{{- end }}
	) userdetail
	INNER JOIN sonar.tb_mapping_user_structure mapuser 
		ON mapuser.external_part_user_id = userdetail.user_name
		AND mapuser.system_type = 1 
		AND mapuser.use_state = 1 
	RIGHT JOIN (
		SELECT 
			u.username,
			tui.user_identification,
			ti.name AS tiname,
			tri.name AS triname
		FROM 
			team.team_user_info tui
			JOIN team.team_role_info tri ON tui.main_post_id = tri.id  
			JOIN team.USER u ON tui.user_identification = u.account 
			JOIN team.team_info ti ON tri.team_info_id = ti.id
		WHERE 
			tui.team_info_id = 244
			{{- if .Husername }}
			AND u.username = ANY({{.Husername}}::text[])
			{{- end }}
	) team 
		ON mapuser.holder_id = team.user_identification
	ORDER BY 
		2, 4
`

type SQLParams struct {
	ProjectName    string
	TypeCode       string
	SeveritiesType string
	StartTime      string
	Husername      string
}

func TestTemplateDashboard2330(t *testing.T) {
	params := SQLParams{
		//ProjectName:    `{"Project1", "Project2"}`,
		//TypeCode:       `{"Type1", "Type2"}`,
		SeveritiesType: `{MEDIUM}`,
		StartTime:      `'2024-04-01'`,
		Husername:      `{王成皓}`,
	}

	tmpl, err := template.New("sql").Parse(sqlTemplate)
	if err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	if err != nil {
		t.Fatalf("Error executing template: %v", err)
	}

	fmt.Printf("%s\n", buf.String())
}

// 接口类型断言检查
var (
	_ CCC = (*AAA)(nil)
	_ CCC = (*BBB)(nil)
)

type CCC interface {
	do()
}

type AAA struct {
}

type BBB struct {
}

func (a *AAA) do() {
	fmt.Println("aaa")
}

func (b *BBB) do() {

	fmt.Println("bbbb")
}

func TestTemplateDashboard230(t *testing.T) {
	var a CCC
	a = &BBB{}
	switch a.(type) {
	case *BBB:
		fmt.Printf("bbb")
	case *AAA:
		fmt.Println("bbb")
	}
}
