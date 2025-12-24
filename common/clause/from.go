package clause

type Table struct {
	// 必填: 程序名称, 用于拼接sql的字符串
	TableName   string `json:"tableName"`
	TableSchema string `json:"tableSchema"`
	// 必填: 程序别名, 用于拼接sql的字符串, 调用方赋值取 tableName
	TableAlias string `json:"tableAlias"`
}

type FromClause struct {
	// 主表, 通常1个, 不支持 from a, b 形式的写法
	Table Table `json:"table"`
}
