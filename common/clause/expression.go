package clause

type ExpressionCallType string

const (
	CallTypeAgg    ExpressionCallType = "agg"
	CallTypeInner  ExpressionCallType = "inner"
	CallTypeCustom ExpressionCallType = "custom"
)

// Expression 函数表达式
// e.g.
// `DATE_TRUNC('month', "t1"."created_at") as ""created_at`
// `sum("p"."price") as "合计价格"`
type Expression struct {
	// interface{}? 表达式 = 其他表达式 + 通用聚合 e.g. count, sum, avg, max, min, distinct
	Call string `json:"call"`
	// 函数表达式类型: 内置函数, 自定义函数, 聚合函数
	CallType ExpressionCallType `json:"callType"`
	// 类型: Expression, string, int, Column . 表达式入参: 元组, 按照表达式参数顺序或者关键字参数, 包括作用的物理表列, e.g. 内置函数规定函数签名
	Vars []interface{} `json:"vars"`
	// 函数别名, select 有值, where 无值
	CallAs string `json:"callAs"`
	// 是否使用别名
	UseAs bool `json:"useAs"`
}

// ExpressionOption 类型定义
type ExpressionOption func(*Expression)

// WithExpressionCallAs 设置 CallAs 字段
func WithExpressionCallAs(callAs string) ExpressionOption {
	return func(c *Expression) {
		c.CallAs = callAs
	}
}

//func (b *Expression) mix() {
//}

// ExpressionCallType 映射函数 AggregationFuncFormat 和 InnerFuncMap
