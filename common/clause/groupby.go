package clause

type GroupByClause struct {
	// 分组字段
	GroupBys []FieldClause `json:"groupBy"`
}
