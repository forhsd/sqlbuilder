package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	p1 = `
输入 Go text/template 模板...

	用法说明:
	1. 标签名称输入字符串字面量, e.g. STORE_GUID 或 Guids 等
	
	2. 标签值(输入标签值后选择必须选择对应的类型下拉选项):
	- 字符串 
	e.g.  2024-09-05 17:17:37 或 样例文字 或 2203b12d-618a-4974-b09f-86ce3d35d26a ; 
	选择字符串下拉
	- 数值 e.g. 10.145 或 -1 ; 
	选择数值下拉
	- 字符串数组(字面量[]内根据,切割) 
	e.g. [2024-07-05 17:17:37, 2024-09-05 17:17:37] ; 
	选择字符串数组下拉
	- 数字数组(字面量[]内根据,切割) 
	e.g. [1, 2] ; 选择数值数组下拉
	
	3. 一组标签名称&标签值&下拉选项 就绪后, 点击添加, 
	如果填错选择指定键值对后面的删除, 或者清空所键值对
	
	4. 一组标签名称&标签值&下拉选项&模板sql 就绪后, 点击查询
	`
)

// GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o ./TemplateTransformer.exe ./main.go
// GOOS=windows GOARCH=amd64 go build -o ./TemplateTransformer.exe ./main.go
func main() {
	a := app.New()
	w := a.NewWindow("text/template模板翻译(输入kv选择正确下拉选项)")
	w.Resize(fyne.NewSize(1024, 768))

	// 存储键值对的 map
	data := make(map[string]any)

	// 输入框组件
	keyEntry := widget.NewEntry() // 键
	keyEntry.SetPlaceHolder("标签名称, key")
	valueEntry := widget.NewEntry() // 值输入框
	valueEntry.SetPlaceHolder("标签标签值, value")

	// 下拉框选择数据类型
	valueTypeSelect := widget.NewSelect([]string{"数值", "字符串", "数值数组", "字符串数组"}, nil)
	valueTypeSelect.SetSelected("字符串") // 默认值

	// 显示当前数据的列表（带删除按钮）
	dataList := container.NewVBox()

	// 先声明变量
	var updateDataList func()

	// 更新数据列表
	updateDataList = func() {
		dataList.Objects = nil
		for k, v := range data {
			keyLabel := widget.NewLabel(fmt.Sprintf("%s: %v", k, v))
			deleteButton := widget.NewButton("删除", func(k string) func() {
				return func() {
					delete(data, k)
					updateDataList()
				}
			}(k)) // 解决闭包问题
			dataList.Add(container.NewHBox(keyLabel, deleteButton))
		}
		dataList.Refresh()
	}

	// 添加键值对按钮
	addButton := widget.NewButton("添加", func() {
		key := keyEntry.Text
		valType := valueTypeSelect.Selected
		val := parseInput(valueEntry.Text, valType)

		if key != "" {
			data[key] = val
			updateDataList()
		}
		keyEntry.SetText("")
		valueEntry.SetText("")
	})
	addButton.Importance = widget.SuccessImportance

	// 清空按钮
	clearButton := widget.NewButton("清空", func() {
		data = make(map[string]any)
		updateDataList()
	})
	clearButton.Importance = widget.DangerImportance

	///////////////////////////////////// 以上不变 //////////////////////////////////////////////

	// 模板输入框
	templateEntry := widget.NewMultiLineEntry()
	templateEntry.SetPlaceHolder(p1)
	templateEntry.Wrapping = fyne.TextWrapWord

	// 结果显示
	resultEntry := widget.NewMultiLineEntry()
	resultEntry.SetPlaceHolder("最终生成sql, 用法: 复制到数据库IDE中排查哪部分模板SQL写错")
	resultEntry.Wrapping = fyne.TextWrapWord

	// 添加灰色背景
	background := canvas.NewRectangle(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	resultContainer := container.NewStack(
		background,
		resultEntry,
	)

	// 带滚动条的结果区
	resultScroll := container.NewVScroll(resultContainer)
	resultScroll.SetMinSize(fyne.NewSize(400, 250))

	// 处理master查询逻辑
	queryMasterButton := widget.NewButton("查询", func() {
		renderedSQL := processTemplate(templateEntry.Text, data)
		resultEntry.SetText(renderedSQL)
	})
	queryMasterButton.Importance = widget.HighImportance

	queryButtonRow := container.NewVBox(
		layout.NewSpacer(),
		container.NewVBox(addButton, clearButton),
		queryMasterButton,
	)
	rightPanel := container.NewBorder(
		queryButtonRow,
		nil, nil, nil,
		resultScroll,
	)
	// 底部布局 (模板输入框 + 结果展示)
	bottomLayout := container.NewHSplit(
		templateEntry,
		rightPanel,
	)

	bottomLayout.SetOffset(0.5) // 左右平分

	///////////////////////////////////// 以上底部 //////////////////////////////////////////////

	// 顶部输入区
	topInput := container.NewVBox(
		container.NewGridWithColumns(3, keyEntry, valueEntry, valueTypeSelect),
		dataList,
	)

	//// 顶部布局 (输入区 + 按钮)
	topLayout := container.NewGridWithColumns(
		1,
		topInput,
	)

	///////////////////////////////////// 以上顶部 //////////////////////////////////////////////

	// 主窗口布局 ======================================
	content := container.NewBorder(
		topLayout, nil, nil, nil,
		bottomLayout,
	)
	w.SetContent(content)
	w.ShowAndRun()
}

// 解析输入内容（根据选择的类型）
func parseInput(input, valType string) any {
	input = strings.TrimSpace(input)

	switch valType {
	case "数值":
		if num, err := strconv.ParseFloat(input, 64); err == nil {
			return num
		}
	case "字符串":
		return input
	case "数值数组":
		parts := strings.Split(strings.Trim(input, "[]"), ",")
		var nums []int
		for _, p := range parts {
			if num, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				nums = append(nums, num)
			}
		}
		return nums
	case "字符串数组":
		var strArr []any
		parts := strings.Split(strings.Trim(input, "[]"), ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
			strArr = append(strArr, parts[i])
		}
		return strArr
	}
	return input
}

