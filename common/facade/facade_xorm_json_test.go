package facade

import (
	"fmt"
	"testing"

	"github.com/forhsd/sqlbuilder/util"

	"github.com/forhsd/logger"
	"github.com/forhsd/sqlbuilder/common/clause"
	_ "github.com/forhsd/sqlbuilder/util"
)

/**
SELECT
    DATE_TRUNC('month', "t1"."created_at") as created_at,
    "t1"."category",
    sum("t1"."price")
FROM "sample_data"."products" t1
GROUP BY DATE_TRUNC('month', "t1"."created_at"), "t1"."category"
ORDER BY DATE_TRUNC('month', "t1"."created_at") ASC, "t1"."category" ASC
*/

/**
SELECT
    DATE_TRUNC('month', CAST("sample_data"."products"."created_at" AS timestamp)) AS "created_at",
    "sample_data"."products"."category" AS "category",
    "order"."id" AS "order__id",
    SUM("sample_data"."products"."price") AS "sum"
FROM "sample_data"."products"
LEFT JOIN "sample_data"."order" AS "order" ON "sample_data"."products"."id" = "order"."product_id"
GROUP BY DATE_TRUNC('month', CAST("sample_data"."products"."created_at" AS timestamp)), "sample_data"."products"."category", "order"."id"
ORDER BY DATE_TRUNC('month', CAST("sample_data"."products"."created_at" AS timestamp)) ASC, "sample_data"."products"."category" ASC, "order"."id" ASC
*/

/**
{
    "strategy": 0,
    "driver": 0,
    "sqlBuilders": [
        {
            "deep": 0,
            "sql": {
                "from": {
                    "tableName": "tableName",
                    "tableSchema": "tableSchema",
                    "tableAlias": "tableAlias"
                },
                "join": {
                    "type": "left join",
                    "table": {
                        "tableName": "tableName",
                        "tableSchema": "tableSchema",
                        "tableAlias": "tableAlias"
                    },
                    "left": {
                        "field":"field",
                        "table":"tableName",
                        "schema":"schema",
                        "alias":"alias",
                        "hasAgg":false
                    },
                    "right":{
                        "field":"field",
                        "table":"tableName",
                        "schema":"schema",
                        "alias":"alias",
                        "hasAgg":false
                    },
                    "on": "="
                },
                "where": {
                    "call": "=",
                    "vars": [
                        {
                            "field": "field",
                            "table": "tableName",
                            "schema": "schema",
                            "alias": "alias",
                            "hasAgg": false
                        },
                        1
                    ]
                },
                // 类型断言: Column 或者 Expression
                "groupBy": [
                    {
                        "field": "field",
                        "table": "tableName",
                        "schema": "schema",
                        "alias": "alias",
                        "hasAgg": false
                    },
                    {
                        "call": "DATE_TRUNC",
                        "vars": [
                            "month",
                            {
                                "field": "field",
                                "table": "tableName",
                                "schema": "schema",
                                "alias": "alias",
                                "hasAgg": false
                            }
                        ]
                    }
                ],
                // Aggregation 会加入 Select 片段
                "agg": [
                    {
                        "call": "SUM",
                        "vars": [
                            {
                                "field": "field",
                                "table": "tableName",
                                "schema": "schema",
                                "alias": "alias",
                                "hasAgg": false
                            }
                        ]
                    }
                ],
                // 类型断言: Column 或者 Expression
                "select": [
                    {
                        "field": "field",
                        "table": "tableName",
                        "schema": "schema",
                        "alias": "alias",
                        "hasAgg": false
                    }
                ],
                // dependent 类型断言: Column 或者 Expression
                "orderBy": [
                    {
                        "order":"asc",
                        "dependent": {
                            "field": "field",
                            "table": "tableName",
                            "schema": "schema",
                            "alias": "alias",
                            "hasAgg": false
                        }
                    },
                    {
                        "order":"asc",
                        "dependent": {
                            "call": "DATE_TRUNC",
                            "vars": [
                                "month",
                                {
                                    "field": "field",
                                    "table": "tableName",
                                    "schema": "schema",
                                    "alias": "alias",
                                    "hasAgg": false
                                }
                            ]
                        }
                    }
                ],
                "limit": {
                    "limit": 10,
                    "offset": 0
                }
            }
        }
    ]
}

// .InnerJoin("table3", "table2.id = table3.tid")
// .LeftJoin("table2", Eq{"table1.id": 1}.And(Lt{"table2.id": 3})).
// .Where(Eq{"a": 1})

*/

func TestXORMSQLBuilderSelect(t *testing.T) {
	fromTable := clause.Table{TableName: "products", TableSchema: "sample_data", TableAlias: "products"}

	selects := []interface{}{
		clause.Column{
			Field:   "category",
			Table:   "products",
			Schema:  "sample_data",
			Alias:   "分类",
			AggAble: false,
			UseAs:   true,
		},
		clause.Expression{
			Call:     "sum",
			CallType: "agg",
			Vars: []interface{}{
				//"products",
				//"price",
				clause.Column{
					Field:   "price",
					Table:   "products",
					Schema:  "sample_data",
					Alias:   "price表达式内部as",
					AggAble: false,
					UseAs:   false,
				},
			},
			CallAs: "合计价格",
			UseAs:  true,
		},
		clause.Expression{
			Call:     "DATE_TRUNC",
			CallType: "inner",
			Vars: []interface{}{
				"month",
				clause.Column{
					Field:   "created_at",
					Table:   "products",
					Schema:  "sample_data",
					Alias:   "created_at表达式内部as",
					AggAble: false,
					UseAs:   false,
				},
			},
			//CallAs: "date_trunc__products_created_at",
			CallAs: "离散日期(月)",
			UseAs:  true,
		},
	}

	groupBy := []interface{}{
		clause.Expression{
			Call:     "DATE_TRUNC",
			CallType: "inner",
			Vars: []interface{}{
				"month",
				clause.Column{
					Field:   "created_at",
					Table:   "products",
					Schema:  "sample_data",
					Alias:   "created_at表达式G内部as",
					AggAble: false,
				},
			},
			CallAs: "离散日期(月)",
			UseAs:  false,
		},
		clause.Column{
			Field:   "category",
			Table:   "products",
			Schema:  "sample_data",
			Alias:   "分类",
			AggAble: false,
			UseAs:   false,
		},
	}

	sqlRef := clause.SQLReference{
		From:    fromTable,
		Select:  selects,
		GroupBy: groupBy,
	}

	request := clause.BuilderRequest{
		Driver:      clause.DriverPostgres,
		Strategy:    clause.ModelStrategy,
		SQLBuilders: []clause.DeepWrapper{{Deep: 0, Sql: sqlRef}},
	}

	serialize, _ := util.MarshalTool.Serialize(request)
	fmt.Println(string(serialize))

	builder, bErr := RunBuilder(request)
	if bErr != nil {
		logger.Error(bErr)
	}
	boundSQL, sqlErr := builder.ToBoundSQL()
	fmt.Println(boundSQL)
	fmt.Println(sqlErr)
}
