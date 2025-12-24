package proto

import (
	"fmt"
	"testing"

	"github.com/forhsd/sqlbuilder/gen/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

var protoJsonOP = protojson.MarshalOptions{
	EmitUnpopulated: true,
	// UseEnumNumbers:  true,
}

func makeSelect1Vars() []*proto.MixVars {
	vars := make([]*proto.MixVars, 0)
	col := &proto.MixVars_Column{
		Column: &proto.Column{Schema: "sample_data", Table: "products", Field: "price", Alias: "合计", AggAble: false, UseAs: false},
	}
	vars = append(vars, &proto.MixVars{Vars: col})

	return vars
}

func makeSelect2Vars() []*proto.MixVars {
	vars := make([]*proto.MixVars, 0)
	col := &proto.MixVars_Column{
		Column: &proto.Column{Schema: "sample_data", Table: "products", Field: "created_at", Alias: "created_at表达式内部as", AggAble: false, UseAs: false},
	}
	colStr := &proto.MixVars_Context{
		Context: "month",
	}
	vars = append(vars, &proto.MixVars{Vars: colStr}, &proto.MixVars{Vars: col})

	return vars
}

func makeSelect() []*proto.MixField {
	selects := make([]*proto.MixField, 0)
	selectCol1 := &proto.MixField_Column{
		Column: &proto.Column{Schema: "sample_data", Table: "products", Field: "category", Alias: "分类", AggAble: false, UseAs: true},
	}
	selectExpr1 := &proto.MixField_Expression{
		Expression: &proto.Expression{
			Call:     "sum",
			CallType: proto.CallType_CALL_TYPE_AGG,
			CallAs:   "合计价格",
			UseAs:    true,
			Vars:     makeSelect1Vars(),
		},
	}

	selectExpr2 := &proto.MixField_Expression{
		Expression: &proto.Expression{
			Call:     "DATE_TRUNC",
			CallType: proto.CallType_CALL_TYPE_INNER,
			CallAs:   "离散日期(月)",
			UseAs:    true,
			Vars:     makeSelect2Vars(),
		},
	}

	selects = append(selects, &proto.MixField{Mix: selectCol1}, &proto.MixField{Mix: selectExpr1}, &proto.MixField{Mix: selectExpr2})

	return selects
}

func makeCaseWhenWhere1() []*proto.MixWhere {
	wheres := make([]*proto.MixWhere, 0)

	// 0 < x <= 3
	condition1 := makeCaseWhenWhereCondition(proto.Op_OP_GTE, 0.0)
	condition2 := makeCaseWhenWhereCondition(proto.Op_OP_LTE, 3.0)
	wheres = append(wheres, &proto.MixWhere{Filter: condition1}, &proto.MixWhere{Filter: condition2})

	return wheres
}

func makeCaseWhenWhere2() []*proto.MixWhere {
	wheres := make([]*proto.MixWhere, 0)

	// 3.0 < x <= 5.0
	condition1 := makeCaseWhenWhereCondition(proto.Op_OP_GT, 3.0)
	condition2 := makeCaseWhenWhereCondition(proto.Op_OP_LTE, 5.0)
	wheres = append(wheres, &proto.MixWhere{Filter: condition1}, &proto.MixWhere{Filter: condition2})

	return wheres
}

func makeCaseWhenWhereCondition(op proto.Op, ratingDouble float64) *proto.MixWhere_Condition {
	conditionCol1 := &proto.MixField_Column{
		Column: &proto.Column{Schema: "sample_data", Table: "products", Field: "rating", Alias: "比率", AggAble: false, UseAs: false},
	}

	args := make([]*proto.BasicData, 0)
	args = append(args, &proto.BasicData{Data: &proto.BasicData_DoubleVal{DoubleVal: ratingDouble}})

	condition1 := &proto.MixWhere_Condition{
		Condition: &proto.Condition{Field: &proto.MixField{Mix: conditionCol1}, Operator: op, Args: args, Logic: proto.Logic_LOGIC_AND},
	}
	return condition1
}

func makeCaseWhenSelect() []*proto.MixField {
	selects := make([]*proto.MixField, 0)

	productCaseWhen1 := &proto.MixField_CaseWhen{
		CaseWhen: &proto.CaseWhen{
			Alias:     "比率区间",
			ElseValue: &proto.BasicData{Data: &proto.BasicData_StrVal{StrVal: ""}},
			Conditions: []*proto.CaseWhenItem{
				{
					Then: &proto.BasicData{Data: &proto.BasicData_StrVal{StrVal: "不合格"}},
					When: makeCaseWhenWhere1(),
				},
				{
					Then: &proto.BasicData{Data: &proto.BasicData_StrVal{StrVal: "合格"}},
					When: makeCaseWhenWhere2(),
				},
			},
		},
	}

	selects = append(selects, &proto.MixField{Mix: productCaseWhen1})

	return selects
}

func makeGroupByVars() []*proto.MixVars {
	vars := make([]*proto.MixVars, 0)
	col := &proto.MixVars_Column{
		Column: &proto.Column{Schema: "sample_data", Table: "products", Field: "created_at", Alias: "created_at表达式内部as", AggAble: false, UseAs: false},
	}
	colStr := &proto.MixVars_Context{
		Context: "month",
	}
	vars = append(vars, &proto.MixVars{Vars: colStr}, &proto.MixVars{Vars: col})

	return vars
}

func makeGroupBy() []*proto.MixField {
	selects := make([]*proto.MixField, 0)
	groupByCol1 := &proto.MixField_Column{
		Column: &proto.Column{Schema: "sample_data", Table: "products", Field: "category", Alias: "分类", AggAble: false, UseAs: false},
	}

	groupByExpr2 := &proto.MixField_Expression{
		Expression: &proto.Expression{
			Call:     "DATE_TRUNC",
			CallType: proto.CallType_CALL_TYPE_INNER,
			CallAs:   "离散日期(月)",
			UseAs:    false,
			Vars:     makeGroupByVars(),
		},
	}

	selects = append(selects, &proto.MixField{Mix: groupByExpr2}, &proto.MixField{Mix: groupByCol1})

	return selects
}

func makeWhere() []*proto.MixWhere {
	wheres := make([]*proto.MixWhere, 0)

	condition1 := makeWhereCondition1Eq()
	condition2 := makeWhereCondition2Gt()
	condition3 := makeWhereCondition3In()
	condition4 := makeWhereCondition4Between()

	wheres = append(wheres, &proto.MixWhere{Filter: condition1}, &proto.MixWhere{Filter: condition2}, &proto.MixWhere{Filter: condition3}, &proto.MixWhere{Filter: condition4})

	return wheres
}

func makeWhereCondition1Eq() *proto.MixWhere_Condition {
	conditionCol1 := &proto.MixField_Column{
		Column: &proto.Column{Schema: "sample_data", Table: "products", Field: "category", Alias: "分类", AggAble: false, UseAs: false},
	}

	args := make([]*proto.BasicData, 0)
	args = append(args, &proto.BasicData{Data: &proto.BasicData_StrVal{StrVal: "Doohickey"}})

	condition1 := &proto.MixWhere_Condition{
		Condition: &proto.Condition{Field: &proto.MixField{Mix: conditionCol1}, Operator: proto.Op_OP_EQ, Args: args, Logic: proto.Logic_LOGIC_AND},
	}
	return condition1
}

func makeWhereCondition2Gt() *proto.MixWhere_Condition {
	conditionCol1 := &proto.MixField_Column{
		Column: &proto.Column{Schema: "sample_data", Table: "products", Field: "price", Alias: "价格", AggAble: false, UseAs: false},
	}

	args := make([]*proto.BasicData, 0)
	args = append(args, &proto.BasicData{Data: &proto.BasicData_IntVal{IntVal: 3}})

	condition1 := &proto.MixWhere_Condition{
		Condition: &proto.Condition{Field: &proto.MixField{Mix: conditionCol1}, Operator: proto.Op_OP_GT, Args: args, Logic: proto.Logic_LOGIC_AND},
	}
	return condition1
}

func makeWhereCondition3In() *proto.MixWhere_Condition {
	conditionCol1 := &proto.MixField_Column{
		Column: &proto.Column{Schema: "sample_data", Table: "products", Field: "id", Alias: "主键", AggAble: false, UseAs: false},
	}

	args := make([]*proto.BasicData, 0)
	args = append(args, &proto.BasicData{Data: &proto.BasicData_IntVal{IntVal: 3}})
	args = append(args, &proto.BasicData{Data: &proto.BasicData_IntVal{IntVal: 4}})
	args = append(args, &proto.BasicData{Data: &proto.BasicData_IntVal{IntVal: 10}})

	condition1 := &proto.MixWhere_Condition{
		Condition: &proto.Condition{Field: &proto.MixField{Mix: conditionCol1}, Operator: proto.Op_OP_IN, Args: args, Logic: proto.Logic_LOGIC_AND},
	}
	return condition1
}

func makeWhereCondition4Between() *proto.MixWhere_Condition {
	conditionCol1 := &proto.MixField_Column{
		Column: &proto.Column{Schema: "sample_data", Table: "products", Field: "created_at", Alias: "主键", AggAble: false, UseAs: false},
	}

	args := make([]*proto.BasicData, 0)
	args = append(args, &proto.BasicData{Data: &proto.BasicData_StrVal{StrVal: "2017-07-19 00:00:00.000"}})
	args = append(args, &proto.BasicData{Data: &proto.BasicData_StrVal{StrVal: "2017-09-14 00:00:00.000"}})

	condition1 := &proto.MixWhere_Condition{
		Condition: &proto.Condition{Field: &proto.MixField{Mix: conditionCol1}, Operator: proto.Op_OP_BETWEEN, Args: args, Logic: proto.Logic_LOGIC_AND},
	}
	return condition1
}

func TestBuilderRequest_Descriptor(t *testing.T) {
	sql := &proto.SqlReference{
		// From:        &proto.Table{TableName: "products", TableSchema: "sample_data", TableAlias: "products"},
		From: &proto.MixTable{
			Mt: &proto.MixTable_NormalTable{
				NormalTable: &proto.Table{
					TableName: "products", TableSchema: "sample_data", TableAlias: "products",
				},
			},
		},
		Join:        []*proto.Join{},
		Where:       makeWhere(),
		GroupBy:     makeGroupBy(),
		Aggregation: []*proto.Expression{},
		Select:      makeSelect(),
		OrderBy:     []*proto.OrderBy{},
		Limit:       &proto.Limit{},
	}

	modelSql := &proto.MixSql_Model{Model: sql}

	deep0 := &proto.DeepWrapper{Deep: 0, Sql: &proto.MixSql{Ref: modelSql}}

	request := &proto.BuilderRequest{
		Driver:   proto.Driver_DRIVER_POSTGRES,
		Strategy: proto.BuilderStrategy_BUILDER_STRATEGY_MODEL,
		Builders: []*proto.DeepWrapper{deep0},
	}

	buf, err := protoJsonOP.Marshal(request)

	if err != nil {
		fmt.Println(err)
	}

	json := string(buf)
	fmt.Printf(json)

	t.Run("测试", func(t *testing.T) {
		if len(request.Builders[0].Sql.GetModel().GroupBy) != 2 && len(request.Builders[0].Sql.GetModel().Select) != 3 {
			t.Errorf("断言失败")
		}
	})
}

func TestCaseWhen_Descriptor(t *testing.T) {
	sql := &proto.SqlReference{
		// From:        &proto.Table{TableName: "products", TableSchema: "sample_data", TableAlias: "products"},
		From: &proto.MixTable{
			Mt: &proto.MixTable_NormalTable{
				NormalTable: &proto.Table{
					TableName: "products", TableSchema: "sample_data", TableAlias: "products",
				},
			},
		},
		Join:        []*proto.Join{},
		Where:       []*proto.MixWhere{},
		GroupBy:     []*proto.MixField{},
		Aggregation: []*proto.Expression{},
		Select:      makeCaseWhenSelect(),
		OrderBy:     []*proto.OrderBy{},
		Limit:       &proto.Limit{},
	}

	modelSql := &proto.MixSql_Model{Model: sql}

	deep0 := &proto.DeepWrapper{Deep: 0, Sql: &proto.MixSql{Ref: modelSql}}

	request := &proto.BuilderRequest{
		Driver:   proto.Driver_DRIVER_POSTGRES,
		Strategy: proto.BuilderStrategy_BUILDER_STRATEGY_MODEL,
		Builders: []*proto.DeepWrapper{deep0},
	}

	buf, err := protoJsonOP.Marshal(request)

	if err != nil {
		fmt.Println(err)
	}

	json := string(buf)
	fmt.Printf(json)
	fmt.Println()

	t.Run("测试", func(t *testing.T) {
		if len(request.Builders[0].Sql.GetModel().Select) != 1 {
			t.Errorf("断言失败")
		}
	})
}
