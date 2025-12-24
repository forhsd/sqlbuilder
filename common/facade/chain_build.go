package facade

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/forhsd/logger"
	"github.com/forhsd/sqlbuilder/common/clause"
	pb "github.com/forhsd/sqlbuilder/gen/proto"
	"github.com/forhsd/sqlbuilder/util"
	xorm "xorm.io/builder"
)

// -----------------------------------------------   facade: 通过模型构建sql跨方言入口  --------------------------------------------

func ChainSqlBuilder(
	appendBuilder *xorm.Builder,
	dialectBuilder *xorm.Builder,
	sqlRef *pb.SqlReference,
	ctx *ModelBuilderCtx) *xorm.Builder {
	// builder(selects, from , as)
	basicBuilder := buildChainBasic(appendBuilder, dialectBuilder, sqlRef, ctx)
	// 调用BuildOther方法处理其他逻辑
	builder := buildChainOther(basicBuilder, sqlRef, ctx)
	return builder
}

// 抽象方法: 构建 select & from, 具体实现由子类提供
func buildChainBasic(
	appendBuilder *xorm.Builder,
	dialectBuilder *xorm.Builder,
	sqlRef *pb.SqlReference,
	ctx *ModelBuilderCtx) *xorm.Builder {
	logger.Debug("[模板方法模式]- 实现: 构建 builder 基础内容...")

	// 嵌套sql, 构建 select + from
	if appendBuilder != nil {
		nestBasicBuilder, err := buildNestBasicByProto(appendBuilder, dialectBuilder, sqlRef, ctx)
		if err != nil {
			panic(err.Error())
		}
		return nestBasicBuilder
	}
	// 一层sql, 构建 select + from
	selects, tableSchema, fromAlias := createBasicArgsByProto(sqlRef, ctx)

	basicBuilder := buildBasicByProto(dialectBuilder, selects, tableSchema, fromAlias)

	return basicBuilder
}

// 抽象方法: 构建SQL其他关键字内容，具体实现由子类提供
func buildChainOther(
	builder *xorm.Builder,
	sqlRef *pb.SqlReference,
	ctx *ModelBuilderCtx) *xorm.Builder {
	logger.Debug("[模板方法模式]- 实现: 构建 builder 其他内容...")

	// 判断是否组装 Join
	if len(sqlRef.Join) > 0 {
		builder = buildJoinByProto(builder, sqlRef, ctx)
	}

	// 判断是否组装 Where
	if len(sqlRef.Where) > 0 || sqlRef.LogicWhere != nil {
		builder = buildWhereByProto(builder, sqlRef, ctx)
	}

	// 判断是否组装 GroupBy
	if len(sqlRef.GroupBy) > 0 {
		builder = buildGroupByByProto(builder, sqlRef, ctx)
	}

	// 判断是否组装 OrderBy
	if len(sqlRef.OrderBy) > 0 {
		builder = buildOrderByProto(builder, sqlRef, ctx)
	}

	// 判断是否组装 Limit
	if sqlRef.Limit != nil {
		builder = buildLimitProto(builder, sqlRef)
	}

	return builder
}

// -----------------------------------------------   facade: 通过模型构建sql 公共函数  --------------------------------------------

// buildNestBasicByProto 构建嵌套sql的 select + from
// 参数:
//   - appendBuilder: 上一次的 xorm.Builder 指针对象
//   - dialectBuilder: 当前层级 *xorm.Builder 初始化指针对象
//   - sqlRef: 当前层级的sqlRef对象
//
// 返回值:
//   - 返回 *xorm.Builder
func buildNestBasicByProto(
	appendBuilder *xorm.Builder,
	dialectBuilder *xorm.Builder,
	sqlRef *pb.SqlReference,
	ctx *ModelBuilderCtx,
) (*xorm.Builder, error) {
	from := sqlRef.GetFrom()

	// 如果 from 类型是 MixTable.LiteralTable 报错
	if _, ok := from.GetMt().(*pb.MixTable_LiteralTable); ok {
		return nil, errors.New("模型层级构建模式不支持字面量Table")
	}
	// 强转
	normalTable := from.GetNormalTable()

	// 类型断言: Column 或者 Expression
	selects := extraColumnOrExpressionStringsByProto(sqlRef.GetSelect(), ctx)

	string1LiteralSafeFormat := ctx.String1LiteralSafeFormat
	nestBasicBuilder := dialectBuilder.
		Select(selects...).
		From(appendBuilder, fmt.Sprintf(string1LiteralSafeFormat, normalTable.GetTableAlias()))

	return nestBasicBuilder, nil
}

