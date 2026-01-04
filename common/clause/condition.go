package clause

type ConditionOperator string

const (
	ConditionEq   ConditionOperator = "="
	ConditionNeq  ConditionOperator = "<>"
	ConditionLt   ConditionOperator = "<"
	ConditionLte  ConditionOperator = "<="
	ConditionGt   ConditionOperator = ">"
	ConditionGte  ConditionOperator = ">="
	ConditionLike ConditionOperator = "like"

	ConditionBetween ConditionOperator = "between"
	ConditionIn      ConditionOperator = "in"
	ConditionNotIn   ConditionOperator = "not in"

	ConditionIsNull    ConditionOperator = "is null"
	ConditionIsNotNull ConditionOperator = "is not null"

	ConditionAnd ConditionOperator = "and"
	ConditionOr  ConditionOperator = "or"
)

// Condition
// 遍历 Where-Condition, 元素 item
// switch item.Operator
// case > = < >= <= <> like: for field, value := range item.Args, xormCond builder.Cond = builder.$Operator{field: value}
// case between: for field, value := range item.Args, value.(betweenObj), Between{Col: $field, LessVal: $left, MoreVal: $right}
// case in: for field, value := range item.Args, In($field, []any{$value})
// case is null, is not null: for field, value := range item.Args, value.(bool), IsNull{$field} / NotNull{$field}
// case and , or: for field, value := range item.Args, Expr($exprFmt, ...value)
type Condition struct {
	Field any `json:"field"`
	// 条件自身参数. e.g. xorm.Eq{"column1": 4}, args 默认1组, 多组之间逻辑关系是 and;
	// value 值类型: string & int(compare), slice(in),betweenObj(between), bool(is null/is not null), exprObj(and/or, k:fmtStr, v:args)
	Args []any `json:"args"`
	// 条件自身操作符: > = < >= <= <> like, between, in, not in, is null, expr
	Operator string `json:"operator"`
	// 和下一个条件的逻辑关系, 默认 and, 可选or
	Logic string `json:"logic"`
}
