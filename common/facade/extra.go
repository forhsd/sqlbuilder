package facade

import (
	"bytes"
	"encoding/json"
	"errors"
	"sort"
	"text/template"
	"text/template/parse"

	pb "github.com/forhsd/sqlbuilder/gen/proto"
	"github.com/forhsd/sqlbuilder/util"
	xorm "xorm.io/builder"
)

// -------------------------------------------  提取变量 ----------------------------------------------

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
	newVars := util.RemoveElements(vars, tmplNames)
	// 变量名称顺序
	sort.Strings(newVars)

	return newVars, nil
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

// -------------------------------------------  提取物理表 ----------------------------------------------

func ExtractAdditionFromTemplate(
	additionTmplText string) (*NativeSqlHeader, error) {
	// 创建模板画布, 注册所有自定义函数, 解析模板字符串
	tmpl, err := template.New("addition").Funcs(InjectFunc()).Parse(additionTmplText)
	if err != nil {
		return nil, err
	}

	// 创建一个 bytes.Buffer 来作为 io.Writer
	var buf bytes.Buffer

	// 执行模板并将结果输出到 bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		return nil, err
	}

	// 假设渲染出的结果是 JSON 数组，我们可以反序列化它并填充结构体
	var specifications []Specifications
	err = json.Unmarshal(buf.Bytes(), &specifications)
	if err != nil {
		return nil, err
	}

	return &NativeSqlHeader{
		Specs: specifications,
	}, nil
}

// 构建方言
func Dialect(req *pb.BuilderRequest) *xorm.Builder {
	switch req.Driver {
	case pb.Driver_DRIVER_DORIS:
		return xorm.Dialect(xorm.MYSQL)
	case pb.Driver_DRIVER_MYSQL:
		return xorm.Dialect(xorm.MYSQL)
	case pb.Driver_DRIVER_POSTGRES:
		return xorm.Dialect(xorm.POSTGRES)
	default:
		panic(errors.New("未适配的数据库"))
	}
	// return nil
}

// 提取SELECT别名
func ExtractSelectAlias(sel []*pb.MixField) (alias []string) {

	for _, s := range sel {
		var as string
		switch s.GetMix().(type) {
		case *pb.MixField_Column:
			as = s.GetColumn().GetAlias()
		case *pb.MixField_Expression:
			as = s.GetExpression().GetCallAs()
		case *pb.MixField_CaseWhen:
			as = s.GetCaseWhen().GetAlias()
		}
		alias = append(alias, as)
	}
	return
}

// 提取ORDER别名
func ExtractOrderAlias(order []*pb.OrderBy) (alias []string) {
	for _, o := range order {
		var as string
		dep := o.GetDependent()
		switch dep.GetMix().(type) {
		case *pb.MixField_Column:
			as = dep.GetColumn().GetAlias()
		case *pb.MixField_Expression:
			as = dep.GetExpression().GetCallAs()
		case *pb.MixField_CaseWhen:
			as = dep.GetCaseWhen().GetAlias()
		}
		as = as + " " + o.Order
		alias = append(alias, as)
	}
	return
}