func processTemplate(tmplStr string, data map[string]any) string {
	tmpl, err := template.New("root").Funcs(InjectFunc()).Parse(tmplStr)
	if err != nil {
		return "模板解析错误: " + err.Error()
	}
	slaveMap := getSlaveTemplate(tmpl, data)

	finalData := unionMap(data, slaveMap)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "master", finalData)
	if err != nil {
		return "模板渲染错误: " + err.Error()
	}
	sql := buf.String()

	return sql
}

func getSlaveTemplate(tmpl *template.Template, data map[string]any) map[string]any {
	slaveMap := make(map[string]any)
	for _, t := range tmpl.Templates() {
		// 当前模板名称
		tmplName := t.Name()
		// slava 模板, 需要注入参数
		if strings.HasPrefix(strings.ToLower(tmplName), "slave") {
			buf := new(strings.Builder)
			// 子模板必须先解析
			err := t.Execute(buf, data)
			if err != nil {
				return slaveMap
			}
			slaveMap[t.Name()] = buf.String()
			continue
		}
	}
	return slaveMap
}

func unionMap(m1, m2 map[string]any) map[string]any {
	// 思路：先把其中一个map 放到新的对象中，把m2中key不存在于本对象中合并即可
	result := make(map[string]any)
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		if _, ok := result[k]; !ok {
			result[k] = v
		}
	}
	return result
}

// ====================== InjectFunc =================================

const (
	TemplateDataStringFormat = `'%s'`
	TemplateDataNumberFormat = `%v`
	TemplateDataOtherFormat  = `'%v'`
)

func InjectFunc() template.FuncMap {
	return template.FuncMap{
		// MBQL用法

		"PowerList": PowerList,

		// 新版模板语法内置函数

		// 规范化单参数
		"N1V": Normalize1Val,
		// 规范化数组
		"N1Arr": Normalize1Arr,

		// kv对封装为map
		"dictCvt": DictConvert,

		// map元素转换为json
		"jsonCvt": JsonConvert,

		// json元素封装为切片
		"arrCvt": ArrConvert,

		// 默认值处理
		"dfv": DefaultValue,
	}
}

func PowerList(guids []any) string {
	var tempArr []string
	for i := range guids {
		val, ok := guids[i].(string)
		if ok {
			tempArr = append(tempArr, val)
		}
	}

	powerListStr := strings.Join(tempArr, ",")

	oneStr := fmt.Sprintf(TemplateDataStringFormat, powerListStr)

	return oneStr
}

func PowerListByString(guids []string) string {

	powerListStr := strings.Join(guids, ",")

	oneStr := fmt.Sprintf(TemplateDataStringFormat, powerListStr)

	return oneStr
}

func DefaultValue(expect any, defaultVal any) any {
	if expect == nil {
		return defaultVal
	}
	switch arg := expect.(type) {
	case string:
		// 字符串= '' , 取默认值
		if arg == "" {
			return defaultVal
		}
	case float32, float64, int, int32, int64, bool:
		// 数值是零值, 取默认值
		if arg == 0 {
			return defaultVal
		}
	default:
		// 处理 any 类型，判断是否为零值
		v := reflect.ValueOf(expect)
		if v.IsZero() {
			return defaultVal
		}
	}
	return expect
}

// Normalize1Val 用法: {{N1V .Arg}} , 目的: 将1个参数值, 规范化为模板SQL的参数值
// 参数:
//   - arg: 1个参数值
//
// 返回值:
//   - 规范化(SQL)参数值
func Normalize1Val(arg any) string {
	if nil == arg {
		return ""
	}
	switch arg := arg.(type) {
	case string:
		return fmt.Sprintf(TemplateDataStringFormat, arg)
	case float32, float64, int, int32, int64, bool:
		return fmt.Sprintf(TemplateDataNumberFormat, arg)
	default:
		return fmt.Sprintf(TemplateDataOtherFormat, arg)
	}
}

// Normalize1Arr 用法: {{N1Arr .Arr}} , 目的: 将1数组值, 规范化为模板SQL的参数值
// e.g. string : 'a','b','c' ; number/bool: a,b,c ; other: 'a','b','c'
// 参数:
//   - arr: 数组
//
// 返回值:
//   - 规范化(SQL)参数值
func Normalize1Arr(arr any) string {

	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		return ""
	}

	normalizeArr := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		normalizeArr[i] = Normalize1Val(item)
	}
	// string : 'a','b','c' ; num/bool: a,b,c ; other: 'a','b','c'
	return strings.Join(normalizeArr, ",")
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
