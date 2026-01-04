package clause

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	xorm "xorm.io/builder"
)

func TestCaseWhenCond(t *testing.T) {

	sql, err := xorm.ToBoundSQL(xorm.Eq{"name": 1})
	if err != nil {
		panic(err)
	}
	fmt.Println("xorm cond 建模转换sql片段基本用法")
	fmt.Println(sql)
	fmt.Println()
	//fmt.Println(xorm.ToBoundSQL(xorm.Expr(sql)))

	caseWhen := CaseWhen{
		Conditions: []CaseWhenCondition{
			{
				When: xorm.Gte{"age": 1}.And(xorm.Lt{"age": 18}),
				Then: "未成年",
			},
			{
				When: xorm.Gt{"name": 18},
				Then: "成年",
			},
		},
		ElseValue: "婴幼儿",
		Alias:     "age_group",
	}

	// case-when 片段
	fragment := caseWhen.BuilderCaseWhenFragment(DriverPostgres)
	fmt.Printf("case-when 内容: \n%s ", fragment)
	fmt.Println()

	// case-when 子查询内容
	basicBuilder := xorm.Dialect(xorm.POSTGRES).Select(fragment).From(`"sample_data"."person"`)
	sql, err = basicBuilder.ToBoundSQL()
	if err != nil {
		panic(err)
	}
	fmt.Printf("select case when : \n%s", sql)
	fmt.Println()
	expected := `SELECT CASE  WHEN age>=1 AND age<18 THEN '未成年' WHEN name>18 THEN '成年' ELSE '婴幼儿' END as "age_group" FROM "sample_data"."person"`
	assert.EqualValues(t, expected, sql)
}

func TestIsVariadicArgsFunc(t *testing.T) {
	assert.EqualValues(t, true, IsVariadicArgsFunc("concat"))
	assert.EqualValues(t, false, IsVariadicArgsFunc("array"))
	assert.EqualValues(t, false, IsVariadicArgsFunc("sum"))

	// 参数>占位符: qqq add ccc%!(EXTRA string=ddd)
	vars := []any{"qqq", "ccc", "ddd"}
	fmt.Println(fmt.Sprintf(`%s add %s`, vars...))
	fmt.Println(CalArithFormat("+", 3))
	fmt.Println(CalArithFormatWithBuilder("+", 3))
}