// buildBasicByProto 构建基本的SQL查询器
// 该函数旨在通过XORM库构建一个基本的SQL查询器对象
// 参数:
//   - dialectBuilder: XORM的Builder对象，用于构建SQL语句
//   - selects: 需要选择的列的列表
//   - sub: &schema.table/子查询的标识符
//   - fromAlias: &schema.table的表别名/FROM子句中的表别名
//
// 返回值:
//   - 返回一个构建好的SQL查询器对象
func buildBasicByProto(
	dialectBuilder *xorm.Builder,
	selects []string,
	sub string,
	fromAlias string) *xorm.Builder {
	// XORM 只有Select不能构造SQL, 至少需要 Select && From
	basicBuilder := dialectBuilder.Select(selects...).From(sub, fromAlias)
	return basicBuilder
}

func createBasicArgsByProto(sqlRef *pb.SqlReference, ctx *ModelBuilderCtx) ([]string, string, string) {
	selectClause := sqlRef.Select

	// 类型断言: Column 或者 Expression
	selects := extraColumnOrExpressionStringsByProto(selectClause, ctx)

	fromClause := sqlRef.From
	// x 的引号表达
	string1LiteralSafeFormat := ctx.String1LiteralSafeFormat
	// x.y 的引号表达
	string2LiteralSafeFormat := ctx.String2LiteralSafeFormat

	switch v := fromClause.GetMt().(type) {
	case *pb.MixTable_NormalTable:
		normalTable := v.NormalTable
		tableSchema := fmt.Sprintf(string2LiteralSafeFormat, normalTable.TableSchema, normalTable.TableName)
		return selects, tableSchema, fmt.Sprintf(string1LiteralSafeFormat, normalTable.TableAlias)
	case *pb.MixTable_LiteralTable:
		literalTable := v.LiteralTable
		// 第二个参数是子查询字面量加小括号，第三个参数是别名
		return selects,
			fmt.Sprintf(clause.PntFormat, literalTable.GetSubLiteral().GetLiteral()),
			fmt.Sprintf(string1LiteralSafeFormat, literalTable.GetLiteralAlias())
	default:
		panic("未知Table类型")
	}
}

// 将MixField转换为对应的string字符串
func extraColumnOrExpressionStringsByProto(mixClause []*pb.MixField, ctx *ModelBuilderCtx) []string {
	var selects []string
	for _, mixItem := range mixClause {
		result := processMixFieldItemByProto(mixItem, ctx)
		if result != "" {
			selects = append(selects, result)
		}
	}
	return selects
}

func processMixFieldItemByProto(mixItem *pb.MixField, ctx *ModelBuilderCtx) string {
	string2LiteralSafeFormat := ctx.String2LiteralSafeFormat
	string2LiteralAsSafeFormat := ctx.String2LiteralAsSafeFormat
	// 使用类型断言判断具体是 Column 还是 Expression
	switch v := mixItem.GetMix().(type) {
	case *pb.MixField_Column:
		column := v.Column
		format := util.Ternary(column.UseAs, string2LiteralAsSafeFormat, string2LiteralSafeFormat).(string)
		return formatColumnByProto(column, format)
	case *pb.MixField_Expression:
		expression := v.Expression
		return formatExpressionByProto(expression, ctx)
	case *pb.MixField_CaseWhen:
		caseWhen := v.CaseWhen
		return formatCaseWhenByProto(caseWhen, ctx)
	default:
		return ""
	}
}

// 构建 case-when 类型的格式化字符串填充
func formatCaseWhenByProto(caseWhen *pb.CaseWhen, ctx *ModelBuilderCtx) string {

	// proto -> struct(case-when) -> string
	caseWhenThens := formatCaseWhenItemByProto(caseWhen, ctx)

	// 存在 case-when 的 else 分支才提取值并填充 else 分支
	ev := util.Ternary(caseWhen.ElseValue == nil, nil, ExtraArgItemValue(caseWhen.ElseValue))

	caseWhenStruct := clause.CaseWhen{
		Conditions: caseWhenThens,
		// case-when 默认没有 else 分支
		ElseValue: ev,
		Alias:     caseWhen.Alias,
		UseAs:     caseWhen.UseAs,
	}

	fragment := caseWhenStruct.BuilderCaseWhenFragment(ctx.Driver)
	// *xorm.Builder.Select(item...)  会自动填充逗号, 这里返回当前 items 切片的填充内容
	return fragment
}

func formatCaseWhenItemByProto(caseWhen *pb.CaseWhen, ctx *ModelBuilderCtx) []clause.CaseWhenCondition {
	var conditions []clause.CaseWhenCondition

	for _, c := range caseWhen.GetConditions() {
		thenValue := ExtraArgItemValue(c.GetThen())
		whenValue := mixWhere2Condition(c.GetWhen(), ctx)
		// 构造一行 case-when-then 语句
		whenThenLine := clause.CaseWhenCondition{
			Then: thenValue,
			When: whenValue,
		}

		conditions = append(conditions, whenThenLine)
	}

	return conditions
}

