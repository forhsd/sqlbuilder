package facade

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	pb "github.com/forhsd/sqlbuilder/gen/proto"
)

// -------------------------------------------  json ----------------------------------------------

type AnalyzeTemplateRequest struct {
	Tmpl string                 `json:"tmpl"`
	Args map[string]interface{} `json:"args"`
}

type TemplateCtx struct {
	Master string `json:"master"`
	// slave1,sql1; slave2,sql2;
	Slave map[string]string `json:"slave"`
	// 附加内容: 目前只有写SQL显式申明使用哪些表
	Addition string `json:"addition"`
}

type NativeSqlHeader struct {
	Specs []Specifications `json:"specs"`
}

type Specifications struct {
	SourceFrom string `json:"source"`
	Schema     string `json:"schema"`
	Table      string `json:"table"`
	Database   string `json:"db"`
}

// AnalyzeTemplatesByJson 解析模板字符串, 提取子模板和主模板
// 参数:
//   - req: 解析请求体
//
// 返回值:
//   - 返回一个 模板上下文对象
//   - 返回一个 异常
func AnalyzeTemplatesByJson(req AnalyzeTemplateRequest) (*TemplateCtx, error) {
	templateStr := req.Tmpl
	if templateStr == "" {
		return nil, errors.New("模板不能为空")
	}

	return AnalyzeSubTmpl(templateStr, req.Args)
}

// AnalyzeSubTmpl 解析模板字符串, 提取子模板和主模板
// 参数:
//   - templateStr: SQL模板字面量
//   - data: 用户输入参数
//
// 返回值:
//   - 返回一个 模板上下文对象
//   - 返回一个 异常
func AnalyzeSubTmpl(
	templateStr string,
	data map[string]interface{}) (*TemplateCtx, error) {
	// 嵌套模板内容(解析注入参数)模板
	slaveMap := make(map[string]string)
	// 主模板内容(不解析)模板
	var masterMap string
	var addition string

	// 创建一个模板画布(渲染整个模板字符串), 并注册自定义函数
	tmpl, err := template.New("canvas").Funcs(InjectFunc()).Parse(templateStr)
	if err != nil {
		return nil, err
	}

	// 提取 slave 所有模板 和 master模板, 不区分大小写
	for _, t := range tmpl.Templates() {
		// 当前模板名称
		tmplName := t.Name()
		// slava 模板, 需要注入参数
		if strings.HasPrefix(strings.ToLower(tmplName), "slave") {
			buf := new(strings.Builder)
			// 子模板必须先解析
			err := t.Execute(buf, data)
			if err != nil {
				return nil, err
			}
			slaveMap[t.Name()] = buf.String()
			continue
		}

		if strings.HasPrefix(strings.ToLower(tmplName), "master") {
			//err := t.Execute(buf, make(map[string]string))
			//if err != nil {
			//	return nil, err
			//}
			// master 模板, 不解析, 获取模板原始字面量
			masterMap = strings.TrimSpace(t.Tree.Root.String())
			continue
		}

		if strings.HasPrefix(strings.ToLower(tmplName), "addition") {
			// addition 模板, 不解析, 获取模板原始字面量
			addition = strings.TrimSpace(t.Tree.Root.String())
			continue
		}

	}

	return &TemplateCtx{
		Master:   masterMap,
		Slave:    slaveMap,
		Addition: addition,
	}, nil
}

// AnalyzeTmplByTemplateStr 解析模板字符串, 提取子模板和主模板名称
// 参数:
//   - templateStr: SQL模板字面量
//
// 返回值:
//   - 返回一个 主从模板名称切片
//   - 返回一个 异常
func AnalyzeTmplByTemplateStr(templateStr string) ([]string, error) {
	var tmplSlice []string

	// 创建一个模板画布(渲染整个模板字符串), 并注册自定义函数
	tmpl, err := template.New("canvas").Funcs(InjectFunc()).Parse(templateStr)
	if err != nil {
		return nil, err
	}

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

// -------------------------------------------  proto ----------------------------------------------

// AnalyzeTemplatesByProto : 实现层, 解析提取主从模板
// 参数:
//   - req: 解析请求体
//
// 返回值:
//   - 返回 模板上下文对象
//   - 返回 错误
func AnalyzeTemplatesByProto(req *pb.AnalyzeTemplateRequest) (*TemplateCtx, error) {
	templateStr := req.Tmpl
	if templateStr == "" {
		return nil, errors.New("模板不能为空")
	}
	// proto 转换为 json
	args := DeconstructBuildArgs(req.GetArgs())

	return AnalyzeSubTmpl(templateStr, args)
}

// DeconstructBuildArgs : 结构proto参数转换为json
// 参数:
//   - args: 模板所需的proto参数
//
// 返回值:
//   - 返回 json参数
func DeconstructBuildArgs(args map[string]*pb.TemplateArg) map[string]interface{} {
	data := make(map[string]interface{})
	for k, v := range args {
		data[k] = extraTemplateArgValue(v)
	}
	return data
}

func extraTemplateArgValue(v *pb.TemplateArg) interface{} {
	switch elem := v.GetData().(type) {
	case *pb.TemplateArg_StrVal:
		return elem.StrVal
	case *pb.TemplateArg_IntVal:
		return elem.IntVal
	case *pb.TemplateArg_DoubleVal:
		return elem.DoubleVal
	case *pb.TemplateArg_BoolVal:
		return elem.BoolVal
	case *pb.TemplateArg_ValItems:
		items := elem.ValItems.GetArgs()
		itemsArr := make([]interface{}, len(items))
		for i, item := range items {
			itemsArr[i] = ExtraArgItemValue(item)
		}
		return itemsArr
	}
	return ""
}

// RenderMasterTemplate : 渲染主模板
// 参数:
//   - templateStr: sqlbuilder 请求体
//
// 返回值:
//   - 返回 sql
//   - 返回 错误信息
func RenderMasterTemplate(
	tmplWrapper *pb.SqlText) (string, error) {
	// 主模板sql
	masterTemplateStr := tmplWrapper.GetText()
	masterTmplArgs := tmplWrapper.GetArgs()

	args := DeconstructBuildArgs(masterTmplArgs)

	// 创建模板画布, 注册所有自定义函数, 解析主模板
	tmpl, err := template.New("root").Funcs(InjectFunc()).Parse(masterTemplateStr)
	if err != nil {
		return "", err
	}

	var masterBuf bytes.Buffer
	// 渲染主模板
	err = tmpl.ExecuteTemplate(&masterBuf, "master", args)
	if err != nil {
		return "", errors.New(fmt.Sprintf("最终sql生成失败, msg: %s", err.Error()))
	}

	return masterBuf.String(), nil
}

func AnalyzeAdditionByProto(req *pb.AnalyzeAdditionRequest) (*NativeSqlHeader, error) {
	templateStr := req.Tmpl
	if templateStr == "" {
		return nil, errors.New("addition 模板不能为空")
	}

	return ExtractAdditionFromTemplate(templateStr)
}
