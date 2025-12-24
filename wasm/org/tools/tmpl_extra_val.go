package tools

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"text/template/parse"
)

const (
	TemplateDataStringFormat = `'%s'`
	TemplateDataNumberFormat = `%v`
	TemplateDataOtherFormat  = `'%v'`
)

// GOOS=js GOARCH=wasm go build -o tmpl_extra_val.wasm tmpl_extra_val.go

// ============================================ common ===============================	==================

// RemoveElements A切片移除存在于B切片元素
// 参数:
//   - sliceA: A切片
//   - sliceB: B切片
//
// 返回值:
//   - 返回: 新A切片
func RemoveElements[T comparable](sliceA []T, sliceB []T) []T {
	// 创建一个 map 来存储 sliceB 中的元素
	elementsToRemove := make(map[T]struct{})
	for _, elem := range sliceB {
		elementsToRemove[elem] = struct{}{}
	}

	// 创建一个新的切片来存储结果
	var result []T
	for _, elem := range sliceA {
		if _, found := elementsToRemove[elem]; !found {
			result = append(result, elem)
		}
	}

	return result
}

func InjectFunc() template.FuncMap {
	return template.FuncMap{
		// MBQL用法

		// 1. 权限
		"PowerList": PowerList,

		// 新版模板语法内置函数

		// 规范化单参数
		"N1V": Normalize1Val,
		// 规范化数组
		"N1Arr": Normalize1Arr,
		// 不使用反射
		"N1Arr2": Normalize1ArrNoReflect,

		// 默认值处理
		"dfv": DefaultValue,
	}
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
		// 处理 interface{} 类型，判断是否为零值
		v := reflect.ValueOf(expect)
		if v.IsZero() {
			return defaultVal
		}
	}
	return expect
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

// Normalize1ArrNoReflect 用法: {{N1Arr .Arr}} , 目的: 将1数组值, 规范化为模板SQL的参数值
// e.g. string : 'a','b','c' ; number/bool: a,b,c ; other: 'a','b','c'
// 参数:
//   - arr: 数组
//
// 返回值:
//   - 规范化(SQL)参数值
func Normalize1ArrNoReflect(arr any) string {

	// 将输入数组转换为 []interface{}
	interfaceArr := toInterfaceSlice(arr)
	if interfaceArr == nil {
		return ""
	}

	normalizeArr := make([]string, len(interfaceArr))
	for i, item := range interfaceArr {
		normalizeArr[i] = Normalize1Val(item)
	}
	// string : 'a','b','c' ; num/bool: a,b,c ; other: 'a','b','c'
	return strings.Join(normalizeArr, ",")
}

// toInterfaceSlice 将不同类型的切片转换为 []interface{}
func toInterfaceSlice(arr any) []any {
	switch v := arr.(type) {
	case []string:
		return convertToInterfaceSlice(v)
	case []int:
		return convertToInterfaceSlice(v)
	case []int32:
		return convertToInterfaceSlice(v)
	case []int64:
		return convertToInterfaceSlice(v)
	case []float32:
		return convertToInterfaceSlice(v)
	case []float64:
		return convertToInterfaceSlice(v)
	case []bool:
		return convertToInterfaceSlice(v)
	case []any:
		return v
	default:
		return nil
	}
}

func convertToInterfaceSlice[T any](arr []T) []any {
	result := make([]any, len(arr))
	for i, item := range arr {
		result[i] = item
	}
	return result
}

// =====================================================================================

// ExtraValFromTemplate 从模板字符串中提取.VAL变量
// 参数:
//   - tmplText: 模板字符串
//
// 返回值:
//   - 返回: 变量名
//   - 返回: 异常
func ExtraValFromTemplate(
	tmplText string) ([]string, error) {
	// 创建模板画布, 注册所有自定义函数, 解析模板字符串
	tmpl, err := template.New("temp").Funcs(InjectFunc()).Parse(tmplText)
	if err != nil {
		return nil, err
	}
	// 获取模板名称
	tmplNames, err := AnalyzeTmplByTemplate(tmpl)
	if err != nil {
		return nil, err
	}
	// 提取的 .val 包含模板名称引用
	vars := ExtractVariables(tmpl)
	// 排除模板名称变量
	newVars := RemoveElements(vars, tmplNames)
	// 变量名称顺序
	sort.Strings(newVars)

	return newVars, nil
}

// AnalyzeTmplByTemplate 解析模板字符串, 提取子模板和主模板名称
// 参数:
//   - templateStr: SQL模板字面量
//
// 返回值:
//   - 返回一个 主从模板名称切片
//   - 返回一个 异常
func AnalyzeTmplByTemplate(tmpl *template.Template) ([]string, error) {
	var tmplSlice []string

	for _, t := range tmpl.Templates() {
		// 当前模板名称
		tmplName := t.Name()
		if isMasterSlaveTemplate(tmplName) {
			tmplSlice = append(tmplSlice, tmplName)
			continue
		}
	}

	return tmplSlice, nil
}

func isMasterSlaveTemplate(tmplName string) bool {
	return strings.HasPrefix(strings.ToLower(tmplName), "slave") || strings.HasPrefix(strings.ToLower(tmplName), "master")
}

func ExtractVariables(t *template.Template) []string {
	var vars []string
	if t.Tree == nil {
		return vars
	}
	for _, node := range t.Tree.Root.Nodes {
		vars = append(vars, findVariables(node)...)
	}
	for _, tmpl := range t.Templates() {
		if tmpl.Tree != nil {
			for _, node := range tmpl.Tree.Root.Nodes {
				vars = append(vars, findVariables(node)...)
			}
		}
	}
	return unique(vars)
}

func findVariables(node parse.Node) []string {
	var vars []string
	switch n := node.(type) {
	case *parse.ActionNode:
		vars = append(vars, findVariablesFromPipe(n.Pipe)...)
	case *parse.BranchNode:
		vars = append(vars, findVariablesFromPipe(n.Pipe)...)
		for _, node := range n.List.Nodes {
			vars = append(vars, findVariables(node)...)
		}
		if n.ElseList != nil {
			for _, node := range n.ElseList.Nodes {
				vars = append(vars, findVariables(node)...)
			}
		}
	case *parse.RangeNode:
		vars = append(vars, findVariablesFromPipe(n.Pipe)...)
	case *parse.WithNode:
		vars = append(vars, findVariablesFromPipe(n.Pipe)...)
	case *parse.IfNode:
		vars = append(vars, findVariablesFromPipe(n.Pipe)...)
	case *parse.TemplateNode:
		// Handle template inclusion
		vars = append(vars, n.Name)
	case *parse.PipeNode:
		vars = append(vars, findVariablesFromPipe(n)...)
	}
	return vars
}

func findVariablesFromPipe(pipe *parse.PipeNode) []string {
	var vars []string
	for _, cmd := range pipe.Cmds {
		for _, arg := range cmd.Args {
			switch arg := arg.(type) {
			case *parse.FieldNode:
				vars = append(vars, arg.Ident[0])
			case *parse.PipeNode:
				vars = append(vars, findVariablesFromPipe(arg)...)
			}
		}
	}
	return vars
}

func unique(vars []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, v := range vars {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