// 构建 Column 类型的格式化字符串填充
func formatColumnByProto(column *pb.Column, format string) string {
	return calAppendTextByColumnAsByProto(format, column, column.UseAs, column.Alias)
}

// formatExpressionByProto 构建 Expression 类型的格式化字符串填充
func formatExpressionByProto(expression *pb.Expression, ctx *ModelBuilderCtx) string {
	// 通常表达式的format中占位符和表达式参数是一一对应的拉链结构
	vars2Strings := calExpressionVarsFormatStringsByProto(expression, ctx)
	format, currentExprIsLiteral := getExpressionFuncFormatByProto(expression, ctx, vars2Strings)
	// 当前嵌套层是字面量表达式不需要fmt填充参数
	if currentExprIsLiteral {
		if expression.UseAs {
			return fmt.Sprintf(format, expression.GetCallAs())
		}
		return format
	}
	return fmt.Sprintf(format, vars2Strings...)
}

func calAppendTextByColumnAsByProto(format string, column *pb.Column, userAs bool, alias string) string {
	if userAs {
		return fmt.Sprintf(format, column.Table, column.Field, alias)
	}
	return fmt.Sprintf(format, column.Table, column.Field)
}

// 获取Expression对应的格式化字符串
func getExpressionFuncFormatByProto(expression *pb.Expression, ctx *ModelBuilderCtx, tmpVars2Strings []interface{}) (string, bool) {
	// 动态构建 Expression 的 format
	var finalFormat string

	// 初始化 Expression 的 format, 取不要 as 的那一部分
	basicFormat, currentExprIsLiteral := expressionFuncNoAsFormat(expression, tmpVars2Strings)
	finalFormat = basicFormat

	// 是否使用小括号
	if expression.UsePnt {
		finalFormat = fmt.Sprintf(clause.PntFormat, finalFormat)
	}

	// 是否使用 as 别名
	// Expression 表达式只有在select关键字使用alias
	asFormat := ctx.AsFormatLiteralSafeFormat
	if expression.UseAs {
		finalFormat = fmt.Sprintf(asFormat, finalFormat)
	}

	return finalFormat, currentExprIsLiteral
}

// 获取Expression没有As的格式化字符串部分
// 返回值:
//   - 驱动对应的 fmt 字符串
//   - 当前层级表达式是否是字面量
func expressionFuncNoAsFormat(expression *pb.Expression, tmpVars2Strings []interface{}) (string, bool) {
	callType := expression.CallType
	switch callType {
	case pb.CallType_CALL_TYPE_AGG:
		return clause.AggregationFuncFormatMap[clause.AggFunc(expression.Call)], false
	case pb.CallType_CALL_TYPE_INNER:
		// 根据模板字符串字面量的表达式格式化得到表达式对应的字符串字面量
		if clause.IsTmplLiteralFunc(expression.Call) {
			return FormatTmplLiteralInnerFunc(expression.Call, tmpVars2Strings), true
		}
		return clause.InnerFuncFormatMap[clause.InnerFunc(expression.Call)], false
	case pb.CallType_CALL_TYPE_ARITH:
		// 算术表达式的format要结合参数个数动态生成
		return clause.CalArithFormatWithBuilder(expression.Call, len(expression.Vars)), false
	case pb.CallType_CALL_TYPE_LITERAL:
		return expression.GetStrLiteral().GetLiteral(), false
	case pb.CallType_CALL_TYPE_BOOLEAN:
		return clause.BoolFormatMap[expression.Call], false
	default:
		return "", false
	}
}

