package facade

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	xorm "xorm.io/builder"
)

func TestXORMSQLBuilderUnion(t *testing.T) {
	builder := xorm.Select("*").From("a").Where(xorm.Eq{"status": "1"}).
		Union("all", xorm.Select("*").From("a").Where(xorm.Eq{"status": "2"})).
		Union("distinct", xorm.Select("*").From("a").Where(xorm.Eq{"status": "3"})).
		Union("", xorm.Select("*").From("a").Where(xorm.Eq{"status": "4"}))

	sql, args, err := builder.ToSQL()
	/**
	(SELECT * FROM a WHERE status=?)
	UNION ALL
	(SELECT * FROM a WHERE status=?)
	UNION DISTINCT
	(SELECT * FROM a WHERE status=?)
	UNION
	(SELECT * FROM a WHERE status=?)

	[1 2 3 4] <nil>
	*/
	t.Log(sql, args, err)

	// (SELECT * FROM a WHERE status='1') UNION ALL (SELECT * FROM a WHERE status='2') UNION DISTINCT (SELECT * FROM a WHERE status='3') UNION (SELECT * FROM a WHERE status='4') <nil>
	boundSQL, err := builder.ToBoundSQL()
	t.Log(boundSQL, err)

}

func TestXORMSQLBuilderSub(t *testing.T) {
	builder := xorm.Select("sub.id").From(xorm.Select("c").From("table1").Where(xorm.Eq{"a": 1}), "sub").Where(xorm.Eq{"b": 1})

	sql, args, err := builder.ToSQL()
	// SELECT sub.id FROM (SELECT c FROM table1 WHERE a=?) sub WHERE b=? [1 1] <nil>
	t.Log(sql, args, err)

	// SELECT sub.id FROM (SELECT c FROM table1 WHERE a=1) sub WHERE b=1 <nil>
	boundSQL, err := builder.ToBoundSQL()
	t.Log(boundSQL, err)

}

func TestXORMSQLBuilder(t *testing.T) {
	//builder := xorm.Dialect(xorm.POSTGRES).
	//	Select(`DATE_TRUNC('month', "t1"."created_at") as created_at`, `"t1"."category"`, `sum("t1"."price")`).
	//	From(`"sample_data"."products"`, "t1").
	//	GroupBy(`DATE_TRUNC('month', "t1"."created_at"), "t1"."category"`).
	//	OrderBy(`DATE_TRUNC('month', "t1"."created_at") asc, "t1"."category" desc`)
	//
	//boundSQL, _ := builder.ToBoundSQL()
	//t.Log(boundSQL)

	// SELECT c FROM table1 ORDER BY CASE WHEN owner_name LIKE 'a' THEN 0 ELSE 1 END
	//sqlB, _ := xorm.Select("c").From("table1").OrderBy(xorm.Expr("CASE WHEN owner_name LIKE ? THEN 0 ELSE 1 END", "a")).ToBoundSQL()
	//fmt.Println(sqlB)

	builderWhere := xorm.Dialect(xorm.POSTGRES).
		Select(`"t1"."category"`).
		From(`"sample_data"."products"`, "t1").
		//InnerJoin(`(table2) as "t2"`, xorm.Expr("1=1")). // 1=1 处理 JOIN-ON 强制要求 ON 的问题
		InnerJoin(xorm.As("table2", "t2"), xorm.Expr("1=1")). // 原生As有坑，不会根据方言带引号 => 还是只能代码自己处理别名引号
		//Where(xorm.Or(xorm.Eq{"c": 3, "a": "asda"}, xorm.Eq{"d": 4}))  // WHERE (a='asda' AND c=3) OR d=4 => 如果And和Or平级无法控制逻辑, 建议每块之间是And, 组内正常处理
		//Where(xorm.Expr("c >1 and d = 6 or (q between 1 and 3)")) // expr 允许一切原生写法
		//Where(xorm.Expr("c >1 and d = ? or (q between 1 and 3)", "asd")) // expr 还有 fmt.Sprintf() 效果
		//Where(xorm.In("col1", []int{1, 2, 10})) // kv中v是数组
		//Where(xorm.Between{Col: "col1", LessVal: 1, MoreVal: 10}) // betweenObj
		//Where(xorm.Eq{"id": 1}).And(xorm.Eq{"aaa": 1}).Or(xorm.Lte{"qqq": 100}).And(xorm.Eq{"ddd": "qwe"}) // and/or 会将当前视作右和前面所有视为左整体圆括号圈定
		//Where(xorm.Eq{"id": 1}).And(xorm.Eq{"aaa": 1}).And(xorm.Lte{"qqq": 100}).And(xorm.Eq{"ddd": "qwe"}) //所有都是and不会加圆括号
		//Where(xorm.Eq{"id": 1, "category": "Doohickey", "price": 10}) // 按照key排序, 且共用一个操作符
		//Where(xorm.Eq{`"product"."id"`: 1}).Where(xorm.Eq{`"product"."category"`: "Doohickey"}) // 默认将多个where按照and粘合
		//And(xorm.Eq{`"product"."id"`: 1}).Or(xorm.Eq{`"product"."category"`: "Doohickey"}) // 可以不申明where, 会自动在前面拼接where关键字
		//Where(xorm.And(xorm.Eq{`"product"."id"`: 1}).Or(xorm.Eq{`"product"."category"`: "Doohickey"})) // 在第一个不申明条件前拼接where关键字不影响
		//Where(xorm.Not{xorm.Between{Col: "created_at", LessVal: "2017-07-19", MoreVal: "2017-09-14"}}) // 取反
		Where(xorm.Like{"created_at", "www"}) // like 默认全模糊

	whereBoundSQL, _ := builderWhere.ToBoundSQL()
	t.Log(whereBoundSQL)

	// 测试 join-on
	//sqlF, _ := xorm.ToBoundSQL(xorm.Lt{"d": 3}.And(xorm.Lt{"e": 4}))
	//fmt.Println(sqlF)
	//var sss = []string{" e<4", "d<3 ", `"t1"."id" = "t2"."t1_id"`}
	//fs := strings.Join(sss, " and ")
	//fmt.Println(fs)

	//// 使用 expr 实现下面效果
	//// (a=1 AND b LIKE '%c%') OR (a=2 AND b LIKE '%g%')
	//segment := xorm.Eq{"a": 1}.And(xorm.Like{"b", "c"}).Or(xorm.Eq{"a": 2}.And(xorm.Like{"b", "g"})).
	//	And(xorm.Expr("c >? or (q between 1 and 3)", 10)) // like 默认全模糊
	//segmentBoundSQL, _ := xorm.ToBoundSQL(segment)
	//t.Log(segmentBoundSQL)
}

