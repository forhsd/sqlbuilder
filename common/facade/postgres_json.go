package facade

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/forhsd/sqlbuilder/common/clause"
	"github.com/forhsd/sqlbuilder/util"
	xorm "xorm.io/builder"
)

func RunBuilder(request clause.BuilderRequest) (*xorm.Builder, error) {
	builders := request.SQLBuilders
	// 目前只有一层
	if len(builders) != 1 {
		return nil, errors.New("SQL Builder 暂时只支持一层")
	}
	sqlRef := builders[0].Sql
	dialectBuilder, _ := chooseDialect(request)

	selects, tableSchema, fromAlias := convert2BasicArgs(sqlRef)

	basicBuilder := buildBasic(dialectBuilder, selects, tableSchema, fromAlias)

	// 判断是否组装 Join
	if len(sqlRef.Join) > 0 {
		basicBuilder = buildJoin(basicBuilder, sqlRef)
	}

	// 判断是否组装 Where
	if len(sqlRef.Where) > 0 {
		basicBuilder = buildWhere(basicBuilder, sqlRef)
	}

	// 判断是否组装 GroupBy
	if len(sqlRef.GroupBy) > 0 {
		basicBuilder = buildGroupBy(basicBuilder, sqlRef)
	}

	// 判断是否组装 OrderBy
	if len(sqlRef.OrderBy) > 0 {
		basicBuilder = buildOrderBy(basicBuilder, sqlRef)
	}

	// 判断是否组装 Limit
	if &sqlRef.Limit != nil {
		basicBuilder = buildLimit(basicBuilder, sqlRef)
	}

	return basicBuilder, nil
}

func chooseDialect(request clause.BuilderRequest) (*xorm.Builder, error) {
	driver := request.Driver
	switch driver {
	case clause.DriverPostgres:
		return xorm.Dialect(xorm.POSTGRES), nil
	default:
		return nil, errors.New("未知SQL构建策略")
	}
}

func buildBasic(
	dialectBuilder *xorm.Builder,
	selects []string,
	sub string,
	fromAlias string) *xorm.Builder {
	// XORM 只有Select不能构造SQL, 至少需要 Select && From
	basicBuilder := dialectBuilder.Select(selects...).From(sub, fromAlias)
	return basicBuilder
}

func convert2BasicArgs(sqlRef clause.SQLReference) ([]string, string, string) {
	selectClause := sqlRef.Select

	// 类型断言: Column 或者 Expression
	selects := extraColumnOrExpressionStrings(selectClause)

	fromClause := sqlRef.From

	tableSchema := fmt.Sprintf(
		clause.PGTableColumn,
		fromClause.TableSchema,
		fromClause.TableName)

	return selects, tableSchema, fmt.Sprintf(clause.PGStringLiteralSafe, fromClause.TableAlias)
}

func extraColumnOrExpressionStrings(mixClause []interface{}) []string {
	var selects []string
	// 走http, case Column 和 case Expression 分支失效,
	for _, mixItem := range mixClause {
		//switch item := mixItem.(type) {
		//case clause.Column:
		//	// `"table"."field"`
		//	//column := mixItem.(clause.Column)
		//	format := util.Ternary(item.UseAs, clause.PGTableColumnAs, clause.PGTableColumn).(string)
		//	//selects = append(
		//	//	selects,
		//	//	calAppendTextByColumnAs(format, column, column.UseAs, column.Alias))
		//	selects = append(selects, formatColumn(item, format))
		//case clause.Expression:
		//	// fmt.Sprintf(Format, Vars...)
		//	//expression := mixItem.(clause.Expression)
		//	//vars2Strings := calExpressionVarsFormatStrings(expression)
		//	//format := getExpressionFuncFormat(expression)
		//	//selects = append(
		//	//	selects,
		//	//	fmt.Sprintf(
		//	//		format,
		//	//		vars2Strings...))
		//	selects = append(selects, formatExpression(item))
		//case map[string]interface{}:
		//	// http接收后会擦除元组内元素类型都变成map[string]interface{}导致断言失效, 因此不能用 .(type)
		//	//anyMap := mixItem.(map[string]interface{})
		//	//identify := identifyTupleItem(anyMap)
		//	//switch identify.(type) {
		//	//case clause.Column:
		//	//	// `"table"."field"` as alias 或者 `"table"."field"`
		//	//	column := identify.(clause.Column)
		//	//	formatResult := util.Ternary(column.UseAs, clause.PGTableColumnAs, clause.PGTableColumn)
		//	//	format := formatResult.(string)
		//	//	selects = append(
		//	//		selects,
		//	//		calAppendTextByColumnAs(format, column, column.UseAs, column.Alias))
		//	//case clause.Expression:
		//	//	// fmt.Sprintf(Format, Vars...)
		//	//	expression := identify.(clause.Expression)
		//	//	vars2Strings := calExpressionVarsFormatStrings(expression)
		//	//	format := getExpressionFuncFormat(expression)
		//	//	selects = append(
		//	//		selects,
		//	//		fmt.Sprintf(
		//	//			format,
		//	//			vars2Strings...))
		//	//}
		//	selects = append(selects, processMapItem(item))
		//default:
		//	continue
		//}
		result := processMixFieldItem(mixItem)
		if result != "" {
			selects = append(selects, result)
		}
	}
	return selects
}

