package service

import (
	"errors"

	"github.com/forhsd/sqlbuilder/common"
	"github.com/forhsd/sqlbuilder/common/clause"
	"github.com/forhsd/sqlbuilder/common/facade"
	pb "github.com/forhsd/sqlbuilder/gen/proto"
)

// BuildSqlByJson http协议
func BuildSqlByJson(builder clause.BuilderRequest) (string, error) {
	mode := builder.Strategy
	// service 根据构建策略决定具体实现, sql生成根据driver方言来生成
	switch mode {
	case clause.ModelStrategy:
		sql, err := common.BuildModelSqlByJson(builder)
		return sql, err
	case clause.TemplateStrategy:
		sql, err := common.BuildTemplateSqlByJson(builder)
		return sql, err
	default:
		return "", errors.New("未知SQL构建策略")
	}
}

// BuildSqlByProto proto 协议
func BuildSqlByProto(req *pb.BuilderRequest) (string, error) {
	mode := req.Strategy
	// service 根据构建策略决定具体实现, sql生成根据driver方言来生成
	switch mode {
	case pb.BuilderStrategy_BUILDER_STRATEGY_MODEL:
		sql, err := common.BuildSqlByModelProto(req)
		return sql, err
	case pb.BuilderStrategy_BUILDER_STRATEGY_TEMPLATE:
		sql, err := common.BuildSqlByTemplateProto(req)
		return sql, err
	default:
		return "", errors.New("未知SQL构建策略")
	}
}

// AnalyzeTemplatesByJson : service 层, 根据http协议解析模板字符串, 提取子模板和主模板
// 参数:
//   - req: 解析请求体
//
// 返回值:
//   - 返回一个 模板上下文对象
//   - 返回一个 异常
func AnalyzeTemplatesByJson(req facade.AnalyzeTemplateRequest) (*facade.TemplateCtx, error) {
	return common.AnalyzeTemplatesByJson(req)
}

// AnalyzeTemplatesByProto : service 层, 根据proto协议解析模板字符串, 提取子模板和主模板
// 参数:
//   - req: 解析请求体
//
// 返回值:
//   - 返回一个 模板上下文对象
//   - 返回一个 异常
func AnalyzeTemplatesByProto(req *pb.AnalyzeTemplateRequest) (*facade.TemplateCtx, error) {
	return common.AnalyzeTemplatesByProto(req)
}

func AnalyzeAdditionByProto(req *pb.AnalyzeAdditionRequest) (*facade.NativeSqlHeader, error) {
	return common.AnalyzeAdditionByProto(req)
}
