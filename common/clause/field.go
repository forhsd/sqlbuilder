package clause

const (
	StringLiteral       = `%s`
	StringSafeLiteral   = `'%s'`
	DoubleStringLiteral = `%s %s`
	AnyLiteral          = `%v`

	PGStringLiteralSafe   = `"%s"`
	PGSchemaTable         = `"%s"."%s"`
	PGTableColumn         = `"%s"."%s"`
	PGTableColumnAs       = `"%s"."%s" as "%s"`
	PGSchemaTableColumn   = `"%s"."%s"."%s"`
	PGSchemaTableColumnAs = `"%s"."%s"."%s" as "%s"`

	MysqlStringLiteralSafe   = "`%s`"
	MysqlTableTableColumn    = "`%s`.`%s`"
	MysqlTableColumnAs       = "`%s`.`%s` as `%s`"
	MysqlSchemaTableColumn   = "`%s`.`%s`.`%s`"
	MysqlSchemaTableColumnAs = "`%s`.`%s`.`%s` as `%s`"
)

// FieldClause 使用需要断言
type FieldClause struct {
	// 包含: Column 或 Expression
	Column interface{} `json:"column"`
}

type Column struct {
	// 字段程序名称, 用于直接拼接sql的字符串
	Field string `json:"field"`
	// 字段所属物理表或者表别名, 用于直接拼接sql的字符串
	Table string `json:"table"`
	// 字段所属模式/库, 用于直接拼接sql的字符串
	Schema string `json:"schema"`
	// 字段别名, 用于直接拼接sql的字符串
	Alias string `json:"alias"`
	// 是否能够聚合
	AggAble bool `json:"aggAble"`
	// 是否使用别名
	UseAs bool `json:"useAs"`

	//// 以下内容可选
	//// 数组索引?, 遍历顺序，避免乱序或重复填充
	//Index int `json:"index"`
	//// 主键，唯一键，自生成策略;e.g. pk/auto, pk/uuid, self/uuid, self/snowfake
	//UniType string `json:"uniType"`
	//// 唯一值
	//UniValue interface{} `json:"uniValue"`
	//// 字段显示名称; 字段别名取field, 展示名取用户输入
	//DisplayName string `json:"displayName"`
	//// 字段类型, 用于校对json值; e.g. 字符串转时间
	//ColumnType string `json:"columnType"`
	//// 是否需要校对?
	//UseValidate bool `json:"useValidate"`
	//// 字段值, 主要针对字符串校正
	//RawColumnValue interface{} `json:"rawColumnValue"`
	//// 字段归属方?, (sourceFrom,belong to)映射唯一标识值, 部分场景需要; e.g. 不属于任何 null, 自己就是根部 'root'
	//UniDependent interface{} `json:"uniDependent"`
	//// 字段归属方程序值?, 物理表别名或者子查询别名, 用于拼接sql的字符串
	//RawDependent string `json:"rawDependent"`
}

// todo 表达式类型字段暂不实现
//type JoinField struct {
//	field    Column
//	isMaster bool // 是否为主表, 主表 column，副表别名 table__column
//}