func processMixFieldItem(mixItem interface{}) string {
	switch item := mixItem.(type) {
	case clause.Column:
		format := util.Ternary(item.UseAs, clause.PGTableColumnAs, clause.PGTableColumn).(string)
		return formatColumn(item, format)

	case clause.Expression:
		return formatExpression(item)

	case map[string]interface{}:
		// http接收后会擦除元组内元素类型都变成map[string]interface{}导致断言失效, 因此不能用 .(type)
		return processMapItem(item)

	default:
		return ""
	}
}

func formatColumn(column clause.Column, format string) string {
	return calAppendTextByColumnAs(format, column, column.UseAs, column.Alias)
}

func formatExpression(expression clause.Expression) string {
	vars2Strings := calExpressionVarsFormatStrings(expression)
	format := getExpressionFuncFormat(expression)
	return fmt.Sprintf(format, vars2Strings...)
}

func processMapItem(anyMap map[string]interface{}) string {
	identify := identifyTupleItem(anyMap)
	switch identifiedItem := identify.(type) {
	case clause.Column:
		format := util.Ternary(identifiedItem.UseAs, clause.PGTableColumnAs, clause.PGTableColumn).(string)
		return formatColumn(identifiedItem, format)
	case clause.Expression:
		return formatExpression(identifiedItem)
	default:
		return ""
	}
}

func calAppendTextByColumnAs(format string, column clause.Column, userAs bool, alias string) string {
	if userAs {
		return fmt.Sprintf(format, column.Table, column.Field, alias)
	}
	return fmt.Sprintf(format, column.Table, column.Field)
}

// identifyTupleItem http接收后会擦除元组内元素类型都变成map[string]interface{}导致断言失效, 因此不能用 .(type)
// 通过特殊字段断言类型, 不增加一层是因为不好元组元素索引
func identifyTupleItem(item map[string]interface{}) interface{} {
	// Column
	if value := item["field"]; value != nil {
		var column clause.Column
		err := mapToStruct(item, &column)
		if err != nil {
			fmt.Println("Column转换失败:", err)
			return nil
		}
		return column
	}
	// Expression
	if value := item["call"]; value != nil {
		var expression clause.Expression
		err := mapToStruct(item, &expression)
		if err != nil {
			fmt.Println("Expression转换失败:", err)
			return nil
		}
		return expression
	}
	return nil
}

func mapToStruct(data map[string]interface{}, out interface{}) error {
	// 将 map 转换为 JSON 字符串
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("转换为JSON失败: %v", err)
	}

	// 将 JSON 字符串反序列化为结构体
	err = json.Unmarshal(jsonData, out)
	if err != nil {
		return fmt.Errorf("反序列化失败: %v", err)
	}

	return nil
}

func getExpressionFuncFormat(expression clause.Expression) string {
	callType := expression.CallType
	if expression.UseAs {
		switch callType {
		case clause.CallTypeAgg:
			return clause.PGAggregationAsFuncFormatMap[clause.AggFunc(expression.Call)]
		case clause.CallTypeInner:
			return clause.InnerAsFuncFormatMap[clause.InnerFunc(expression.Call)]
		default:
			return ""
		}
	}

	switch callType {
	case clause.CallTypeAgg:
		return clause.AggregationFuncFormatMap[clause.AggFunc(expression.Call)]
	case clause.CallTypeInner:
		return clause.InnerFuncFormatMap[clause.InnerFunc(expression.Call)]
	default:
		return ""
	}
}

