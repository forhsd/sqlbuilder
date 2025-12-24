package facade

import (
	"fmt"
	pg_query "github.com/pganalyze/pg_query_go/v5"
	"testing"
)

func TestParse(t *testing.T) {
	var sql = `
		select
			p."category",
			sum(p."price") as "category_total"
		from "sample_data"."products" p
		left join "sample_data"."order" o on p."id" = o."product_id"
		where p."category" = 'Doohickey' and o."product_id" > 10
		group by p."category"
	`

	result, err := pg_query.Parse(sql)
	if err != nil {
		panic(err)
	}

	// 只有from
	fmt.Printf("%s", result.Stmts[0].Stmt.GetSelectStmt().GetFromClause()[0].GetRangeVar())
	// join on
	fmt.Printf("%s", result.Stmts[0].Stmt.GetSelectStmt().GetFromClause()[0].GetJoinExpr())
}
