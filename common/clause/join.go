package clause

type JoinType string //	left join, right join, inner join, full join, cross join

const (
	LeftJoin  JoinType = "left join"
	RightJoin JoinType = "right join"
	InnerJoin JoinType = "inner join"
	FullJoin  JoinType = "full join"
	CrossJoin JoinType = "cross join"
)

const (
	LeftJoinXorm  string = "LEFT"
	RightJoinXorm string = "RIGHT"
	InnerJoinXorm string = "INNER"
	FullJoinXorm  string = "FULL"
	CrossJoinXorm string = "CROSS"

	CrossJoinPlaceholder      = "/*__XORM_CROSS_JOIN__*/"
	CrossJoinPlaceholderRegex = `(?i)\s*ON\s*/\*__XORM_CROSS_JOIN__\*/`
)

// OnFieldFormat e.g. InnerJoin("table3", "table2.id = table3.tid")
const OnFieldFormat string = `%s %s %s`
const SafeCond string = `( %s )`
const Space string = ` `

type Join struct {
	// 连接方式
	Type JoinType `json:"type"`
	// 连接表
	Table Table `json:"table"`
	// 类型断言: FieldClause, 默认主表字段
	Left any `json:"left"`
	// 类型断言: FieldClause, 默认从表字段
	Right any `json:"right"`
	// 连接条件
	On LogicExpression `json:"on"`
}

type JoinClause struct {
	Join []Join
}
