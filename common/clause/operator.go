package clause

import "strings"

type CompareExpression string // > = < >= <= !=
type ArithExpression string   // + - * / %
type LogicExpression string   // and, or, not, between, in, not in, like, is not null,

const (
	Eq  CompareExpression = "="
	Neq CompareExpression = "!="
	Gt  CompareExpression = ">"
	Gte CompareExpression = ">="
	Lt  CompareExpression = "<"
	Lte CompareExpression = "<="
)

const (
	Add  ArithExpression = "+"
	Sub  ArithExpression = "-"
	Mul  ArithExpression = "*"
	Div  ArithExpression = "/"
	Mod  ArithExpression = "%"
	Mod2 ArithExpression = "%%"
)

const (
	And       LogicExpression = " and "
	Or        LogicExpression = " or "
	Not       LogicExpression = " not "
	Between   LogicExpression = " between "
	In        LogicExpression = " in "
	NotIn     LogicExpression = " not in "
	Like      LogicExpression = " like "
	IsNotNull LogicExpression = " is not null "
)

const (
	BoolEq  = "="
	BoolNeq = "<>"
	BoolGt  = ">"
	BoolGte = ">="
	BoolLt  = "<"
	BoolLte = "<="
	// 在select字段使用代表bool函数
	PgBoolFuncIsNull    = "is null"
	PgBoolFuncIsNotNull = "is not null"
	PgBoolFuncBetween   = "between"
)

const (
	COMMA = ", "
)

const (
	PntFormat = `(%s)`

	PGAsFormat             = `%s as "%%s"`
	PGLiteralTableAsFormat = `(%s) as "%s"`
	PGAsAppend             = ` as "%%s"`

	MYSQLAsFormat             = "%s as `%%s`"
	MYSQLLiteralTableAsFormat = "(%s) as `%s`"
	MYSQLAsAppend             = " as `%%s`"

	PgBoolFuncEqFormat        = "(%s = %s)"
	PgBoolFuncNeqFormat       = "(%s != %s)"
	PgBoolFuncGtFormat        = "(%s > %s)"
	PgBoolFuncGteFormat       = "(%s >= %s)"
	PgBoolFuncLtFormat        = "(%s < %s)"
	PgBoolFuncLteFormat       = "(%s <= %s)"
	PgBoolFuncIsNullFormat    = "(%s is null)"
	PgBoolFuncIsNotNullFormat = "(%s is not null)"
	PgBoolFuncBetweenFormat   = "(%s between %s and %s)"
)

type Operator struct {
	Op any // 基于上面3种类型
}

var (
	// ArithFormatMap 内置函数名称和SQL片段 映射
	ArithFormatMap = map[string]string{
		string(Add): string(Add),
		string(Sub): string(Sub),
		string(Mul): string(Mul),
		string(Div): string(Div),
		string(Mod): string(Mod2),
	}

	BoolFormatMap = map[string]string{
		BoolEq:              PgBoolFuncEqFormat,
		BoolNeq:             PgBoolFuncNeqFormat,
		BoolGt:              PgBoolFuncGtFormat,
		BoolGte:             PgBoolFuncGteFormat,
		BoolLt:              PgBoolFuncLtFormat,
		BoolLte:             PgBoolFuncLteFormat,
		PgBoolFuncIsNull:    PgBoolFuncIsNullFormat,
		PgBoolFuncIsNotNull: PgBoolFuncIsNotNullFormat,
		PgBoolFuncBetween:   PgBoolFuncBetweenFormat,
	}
)

// CalArithFormat 计算算术运算符的format
// 参数:
//   - call: 运算符号
//   - argsLen: 参数个数
//
// 返回值:
//   - 返回 算术运算符的format
func CalArithFormat(call string, argLen int) string {
	// strings.Join(formatParts, " "+ call + " ")
	var formatParts []string
	for i := 0; i < argLen; i++ {
		formatParts = append(formatParts, "%s")
	}

	return strings.Join(formatParts, " "+ArithFormatMap[call]+" ")
}

func CalArithFormatWithBuilder(call string, argLen int) string {
	var builder strings.Builder

	// 用一个循环逐个构建字符串部分
	for i := 0; i < argLen; i++ {
		if i > 0 {
			// 如果不是第一个元素，则先添加运算符
			builder.WriteString(" " + ArithFormatMap[call] + " ")
		}
		// 为每个参数添加"%s"
		builder.WriteString("%s")
	}

	return builder.String()
}
