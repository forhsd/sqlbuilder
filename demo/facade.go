package demo

import (
	"database/sql"
	"github.com/forhsd/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func Gen(db *gorm.DB, limitN int) string {

	var startTime = time.Date(2024, 7, 22, 10, 0, 0, 0, time.UTC)
	var endTime = time.Date(2024, 7, 23, 10, 0, 0, 0, time.UTC)

	stmt := gorm.Statement{DB: db, Clauses: map[string]clause.Clause{}}
	// []clause.Interface 顺序：select --> from --> where --> group by --> having --> order by --> limit
	clauses := []clause.Interface{

		clause.Select{Columns: []clause.Column{
			{Table: "ht_order", Name: "id", Alias: "tid"},
			{Name: "SUM(amount)", Alias: "total_amount", Raw: true},
		}},
		clause.From{Tables: []clause.Table{
			{Name: `"public"."ht_order"`, Alias: "ho"},
		}},
		clause.Where{Exprs: []clause.Expression{
			clause.Gt{Column: clause.Column{Table: "ho", Name: "gmt_time"}, Value: &startTime},
			clause.Lt{Column: clause.Column{Table: "ho", Name: "gmt_time"}, Value: &endTime},
		}},

		clause.GroupBy{Columns: []clause.Column{
			{Table: "ho", Name: "id"},
		}},
		clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Table: "ho", Name: "total_amount"}, Desc: true},
		}},
		clause.Limit{Limit: &limitN},
	}

	for _, cs := range clauses {
		stmt.AddClause(cs)
	}
	// 必须按照 []clause.Interface，否则影响sql字符串生成顺序，即：按照sql语句关键字的顺序申明
	stmt.Build("SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "LIMIT")
	// stmt.Vars... : $N, 占位符注入sql, 占位符切片顺序和 []clause.Interface 顺序写入切片时一致
	sql := stmt.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Raw(stmt.SQL.String(), stmt.Vars...)
	})
	logger.Info("sql: \r\n %s", sql)

	return sql
}

/**
select
	p."category",
	sum(p."price") as "category_total"
from "sample_data"."products" p
where p."category" = 'Doohickey'
group by p."category";
*/

func GenProductSql(db *gorm.DB, category string) string {
	stmt := gorm.Statement{DB: db, Clauses: map[string]clause.Clause{}}
	// []clause.Interface 顺序：select --> from --> where --> group by --> having --> order by --> limit

	var clauses []clause.Interface

	if "" == category {
		clauses = []clause.Interface{
			clause.Select{Columns: []clause.Column{
				{Table: "p", Name: "category"},
				{Name: "SUM(price)", Alias: "category_total", Raw: true},
			}},
			clause.From{Tables: []clause.Table{
				{Name: `"sample_data"."products"`, Alias: "p"},
			}},
			clause.GroupBy{Columns: []clause.Column{
				{Table: "p", Name: "category"},
			}},
		}
	} else {
		clauses = []clause.Interface{
			clause.Select{Columns: []clause.Column{
				{Table: "p", Name: "category"},
				{Name: "SUM(price)", Alias: "category_total", Raw: true},
			}},
			clause.From{Tables: []clause.Table{
				{Name: `"sample_data"."products"`, Alias: "p"},
			}},
			clause.Where{Exprs: []clause.Expression{
				clause.Eq{Column: clause.Column{Table: "p", Name: "category"}, Value: &category},
			}},
			clause.GroupBy{Columns: []clause.Column{
				{Table: "p", Name: "category"},
			}},
		}
	}

	for _, cs := range clauses {
		stmt.AddClause(cs)
	}
	// 必须按照 []clause.Interface，否则影响sql字符串生成顺序，即：按照sql语句关键字的顺序申明
	stmt.Build("SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "LIMIT")
	// stmt.Vars... : $N, 占位符注入sql, 占位符切片顺序和 []clause.Interface 顺序写入切片时一致
	nativeSql := stmt.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Raw(stmt.SQL.String(), stmt.Vars...)
	})
	logger.Info("[GenProductSql] sql:\n %s ", nativeSql)

	return nativeSql
}

func GenProduct(db *gorm.DB, category string) []map[string]interface{} {
	logger.Debug("[GenProduct] category: %s", category)

	nativeSql := GenProductSql(db, category)

	rows, err := db.Raw(nativeSql).Rows()
	if err != nil {
		logger.Error("[GenProduct] 查询失败: %s", err)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Error("[GenProduct] 关闭session失败: %s", err)
		}
	}(rows)

	var result []map[string]interface{}

	for rows.Next() {
		var rowData map[string]interface{}
		err := db.ScanRows(rows, &rowData)
		if err != nil {
			logger.Error("[GenProduct] 写入result对象失败: %s", err)
		}
		result = append(result, rowData)
	}

	return result
}