// result := fmt.Sprintf(template, name, accountNumber)

func TestExprCond(t *testing.T) {
	// 两个where之间默认填充 AND, 如果希望一个字符串包含多个表达式使用 Expr, 内置表达式使用对应 type

	b := xorm.Select("id").
		From("table1").
		Where(xorm.Expr("a=? OR b=?", 1, 2)).
		Where(xorm.Or(xorm.Eq{"c": 3}, xorm.Eq{"d": 4}))
	sql, args, err := b.ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "table1", b.TableName())
	assert.EqualValues(t, "SELECT id FROM table1 WHERE (a=? OR b=?) AND (c=? OR d=?)", sql)
	assert.EqualValues(t, []any{1, 2, 3, 4}, args)

	// 下面示例只能结合 Expression 和 Condition 表达:
	// Expression.Vars{Condition无next, Condition有1next}
	logicNest := xorm.Or(xorm.Eq{"a": 3}, xorm.And(xorm.Eq{"b": 1}, xorm.Eq{"c": 2}))
	logicSql, err := xorm.ToBoundSQL(logicNest)
	assert.EqualValues(t, "a=3 OR (b=1 AND c=2)", logicSql)

}

func TestInsertExpr(t *testing.T) {
	sql, err := xorm.Insert(xorm.Eq{`INSERT_INTO_NIL`: xorm.Expr("SELECT b FROM t WHERE d=? LIMIT 1", 2)}).
		Into("table1").ToBoundSQL()
	assert.NoError(t, err)
	// INSERT INTO table1 (INSERT_INTO_NIL) Values ((SELECT b FROM t WHERE d=2 LIMIT 1))
	fmt.Println(sql)

	// 期望: INSERT INTO table1  Values ((SELECT b FROM t WHERE d=2 LIMIT 1))
	output := strings.ReplaceAll(sql, "(INSERT_INTO_NIL)", "")
	fmt.Println(output)
	assert.EqualValues(t, "INSERT INTO table1  Values ((SELECT b FROM t WHERE d=2 LIMIT 1))", output)
}

func TestFromLiteralText(t *testing.T) {

	// from 和 join 子查询带字面量可行性验证

	mainSql := `
		(
			SELECT
				-- 主查询: area、province、sum(order_number)
				org_t_alias."area" AS org_t_area,
				org_t_alias."province" AS org_t_province,
				SUM(org_t_alias."order_number") AS x_dimension_metric
			FROM
				"sample_data"."lod_func_test_table" AS org_t_alias
			GROUP BY
				org_t_alias."area",
				org_t_alias."province"
			ORDER BY
				org_t_area ASC,
				org_t_province ASC
		)
	`

	lodSql := `
		(
			SELECT
				-- lod子查询：area、sum(order_number)
				org_t_alias."area" AS org_t_area,
				SUM(org_t_alias."order_number") AS lod_metric
			FROM
				"sample_data"."lod_func_test_table" AS org_t_alias
			GROUP BY
				org_t_alias."area"
		)
	`

	b := xorm.Select(
		"x_dimension_tmp_t.org_t_area AS select_area ",
		"x_dimension_tmp_t.org_t_province AS select_province",
		"x_dimension_tmp_t.x_dimension_metric AS select_common_metric",
		"lod_dimension_tmp_t.lod_metric AS select_lod_metric",
	).
		From(
			mainSql, "x_dimension_tmp_t",
		).
		InnerJoin(
			xorm.As(lodSql, "lod_dimension_tmp_t"),
			//"schema1.table1", // joinTable 可以是 string
			`x_dimension_tmp_t."org_t_area" = lod_dimension_tmp_t."org_t_area"`,
		)

	sql, _ := b.ToBoundSQL()
	fmt.Println(sql)
}