// 构建Expression表达式占位符所需的字符串填充
func calExpressionVarsFormatStringsByProto(expr *pb.Expression, ctx *ModelBuilderCtx) []interface{} {
	vars := expr.Vars
	// 表达式参数: int, string, clause.Column
	var expressions []interface{}

	// 1. 字面量只需要填充别名
	// 2. 如果函数是 count1/now, 只用填充别名
	if expr.CallType == pb.CallType_CALL_TYPE_LITERAL || clause.IsNoneArgFunc(expr.Call) {
		if expr.UseAs {
			expressions = append(expressions, expr.CallAs)
		}
		return expressions
	}

	// var的索引 -> 是否转义, 只针对 Column 和 Expression, MultiCondition 类型的 var, 不能使用值, mapKey会被覆盖
	var varIndexTypeMap = make(map[int]bool)

	string2LiteralSafeFormat := ctx.String2LiteralSafeFormat
	string2LiteralAsSafeFormat := ctx.String2LiteralAsSafeFormat
	for index, varItem := range vars {
		switch v := varItem.GetVars().(type) {
		case *pb.MixVars_Column:
			column := v.Column
			format := util.Ternary(column.UseAs, string2LiteralAsSafeFormat, string2LiteralSafeFormat).(string)
			colLiteral := calAppendTextByColumnAsByProto(format, column, column.UseAs, expr.CallAs)
			varIndexTypeMap[index] = false
			expressions = append(expressions, colLiteral)
		case *pb.MixVars_Context:
			exprString := v.Context
			// 字符串由于此处已经转义过, 后面处理可变参数不需要转义, 和 Column/Expression, 不需要加单引号
			varIndexTypeMap[index] = false
			varStrLiteral := fmt.Sprintf(clause.StringSafeLiteral, exprString)
			expressions = append(expressions, varStrLiteral)
		case *pb.MixVars_StrLiteral:
			exprStrLiteral := v.StrLiteral
			varIndexTypeMap[index] = false
			expressions = append(expressions, exprStrLiteral.Literal)
		case *pb.MixVars_Number:
			exprInt := v.Number
			varNumLiteral := fmt.Sprintf(clause.AnyLiteral, exprInt)
			expressions = append(expressions, varNumLiteral)
		case *pb.MixVars_DoubleNum:
			exprDouble := v.DoubleNum
			doubleNumLiteral := fmt.Sprintf(clause.AnyLiteral, exprDouble)
			expressions = append(expressions, doubleNumLiteral)
		case *pb.MixVars_Expression:
			exprVarExpression := v.Expression
			varExpItemStr := formatExpressionByProto(exprVarExpression, ctx)
			varIndexTypeMap[index] = false
			expressions = append(expressions, varExpItemStr)
		case *pb.MixVars_MultiCondition:
			exprMultiCondition := v.MultiCondition
			varMultiCondItemStr := formatExprVarMultiCond2Str(exprMultiCondition.GetConditions(), ctx)
			varIndexTypeMap[index] = false
			expressions = append(expressions, varMultiCondItemStr)
		case *pb.MixVars_Cw:
			cw := v.Cw
			cwStr := formatCaseWhenByProto(cw, ctx)
			varIndexTypeMap[index] = false
			expressions = append(expressions, cwStr)
		default:
			continue
		}
	}

	// 如果是可变参数函数，将参数转换为一个字符串
	if clause.IsVariadicArgsFunc(expr.Call) {
		expressions = []interface{}{variadicArgsFuncVars2oneVar(expressions, varIndexTypeMap)}
	}

	if expr.UseAs {
		expressions = append(expressions, expr.CallAs)
	}

	return expressions
}

// 将可变参数函数的参数转换为一个字符串
func variadicArgsFuncVars2oneVar(vars []interface{}, varIndexTypeMap map[int]bool) string {
	var builder strings.Builder
	// 遍历 args，将每个参数拼接成适合的 SQL 字符串
	for i, arg := range vars {
		// 对每个参数进行转换，确保它是一个有效的 SQL 字符串
		if i > 0 {
			builder.WriteString(", ")
		}

		switch v := arg.(type) {
		case string:
			// 如果是字符串类型原本是 Column 和 Expression, MultiCondition, 不需要加单引号
			if varIndexTypeMap[i] == false {
				builder.WriteString(fmt.Sprintf("%s", v))
				break
			}
			// 字符串类型加上单引号
			builder.WriteString(fmt.Sprintf("'%s'", v))

		case int:
			// int 类型转换为字符串
			builder.WriteString(strconv.Itoa(v))
		case int32:
			// int32 类型转换为字符串
			builder.WriteString(strconv.FormatInt(int64(v), 10))
		case int64:
			// 数字类型直接转换为字符串
			builder.WriteString(strconv.FormatInt(v, 10))
		case float32, float64:
			// 浮点数类型直接转换为字符串
			builder.WriteString(fmt.Sprintf("%f", v))
		default:
			// 默认类型处理（其他类型，可以根据需要扩展）
			builder.WriteString(fmt.Sprintf("%v", arg))
		}
	}

	return builder.String()
}

// 构建 sql.join
func buildJoinByProto(builder *xorm.Builder, sqlRef *pb.SqlReference, ctx *ModelBuilderCtx) *xorm.Builder {
	joins := sqlRef.GetJoin()
	for _, join := range joins {
		// 每次将返回的新 Builder 赋值给变量
		builder = buildJoinItemByProto(builder, join, ctx)
	}
	return builder
}

