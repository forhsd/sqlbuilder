package clause

type WhereClause struct {
	//Op    Operator
	//Field []FieldClause // 字段个数基于表达式, 通常表达式左边右边各一个字段, 但一些表达式只要一个字段

	Expr []Expression
}
