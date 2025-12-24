package demo

import (
	"fmt"
	"testing"

	"github.com/forhsd/logger"
	"github.com/forhsd/sqlbuilder/connect"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// assertEqual 是一个辅助函数，用于比较实际值和预期值，如果不相等则输出错误信息
func assertEqual(t *testing.T, actual, expected interface{}, message string) {
	if actual != expected {
		t.Errorf("Assertion failed: %s - Actual: %v, Expected: %v", message, actual, expected)
	}
}

func TestFacadeGen(t *testing.T) {
	db := connect.PostgresConn()
	_ = Gen(db, 10)
}

func TestFacadeGenProduct(t *testing.T) {
	db := connect.PostgresConn()
	_ = GenProductSql(db, "")
}

func TestFacadeGenProductJoin(t *testing.T) {
	db := connect.PostgresConn()
	_ = GenProductJoinSql(db, "")
}

func TestFacadeGenToSqlResult(t *testing.T) {
	db := connect.PostgresConn()
	_ = GenToSqlResult(db)

}

func TestFacadeGenDryRun(t *testing.T) {
	db := connect.PostgresConn()
	_ = GenDryRun(db)
}

func GenDryRun(db *gorm.DB) string {
	stmt := db.Session(&gorm.Session{DryRun: true}).
		Table(`"sample_data"."products" p`).Joins(`LEFT JOIN "sample_data"."order" o on o."product_id" = p."id"`).
		Select(`p."category"`, `sum(p."price") as category_total`).
		Group(`p."category"`).
		// 不建议使用转义符号, 很难拼装字符串,使用模板字符串处理
		Where("p.\"category\" = ?", "Doohickey").
		Statement

	stmt.Build("SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "LIMIT")

	drySql := stmt.SQL.String()

	fmt.Printf("drySql: %s \n", drySql)

	return drySql
}

func GenToSqlResult(db *gorm.DB) []map[string]interface{} {
	rows, err := db.Table(`"sample_data"."products" p`).Joins(`LEFT JOIN "sample_data"."order" o on o."product_id" = p."id"`).
		Select(`p."category"`, `sum(p."price") as category_total`).
		Group(`p."category"`).
		Where("p.\"category\" = ?", "Doohickey").
		Rows()

	if err != nil {
		logger.Error("[GenToSql] 查询失败: %s", err)
	}

	var result []map[string]interface{}

	//sql2 := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
	//	return db.Table("products").
	//		Select("category", "sum(price)").
	//		//Where("category = ?", "Doohickey").
	//		Group("category").Find(&result)
	//})
	//fmt.Printf("sql2: %s \n", sql2)

	for rows.Next() {
		var rowData map[string]interface{}
		err := db.ScanRows(rows, &rowData)
		if err != nil {
			logger.Error("[GenToSql] 写入result对象失败: %s", err)
		}
		result = append(result, rowData)
	}
	logger.Info("result: %+v", result)
	return result
}

func GenProductJoinSql(db *gorm.DB, category string) string {
	stmt := gorm.Statement{DB: db, Clauses: map[string]clause.Clause{}}
	// []clause.Interface 顺序：select --> from --> where --> group by --> having --> order by --> limit

	var clauses []clause.Interface

	// Raw: true 删除最近字符串中的双引号，``模板字符串内的原样保留
	if "" == category {
		clauses = []clause.Interface{
			clause.Select{Columns: []clause.Column{
				{Table: "p", Name: `"category"`, Raw: true},
				{Name: `SUM(p."price")`, Alias: "category_total", Raw: true},
			}},
			clause.From{
				Tables: []clause.Table{{Name: `"sample_data"."products"`, Alias: `p`, Raw: true}},
				Joins: []clause.Join{{
					Type:  clause.LeftJoin,
					Table: clause.Table{Name: `"sample_data"."order"`, Alias: "o", Raw: true},
					ON: clause.Where{Exprs: []clause.Expression{
						clause.Eq{Column: clause.Column{Table: "o", Name: `"product_id"`, Raw: true}, Value: clause.Column{Table: "p", Name: `"id"`, Raw: true}},
					}},
				}},
			},
			clause.GroupBy{Columns: []clause.Column{
				{Table: "p", Name: `"category"`, Raw: true},
			}},
		}
	} else {
		clauses = []clause.Interface{
			clause.Select{Columns: []clause.Column{
				{Table: "p", Name: `"category"`, Raw: true},
				{Name: "SUM(price)", Alias: "category_total", Raw: true},
			}},
			clause.From{
				Tables: []clause.Table{{Name: `"sample_data"."products"`, Alias: `p`, Raw: true}},
				Joins: []clause.Join{{
					Type:  clause.LeftJoin,
					Table: clause.Table{Name: `"sample_data"."order"`, Alias: "o", Raw: true},
					ON: clause.Where{Exprs: []clause.Expression{
						clause.Eq{Column: clause.Column{Table: "o", Name: `"product_id"`, Raw: true}, Value: clause.Column{Table: "p", Name: `"id"`, Raw: true}},
					}},
				}},
			},
			clause.Where{Exprs: []clause.Expression{
				clause.Eq{Column: clause.Column{Table: "p", Name: "category"}, Value: &category},
			}},
			clause.GroupBy{Columns: []clause.Column{
				{Table: "p", Name: "category"},
			}},
		}
	}

	// clauses 载入 stmt 内存
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