func buildJoinItemByProto(builder *xorm.Builder, join *pb.Join, ctx *ModelBuilderCtx) *xorm.Builder {
	joinCond := buildJoinCondByProtoOuter(join.GetJoinCondition(), ctx)
	joinSchemaTable := buildJoinSchemaTableByProto(join, ctx)

	// 由于 xorm 强制要求 join 必须有 on, 因此针对 joinCond = 空数组时，builder 中填充 [on 特殊字符串]，在 toSql 时去掉 [on 特殊字符串]
	if joinCond == nil || strings.TrimSpace(fmt.Sprint(joinCond)) == "" {
		joinCond = xorm.Expr(clause.CrossJoinPlaceholder)
	}

	switch join.GetType() {
	case pb.JoinType_JOIN_TYPE_LEFT:
		return builder.LeftJoin(joinSchemaTable, joinCond)
	case pb.JoinType_JOIN_TYPE_RIGHT:
		return builder.RightJoin(joinSchemaTable, joinCond)
	case pb.JoinType_JOIN_TYPE_INNER:
		return builder.InnerJoin(joinSchemaTable, joinCond)
	case pb.JoinType_JOIN_TYPE_FULL:
		return builder.FullJoin(joinSchemaTable, joinCond)
	case pb.JoinType_JOIN_TYPE_CROSS:
		return builder.CrossJoin(joinSchemaTable, joinCond)
	default:
		return builder
	}
}

func buildJoinSchemaTableByProto(join *pb.Join, ctx *ModelBuilderCtx) interface{} {
	mixTable := join.GetTable()

	switch v := mixTable.GetMt().(type) {
	case *pb.MixTable_NormalTable:
		normalTable := mixTable.GetNormalTable()
		string2LiteralSafeFormat := ctx.String2LiteralSafeFormat
		joinSchemaTable := fmt.Sprintf(string2LiteralSafeFormat, normalTable.TableSchema, normalTable.TableName)
		return joinSchemaTable
	case *pb.MixTable_LiteralTable:
		literalTable := v.LiteralTable
		literalTableFormat := ctx.TableLiteralAsFormatLiteralSafeFormat
		// 坑1: 原生As不会根据方言带引号, 只能代码自己format处理别名引号; 坑2: 一次性format只需要%s就行，但是format(str,带占位符strX)其中strX必须是%%s
		return fmt.Sprintf(literalTableFormat, literalTable.GetSubLiteral().GetLiteral(), literalTable.GetLiteralAlias())

	default:
		panic("未知Join-Table类型")
	}
}

func buildJoinCondByProtoOuter(joinCondProto *pb.JoinCondition, ctx *ModelBuilderCtx) interface{} {
	switch jc := joinCondProto.GetJc().(type) {
	case *pb.JoinCondition_MultiCond:
		multiCond := jc.MultiCond
		return buildJoinCondByProto(multiCond.GetJoinCond(), ctx)
	case *pb.JoinCondition_CondLiteral:
		literalOn := jc.CondLiteral
		return xorm.Expr(literalOn.GetLiteral())
	}
	return "1=1"
}

func buildJoinCondByProto(joinCondProto []*pb.JoinCond, ctx *ModelBuilderCtx) interface{} {
	// 注意: 规定 []JoinCond 内的元素索引优先级: 1. OnField > onCond(or连接) > onCond(and连接), 否则会导致语义不对
	var joinConds []string

	var allCond xorm.Cond
	for i, jcItem := range joinCondProto {
		switch jc := jcItem.GetMix().(type) {
		case *pb.JoinCond_OnField:
			onField := jc.OnField
			left := processMixFieldItemByProto(onField.GetLeft(), ctx)
			right := processMixFieldItemByProto(onField.GetRight(), ctx)
			op := op2ConditionString(onField.GetOn())
			logic := logic2String(onField.GetLogic())
			// 返回字符串 e.g. "t1"."id" = "t2"."t1_id"
			onFieldStr := fmt.Sprintf(clause.OnFieldFormat, left, op, right)
			joinConds = append(joinConds, onFieldStr)
			if i != len(joinCondProto)-1 {
				joinConds = append(joinConds, logic)
			}
		case *pb.JoinCond_OnCond:
			onCond := jc.OnCond
			// 收集所有 Condition
			allCond = buildWhereConditions(allCond, onCond, ctx)
		}
	}

	// Condition 合并加在最后
	if nil != allCond {
		onCondSqlStr, _ := xorm.ToBoundSQL(allCond)
		joinConds = append(joinConds, fmt.Sprintf(clause.SafeCond, onCondSqlStr))
	}

	// join-on 条件融合为一个字符串
	finalJoinCond := strings.Join(joinConds, clause.Space)

	return finalJoinCond
}

func formatExprVarMultiCond2Str(wheres []*pb.MixWhere, ctx *ModelBuilderCtx) string {
	xormCond := mixWhere2Condition(wheres, ctx)
	multiCondStr, err := xorm.ToBoundSQL(xormCond)
	if err != nil {
		panic("构建 MultiCondition 失败")
	}
	return multiCondStr
}

