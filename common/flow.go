package common

import (
	"github.com/forhsd/logger"
	"github.com/forhsd/sqlbuilder/common/clause"
	"github.com/forhsd/sqlbuilder/common/facade"
	pb "github.com/forhsd/sqlbuilder/gen/proto"
)

// ==================================================================================================================
// ==================================================== model-chain ====================================================
// ==================================================================================================================

// todo 1. java Object传参顺序问题影响SQL生成和结果 2. schema.table.column 参考MB 3.format在占位符写死了,根据方言决定 "" 或``

func BuildModelSqlByJson(request clause.BuilderRequest) (string, error) {
	builder, bErr := facade.RunBuilder(request)

	defer func() {
		if err := recover(); bErr != nil {
			logger.Error("构建SQL失败, 原因: %v", err)
		}
	}()

	boundSQL, sqlErr := builder.ToBoundSQL()
	if sqlErr != nil {
		logger.Error("转换SQL失败, 原因: %v", sqlErr)
	}

	logger.Debug("执行模型模式SQL构建策略, pg方言SQL: \n %s", boundSQL)

	return boundSQL, nil
}

func BuildTemplateSqlByJson(request clause.BuilderRequest) (string, error) {
	logger.Debug("执行模板模式SQL构建策略, pg方言SQL: %s", "SELECT 模板")
	return "", nil
}

// BuildSqlByModelProto 使用 proto 解决 http 传输联合类型接收为 map 的元组类型擦除问题
func BuildSqlByModelProto(req *pb.BuilderRequest) (string, error) {
	// 根据方言创建facade
	builderInst, err := facade.ModelChainBuilderFactory(req.Driver)
	if err != nil {
		return "", err
	}
	facadeInstance := facade.NewModelChainSqlBuilderFacadeInstance(builderInst)
	// 构建 xormBuilder -> sql
	boundSQL, err := facadeInstance.Execute(req)
	if err != nil {
		return "", err
	}
	logger.Debug("执行model模式SQL构建策略, SQL:\n%s", boundSQL)

	return boundSQL, nil
}

// ==================================================================================================================
// ==================================================== template ====================================================
// ==================================================================================================================

// BuildSqlByTemplateProto 根据proto协议请求和模板语法构建sql门面接口的实现
// 参数:
//   - request: proto 请求对象
//
// 返回值:
//   - 返回 sql
//   - 异常
func BuildSqlByTemplateProto(request *pb.BuilderRequest) (string, error) {
	// 对于模板语法构建SQL策略来说, 不存在方言一说, SQL语法方言由写SQL的人控制, 因此直接调用模板引擎即可
	sql, err := facade.BuildSqlByTemplate(request)
	logger.Debug("执行template模式SQL构建策略, SQL:\n%s", sql)
	return sql, err
}

// AnalyzeTemplatesByJson : flow 层, 根据http协议解析模板字符串, 提取子模板和主模板
// 参数:
//   - req: 解析请求体
//
// 返回值:
//   - 返回一个 模板上下文对象
//   - 返回一个 异常
func AnalyzeTemplatesByJson(req facade.AnalyzeTemplateRequest) (*facade.TemplateCtx, error) {
	return facade.AnalyzeTemplatesByJson(req)
}

// AnalyzeTemplatesByProto : flow 层, 根据proto协议解析模板字符串, 提取子模板和主模板
// 参数:
//   - req: 解析请求体
//
// 返回值:
//   - 返回一个 模板上下文对象
//   - 返回一个 异常
func AnalyzeTemplatesByProto(req *pb.AnalyzeTemplateRequest) (*facade.TemplateCtx, error) {
	return facade.AnalyzeTemplatesByProto(req)
}

func AnalyzeAdditionByProto(req *pb.AnalyzeAdditionRequest) (*facade.NativeSqlHeader, error) {
	return facade.AnalyzeAdditionByProto(req)
}
