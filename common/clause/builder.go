package clause

type BuilderStrategy int
type Driver int

const (
	// BackQuote 反引号反义处理
	BackQuote = "`"
	// DoubleQuotation 双引号反义处理
	DoubleQuotation = `"`
)

const (
	// ModelStrategy 模型模式
	ModelStrategy BuilderStrategy = iota
	// TemplateStrategy 模板模式
	TemplateStrategy
)

const (
	DriverPostgres Driver = iota
	DriverMysql
	DriverDoris
)

var (
	// DriverKeywordWrapMap 驱动关键字包裹符
	DriverKeywordWrapMap = map[Driver]string{
		DriverMysql:    BackQuote,
		DriverDoris:    BackQuote,
		DriverPostgres: DoubleQuotation,
	}
)

// ==============================================================

type DeepWrapper struct {
	// 深度: 控制层级; e.g. 0, 代表最里层, 1~n, 代表依次包裹 0 层
	Deep int `json:"deep"`
	// 某个层级的SQL引用对象
	Sql SQLReference `json:"sql"`
}

type BuilderRequest struct {
	// 不同层级嵌套构建SQL
	SQLBuilders []DeepWrapper   `json:"builders"`
	Driver      Driver          `json:"driver"`
	Strategy    BuilderStrategy `json:"strategy"`
}

// SQLReference todo Select缺少自定义字段
type SQLReference struct {
	From Table  `json:"from"`
	Join []Join `json:"join"`
	// 类型断言: Condition 或者 Expression (返回值是bool的函数)
	Where []interface{} `json:"where"`
	// 类型断言: Column 或者 Expression
	GroupBy []interface{} `json:"groupBy"`
	// Aggregation 会加入 Select 片段
	Aggregation []Expression `json:"aggregation"`
	// 类型断言: Column 或者 Expression
	Select  []interface{} `json:"select"`
	OrderBy []OrderBy     `json:"orderBy"`
	Limit   LimitClause   `json:"limit"`
}

func (i BuilderStrategy) EnType() string {
	names := [...]string{"model", "template"}
	if i < 0 || int(i) >= len(names) {
		// 处理未知类型的情况
		return "unknown"
	}
	return names[i]
}

func BuilderStrategyEnumType(index BuilderStrategy) string {
	switch index {
	case ModelStrategy, TemplateStrategy:
		return index.EnType()
	default:
		return ModelStrategy.EnType()
	}
}

func (i Driver) EnType() string {
	names := [...]string{"postgres", "mysql", "doris"}
	if i < 0 || int(i) >= len(names) {
		// 处理未知类型的情况
		return "unknown"
	}
	return names[i]
}

func DriverEnumType(index Driver) string {
	switch index {
	case DriverPostgres, DriverMysql, DriverDoris:
		return index.EnType()
	default:
		return "unknown"
	}
}