// 构建 sql.where 总函数
func buildWhereByProto(builder *xorm.Builder, sqlRef *pb.SqlReference, ctx *ModelBuilderCtx) *xorm.Builder {
	var xormCond xorm.Cond

	if sqlRef.LogicWhere != nil && len(sqlRef.LogicWhere.Children) > 0 {
		// ✅ 使用新结构（优先方式）
		xormCond = buildCondFromLogicGroup(sqlRef.LogicWhere, ctx)
	} else if len(sqlRef.Where) > 0 {
		// ✅ 向下兼容老结构
		xormCond = mixWhere2Condition(sqlRef.GetWhere(), ctx)
	}

	return builder.Where(xormCond)
}

// 构建 sql.where-group-cond
func buildCondFromLogicGroup(group *pb.LogicGroup, ctx *ModelBuilderCtx) xorm.Cond {
	var cond xorm.Cond

	for _, node := range group.Children {
		var subCond xorm.Cond

		switch t := node.Node.(type) {
		case *pb.LogicNode_Leaf:
			subCond = mixWhere2ConditionWithoutGroup(t.Leaf, ctx)

		case *pb.LogicNode_Group:
			subCond = buildCondFromLogicGroup(t.Group, ctx)

		default:
			continue // 忽略未识别节点
		}

		// 组合逻辑（默认 AND）
		if cond == nil {
			cond = subCond
		} else {
			switch group.Logic {
			case pb.Logic_LOGIC_OR:
				cond = xorm.Or(cond, subCond)
			default:
				cond = xorm.And(cond, subCond)
			}
		}
	}

	// 当前片段是否加括号, 与下一个逻辑连接
	if group.UsePnt {
		sqlStr, _ := xorm.ToBoundSQL(cond)
		// 由于 xorm.Expr() 源码内会自动对 sql片段加括号, 因此不需要主动再次添加括号
		//return xorm.Expr(fmt.Sprintf(clause.PntFormat, sqlStr))
		return xorm.Expr(sqlStr)
	}

	return cond
}

func mixWhere2ConditionWithoutGroup(where *pb.MixWhere, ctx *ModelBuilderCtx) xorm.Cond {
	switch f := where.Filter.(type) {
	case *pb.MixWhere_Condition:
		return buildWhereConditionsWithOuterLogic(f.Condition, ctx)
	case *pb.MixWhere_Expression:
	}
	return nil
}

// // 构建 sql.where-cond
func mixWhere2Condition(wheres []*pb.MixWhere, ctx *ModelBuilderCtx) xorm.Cond {
	// 每组内部 n * cond, 待融合
	var eachGroupXormCond xorm.Cond
	// 组之间的 cond
	var collectGroupCondSlice []xorm.Cond
	whereSize := len(wheres)

	// 对比水位线: 用于对比上一个cond和当前cond是否是同一个组内的cond
	var wtGroupId string
	for i, whereItem := range wheres {
		switch f := whereItem.GetFilter().(type) {
		case *pb.MixWhere_Condition:
			currentCond := f.Condition
			// 为什么抽取方法后 collectGroupCondSlice 每次经过forr都没有刷新值,而都是nil
			currentCondGroupId := currentCond.GetGroupId()
			if wtGroupId == "" {
				wtGroupId = currentCondGroupId
			}
			// 组id不同, 融合之前cond转换为 Expr, 收集转换后的 Expr, 并重置 eachGroupXormCond
			if wtGroupId != currentCondGroupId {
				// 置换组内多个Cond 转换为1个 Expr, 添加到切片
				groupedCondSqlStr, _ := xorm.ToBoundSQL(eachGroupXormCond)
				collectGroupCondSlice = append(collectGroupCondSlice, xorm.Expr(groupedCondSqlStr))
				// 结束后每一组cond重置, 刷新wtGroupId
				eachGroupXormCond = nil
				wtGroupId = currentCondGroupId
			}
			eachGroupXormCond = buildWhereConditions(eachGroupXormCond, currentCond, ctx)
			// 最后一个解析的要添加进去
			if i == whereSize-1 {
				collectGroupCondSlice = append(collectGroupCondSlice, eachGroupXormCond)
			}
		case *pb.MixWhere_Expression:
			// 暂时不考虑 Where-Expression, e.g. WHERE ARRAY_CONTAINS(tags, 'urgent')
		}
	}

	var finalXormCond xorm.Cond
	// 遍历 toConvertCond 使用 and 连接 = > xorm.And(toConvertCond元素1, toConvertCond元素2 ...)
	for _, condItem := range collectGroupCondSlice {
		finalXormCond = xorm.And(finalXormCond, condItem)
	}

	return finalXormCond
}

func buildWhereConditionsWithOuterLogic(condition *pb.Condition, ctx *ModelBuilderCtx) xorm.Cond {
	currentCondTemp := buildWhereConditionItem(condition, ctx)
	// 通过 reverse 判断是否对当前条件取反
	currentCondition1 := refreshReverseWhereConditionItem(condition, currentCondTemp)
	currentCondition := refreshPntWhereConditionItem(condition, currentCondition1)
	return currentCondition
}