func calExpressionVarsFormatStrings(expr clause.Expression) []interface{} {
	vars := expr.Vars
	// 表达式参数: int, string, clause.Column
	var expressions []interface{}
	for _, varItem := range vars {
		switch item := varItem.(type) {
		case clause.Column:
			// `"alias"."field"`
			//tempColumn := varItem.(clause.Column)
			//format := util.Ternary(tempColumn.UseAs, clause.PGTableColumnAs, clause.PGTableColumn).(string)
			//expressions = append(expressions, calAppendTextByColumnAs(format, tempColumn, tempColumn.UseAs, expr.CallAs))
			format := util.Ternary(item.UseAs, clause.PGTableColumnAs, clause.PGTableColumn).(string)
			expressions = append(expressions, calAppendTextByColumnAs(format, item, item.UseAs, expr.CallAs))
		case string:
			exprString := varItem.(string)
			expressions = append(expressions, exprString)
		case int, int8, int16, int32, int64:
			exprInt := varItem.(int)
			expressions = append(expressions, fmt.Sprintf(clause.AnyLiteral, exprInt))
		case map[string]interface{}:
			// http接收后会擦除元组内元素类型都变成map[string]interface{}导致断言失效, 因此不能用 .(type)
			//anyMap := varItem.(map[string]interface{})
			//identify := identifyTupleItem(anyMap)
			//switch identify.(type) {
			//case clause.Column:
			//	// `"alias"."field"`
			//	column := identify.(clause.Column)
			//	format := util.Ternary(column.UseAs, clause.PGTableColumnAs, clause.PGTableColumn).(string)
			//	expressions = append(expressions, calAppendTextByColumnAs(format, column, column.UseAs, expr.CallAs))
			//}
			expressions = append(expressions, processVarsMapItem(item, expr.CallAs))
		default:
			continue
		}
	}

	if expr.UseAs {
		expressions = append(expressions, expr.CallAs)
	}

	return expressions
}

func processVarsMapItem(anyMap map[string]interface{}, callAs string) interface{} {
	identify := identifyTupleItem(anyMap)
	switch identifiedItem := identify.(type) {
	case clause.Column:
		format := util.Ternary(identifiedItem.UseAs, clause.PGTableColumnAs, clause.PGTableColumn).(string)
		return calAppendTextByColumnAs(format, identifiedItem, identifiedItem.UseAs, callAs)
	default:
		return nil // Handle unexpected types or return an appropriate default value
	}
}

func buildGroupBy(builder *xorm.Builder, sqlRef clause.SQLReference) *xorm.Builder {
	groupByClause := sqlRef.GroupBy
	// 类型断言: Column 或者 Expression
	groupByItems := extraColumnOrExpressionStrings(groupByClause)

	// 使用 strings.Join 来合并切片元素，以逗号分隔
	joinedItems := strings.Join(groupByItems, clause.COMMA)

	// xorm 的 GroupBy 方法只接收一个字符串参数，因此要自己将多个 GroupBy 拼接为1个
	groupByFragment := fmt.Sprintf(clause.StringLiteral, joinedItems)

	return builder.GroupBy(groupByFragment)
}

func buildJoin(builder *xorm.Builder, sqlRef clause.SQLReference) *xorm.Builder {
	return builder
}

func buildWhere(builder *xorm.Builder, sqlRef clause.SQLReference) *xorm.Builder {
	return builder
}

func buildOrderBy(builder *xorm.Builder, sqlRef clause.SQLReference) *xorm.Builder {
	orderByClause := sqlRef.OrderBy
	// 提取 orderBy 字符串片段
	var orderBys []string
	for _, orderBy := range orderByClause {
		field := processMixFieldItem(orderBy.Dependent)
		orderByItem := fmt.Sprintf(clause.DoubleStringLiteral, field, orderBy.Order)
		orderBys = append(orderBys, orderByItem)
	}
	// 合并成一个字符串片段
	orderByItems := strings.Join(orderBys, clause.COMMA)
	// xorm 的 orderBy 方法只接收一个字符串参数，因此要自己将多个 orderBy 拼接为1个
	orderByFragment := fmt.Sprintf(clause.StringLiteral, orderByItems)

	return builder.OrderBy(orderByFragment)
}

func buildLimit(builder *xorm.Builder, sqlRef clause.SQLReference) *xorm.Builder {
	return builder
}
