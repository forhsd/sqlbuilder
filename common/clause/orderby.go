package clause

type OrderByClause struct {
	OrderBy []OrderBy `json:"orderBy"`
}

type OrderBy struct {
	// 类型断言: Column 或者 Expression
	Dependent any `json:"dependent"`
	// 枚举排序方式; e.g. asc, desc
	Order string `json:"order"`
}