// 构建 sql.where-cond 的 cond 整个片段
func buildWhereConditions(xormCond xorm.Cond, condition *pb.Condition, ctx *ModelBuilderCtx) xorm.Cond {
	currentCondTemp := buildWhereConditionItem(condition, ctx)
	// 通过 reverse 判断是否对当前条件取反
	currentCondition1 := refreshReverseWhereConditionItem(condition, currentCondTemp)
	currentCondition := refreshPntWhereConditionItem(condition, currentCondition1)
	if nil == xormCond {
		// 没有where条件, 作为第一个条件
		return currentCondition
	}
	// 处理当前条件和之前条件的逻辑连接; Where().And().Or(), 会将 当前cond片段 和 之前cond片段(左侧片段可能很长)分别视作两个不同整体
	if pb.Logic_LOGIC_AND == condition.Logic {
		return xorm.And(xormCond, currentCondition)
	}
	return xorm.Or(xormCond, currentCondition)
}

// NOT 反转 sql.where-cond 的 cond 其中一个片段
func refreshReverseWhereConditionItem(pbCond *pb.Condition, builderCond xorm.Cond) xorm.Cond {
	if pbCond.GetReverse() {
		return xorm.Not{builderCond}
	}
	return builderCond
}

// 加括号: sql.where-cond 的 cond 其中一个片段
func refreshPntWhereConditionItem(pbCond *pb.Condition, builderCond xorm.Cond) xorm.Cond {
	// 是否使用小括号
	if pbCond.GetUsePnt() {
		finalCondExprLiteral, _ := xorm.ToBoundSQL(builderCond)
		var finalCondStr string
		finalCondStr = fmt.Sprintf(clause.PntFormat, finalCondExprLiteral)
		return xorm.Expr(finalCondStr)
	}
	return builderCond
}

// sql.where-cond 的 cond 其中一个片段的表达式构建
func buildWhereConditionItem(cond *pb.Condition, ctx *ModelBuilderCtx) xorm.Cond {
	if cond.GetLiteralCond() != nil && cond.GetLiteralCond().GetLiteral() != "" {
		// 由于 xorm.Expr() 源码内会自动对 sql片段加括号, 因此字面量不需要主动再次添加括号
		return xorm.Expr(cond.GetLiteralCond().GetLiteral())
	}
	operator := cond.Operator
	condField := extraConditionOpField(cond, ctx)
	switch operator {
	//column op arg[0]
	case pb.Op_OP_EQ:
		return &xorm.Eq{condField: ExtraArgItemValue(cond.Args[0])}
	case pb.Op_OP_NEQ:
		return &xorm.Neq{condField: ExtraArgItemValue(cond.Args[0])}
	case pb.Op_OP_LT:
		return &xorm.Lt{condField: ExtraArgItemValue(cond.Args[0])}
	case pb.Op_OP_LTE:
		return xorm.Lte{condField: ExtraArgItemValue(cond.Args[0])}
	case pb.Op_OP_GT:
		return xorm.Gt{condField: ExtraArgItemValue(cond.Args[0])}
	case pb.Op_OP_GTE:
		return xorm.Gte{condField: ExtraArgItemValue(cond.Args[0])}
	case pb.Op_OP_LIKE:
		return xorm.Like{condField, ExtraArgItemValue(cond.Args[0]).(string)}
		// 不包含, 以XX开头, 以XX结尾, 调用方在ARG上处理传过来
	case pb.Op_OP_NOT_LIKE:
		// reverse=true
		return xorm.Like{condField, ExtraArgItemValue(cond.Args[0]).(string)}
	case pb.Op_OP_PREFIX_LIKE:
		return xorm.Like{condField, ExtraArgItemValue(cond.Args[0]).(string)}
	case pb.Op_OP_LIKE_SUFFIX:
		return xorm.Like{condField, ExtraArgItemValue(cond.Args[0]).(string)}
		// column op ...arg
	case pb.Op_OP_BETWEEN:
		return xorm.Between{Col: condField, LessVal: ExtraArgItemValue(cond.Args[0]), MoreVal: ExtraArgItemValue(cond.Args[1])}
	case pb.Op_OP_IN:
		inArgs := extraConditionArgsSlice(cond)
		return xorm.In(condField, inArgs)
	case pb.Op_OP_NOT_IN:
		inArgs := extraConditionArgsSlice(cond)
		return xorm.NotIn(condField, inArgs)
		// column op
	case pb.Op_OP_IS_NULL:
		return xorm.IsNull{condField}
	case pb.Op_OP_IS_NOT_NULL:
		return xorm.NotNull{condField}
		// Where(xorm.Expr(Column, ...arg)), and 和 or 暂时使用Expr处理, 需要sqlbuilder调用方提供完整 sqlFormat;; => 未测试
	case pb.Op_OP_AND:
		formatArgs := extraConditionArgsSlice(cond)
		return xorm.And(xorm.Expr(condField, formatArgs))
	case pb.Op_OP_OR:
		formatArgs := extraConditionArgsSlice(cond)
		return xorm.Or(xorm.Expr(condField, formatArgs))
	}
	return nil
}

func op2ConditionString(operator pb.Op) string {
	switch operator {
	case pb.Op_OP_EQ:
		return string(clause.ConditionEq)
	case pb.Op_OP_NEQ:
		return string(clause.ConditionNeq)
	case pb.Op_OP_LT:
		return string(clause.ConditionLt)
	case pb.Op_OP_LTE:
		return string(clause.ConditionLte)
	case pb.Op_OP_GT:
		return string(clause.ConditionGt)
	case pb.Op_OP_GTE:
		return string(clause.ConditionGte)
	}
	return string(clause.ConditionEq)
}

func logic2String(operator pb.Logic) string {
	switch operator {
	case pb.Logic_LOGIC_AND:
		return string(clause.ConditionAnd)
	case pb.Logic_LOGIC_OR:
		return string(clause.ConditionOr)
	}
	return string(clause.ConditionAnd)
}

// sql 建模中 MixField 处理
func extraConditionOpField(cond *pb.Condition, ctx *ModelBuilderCtx) string {
	return processMixFieldItemByProto(cond.GetField(), ctx)
}

func ExtraArgItemValue(item *pb.BasicData) interface{} {
	switch elem := item.GetData().(type) {
	case *pb.BasicData_StrVal:
		return elem.StrVal
	case *pb.BasicData_StrLiteral:
		return elem.StrLiteral.Literal
	case *pb.BasicData_IntVal:
		return elem.IntVal
	case *pb.BasicData_DoubleVal:
		return elem.DoubleVal
	case *pb.BasicData_Logic:
		return elem.Logic
	}
	return ""
}

func extraConditionArgsSlice(cond *pb.Condition) interface{} {
	var inArgs []interface{}
	for _, data := range cond.GetArgs() {
		switch d := data.GetData().(type) {
		case *pb.BasicData_IntVal:
			inArgs = append(inArgs, d.IntVal)
		case *pb.BasicData_DoubleVal:
			inArgs = append(inArgs, d.DoubleVal)
		case *pb.BasicData_StrVal:
			inArgs = append(inArgs, d.StrVal)
		case *pb.BasicData_Logic:
			inArgs = append(inArgs, d.Logic)
		}
	}
	return inArgs
}

// 构建 sql.order-by
func buildOrderByProto(builder *xorm.Builder, sqlRef *pb.SqlReference, ctx *ModelBuilderCtx) *xorm.Builder {
	orderByClause := sqlRef.OrderBy
	// 提取 orderBy 字符串片段
	var orderBys []string
	for _, orderBy := range orderByClause {
		field := processMixFieldItemByProto(orderBy.Dependent, ctx)
		orderByItem := fmt.Sprintf(clause.DoubleStringLiteral, field, orderBy.Order)
		orderBys = append(orderBys, orderByItem)
	}
	// 合并成一个字符串片段
	orderByItems := strings.Join(orderBys, clause.COMMA)
	// xorm 的 orderBy 方法只接收一个字符串参数，因此要自己将多个 orderBy 拼接为1个
	orderByFragment := fmt.Sprintf(clause.StringLiteral, orderByItems)

	return builder.OrderBy(orderByFragment)
}

// 构建 sql.limit
func buildLimitProto(builder *xorm.Builder, sqlRef *pb.SqlReference) *xorm.Builder {
	limit := sqlRef.Limit
	// 分页大小为0数据库允许但代码视为无效分页
	if limit.LimitN == 0 {
		return builder
	}
	return builder.Limit(int(limit.LimitN), int(limit.Offset))
}

// 构建 sql.group-by
func buildGroupByByProto(builder *xorm.Builder, sqlRef *pb.SqlReference, ctx *ModelBuilderCtx) *xorm.Builder {
	groupByClause := sqlRef.GroupBy
	// 类型断言: Column 或者 Expression
	groupByItems := extraColumnOrExpressionStringsByProto(groupByClause, ctx)

	// 使用 strings.Join 来合并切片元素，以逗号分隔
	joinedItems := strings.Join(groupByItems, clause.COMMA)

	// xorm 的 GroupBy 方法只接收一个字符串参数，因此要自己将多个 GroupBy 拼接为1个
	groupByFragment := fmt.Sprintf(clause.StringLiteral, joinedItems)

	return builder.GroupBy(groupByFragment)
}
