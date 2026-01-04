package facade

import (
	"errors"
	"regexp"

	"github.com/forhsd/sqlbuilder/common/clause"
	pb "github.com/forhsd/sqlbuilder/gen/proto"
	xorm "xorm.io/builder"
)

// -----------------------------------------------   通过模型构建sql 公共常量  --------------------------------------------

var (
	String1LiteralSafeMap = map[pb.Driver]string{
		pb.Driver_DRIVER_POSTGRES: clause.PGStringLiteralSafe,
		pb.Driver_DRIVER_MYSQL:    clause.MysqlStringLiteralSafe,
		pb.Driver_DRIVER_DORIS:    clause.MysqlStringLiteralSafe,
	}

	String2LiteralSafeMap = map[pb.Driver]string{
		pb.Driver_DRIVER_POSTGRES: clause.PGTableColumn,
		pb.Driver_DRIVER_MYSQL:    clause.MysqlTableTableColumn,
		pb.Driver_DRIVER_DORIS:    clause.MysqlTableTableColumn,
	}

	String2LiteralSafeAsMap = map[pb.Driver]string{
		pb.Driver_DRIVER_POSTGRES: clause.PGTableColumnAs,
		pb.Driver_DRIVER_MYSQL:    clause.MysqlTableColumnAs,
		pb.Driver_DRIVER_DORIS:    clause.MysqlTableColumnAs,
	}

	AsFormatLiteralSafeMap = map[pb.Driver]string{
		pb.Driver_DRIVER_POSTGRES: clause.PGAsFormat,
		pb.Driver_DRIVER_MYSQL:    clause.MYSQLAsFormat,
		pb.Driver_DRIVER_DORIS:    clause.MYSQLAsFormat,
	}

	TableLiteralAsFormatLiteralSafeMap = map[pb.Driver]string{
		pb.Driver_DRIVER_POSTGRES: clause.PGLiteralTableAsFormat,
		pb.Driver_DRIVER_MYSQL:    clause.MYSQLLiteralTableAsFormat,
		pb.Driver_DRIVER_DORIS:    clause.MYSQLLiteralTableAsFormat,
	}

	DriverMap = map[pb.Driver]clause.Driver{
		pb.Driver_DRIVER_POSTGRES: clause.DriverPostgres,
		pb.Driver_DRIVER_MYSQL:    clause.DriverMysql,
		pb.Driver_DRIVER_DORIS:    clause.DriverDoris,
	}

	DialectDriverMap = map[clause.Driver]string{
		clause.DriverPostgres: xorm.POSTGRES,
		clause.DriverMysql:    xorm.MYSQL,
		clause.DriverDoris:    xorm.MYSQL,
	}

	ChainBuilderInst = map[pb.Driver]ModelChainSqlBuilder{
		pb.Driver_DRIVER_POSTGRES: &PostgresModelBuilderSql{},
		pb.Driver_DRIVER_MYSQL:    &MysqlModelBuilderSql{},
		pb.Driver_DRIVER_DORIS:    &MysqlModelBuilderSql{},
	}
)

type ModelBuilderCtx struct {
	String1LiteralSafeFormat              string
	String2LiteralSafeFormat              string
	String2LiteralAsSafeFormat            string
	AsFormatLiteralSafeFormat             string
	TableLiteralAsFormatLiteralSafeFormat string
	Driver                                clause.Driver
}

// -----------------------------------------------   代码结构对比 java 如下  --------------------------------------------

//// 功能(或函数式)接口
//public interface BuilderSql<T> {
//	String foo1(T request);
//	default String foo2(String step1) { return "Common foo2"; }
//}
//// 抽象门户入口
//public abstract class AbstractFacade {
//	protected final BuilderSql builderSql;
//	public AbstractFacade(BuilderSql builderSql) {
//	this.builderSql = builderSql;
//	}
//
//	public <T> void execute(T request) {
//		// 模板方法
//		String step1 = builderSql.foo1(request);
//		String step2 = builderSql.foo2(step1);
//		String stepFinal = this.childFoo(step2);
//	}
//	protected abstract String childFoo(String step) {}
//}
//// 门面实例
//public class FacadeInstance extends AbstractFacade {
//	public ModelBuilderFacade(BuilderSql builderSql) {
//		super(builderSql);
//	}
//}
////使用
//BuilderSql builderInst = BuilderSqlFactory.createBuilder("driver");
//AbstractFacade facade = new FacadeInstance(builderInst);
//facade.execute(request);

// -----------------------------------------------   facade: 通过模型构建sql  --------------------------------------------

// ModelChainBuilderFactory 工厂方法根据Driver选择具体Builder
// 参数:
//   - driver: 驱动
//
// 返回值:
//   - 模型模式SqlBuilder门面接口
//   - 异常
func ModelChainBuilderFactory(driver pb.Driver) (ModelChainSqlBuilder, error) {
	//switch driver {
	//case pb.Driver_DRIVER_POSTGRES:
	//	return &PostgresModelBuilderSql{}, nil
	//case pb.Driver_DRIVER_MYSQL, pb.Driver_DRIVER_DORIS:
	//	return &MysqlModelBuilderSql{}, nil
	//default:
	//	return nil, errors.New("driver not support")
	//}

	// 使用 map 取代 switch，提高可读性，避免膨胀
	if builder, exists := ChainBuilderInst[driver]; exists {
		return builder, nil
	}
	return nil, errors.New("driver not support")
}

// ModelChainSqlBuilder 定义接口(模型模式SqlBuilder接口)
// 参数:
//   - request: proto 请求对象
//
// 返回值:
//   - 返回 sql
//   - 异常
type ModelChainSqlBuilder interface {
	BuildDialect(ctx *ModelBuilderCtx) (*xorm.Builder, error)
	BuildChainSql(req *pb.BuilderRequest, appendBuilder *xorm.Builder, dialectBuilder *xorm.Builder, sqlRef *pb.SqlReference, ctx *ModelBuilderCtx) *xorm.Builder
}

type DefaultChainBuilder struct{}

func (c *DefaultChainBuilder) BuildChainSql(req *pb.BuilderRequest,
	appendBuilder *xorm.Builder, dialectBuilder *xorm.Builder,
	sqlRef *pb.SqlReference, ctx *ModelBuilderCtx) *xorm.Builder {
	return ChainSqlBuilder(req, appendBuilder, dialectBuilder, sqlRef, ctx)
}

// AbstractModelBuilderFacade 定义抽象Facade，提供模板方法模式的基础实现
// Go 语言不支持 继承，而是推荐 组合 方式
// 模板方法步骤:
// 1. BuildDialect, 创建方言
// 2. BuildBasic, 链式创建 select & from
// 3. BuildOther, 链式创建 SQL 其他关键字
type AbstractModelBuilderFacade struct {
	builderSql ModelChainSqlBuilder
}

// NewModelChainSqlBuilderFacadeInstance 构造函数
func NewModelChainSqlBuilderFacadeInstance(builderSql ModelChainSqlBuilder) *AbstractModelBuilderFacade {
	return &AbstractModelBuilderFacade{builderSql: builderSql}
}

// Execute 门户入口
// 参数:
//   - request: proto 请求
//
// 返回值:
//   - 返回 真实可执行的SQL
//   - 异常
func (abs *AbstractModelBuilderFacade) Execute(req *pb.BuilderRequest) (string, error) {
	builders := req.Builders
	ctx := abs.BuildCtx(req)
	var appendBuilder *xorm.Builder

	// 循环 append *xorm.Builder 内容
	for _, ref := range builders {
		xormBuilder, err := abs.processChainBuilder(req, appendBuilder, ref, ctx)
		if err != nil {
			return "", err
		}
		appendBuilder = xormBuilder
	}

	if appendBuilder == nil {
		return "", errors.New("nest appendBuilder is nil")
	}

	// builder -> sql
	boundSQL, err := appendBuilder.ToBoundSQL()
	if err != nil {
		return "", err
	}

	// 替换掉 CROSS JOIN 的 ON /*__XORM_CROSS_JOIN__*/
	re := regexp.MustCompile(clause.CrossJoinPlaceholderRegex)
	boundSQL = re.ReplaceAllString(boundSQL, "")

	return boundSQL, nil
}

// BuildCtx 构建单层上下文
// 参数:
//   - request: proto 请求
//
// 返回值:
//   - 返回 *ModelBuilderCtx
func (abs *AbstractModelBuilderFacade) BuildCtx(req *pb.BuilderRequest) *ModelBuilderCtx {
	return &ModelBuilderCtx{
		String1LiteralSafeFormat:              String1LiteralSafeMap[req.Driver],
		String2LiteralSafeFormat:              String2LiteralSafeMap[req.Driver],
		String2LiteralAsSafeFormat:            String2LiteralSafeAsMap[req.Driver],
		AsFormatLiteralSafeFormat:             AsFormatLiteralSafeMap[req.Driver],
		TableLiteralAsFormatLiteralSafeFormat: TableLiteralAsFormatLiteralSafeMap[req.Driver],
		Driver:                                DriverMap[req.Driver],
	}
}

// processChainBuilder 构建单层完整构建器
// 参数:
//   - appendBuilder: proto 请求对象
//   - builderRef: proto 请求中当前层级sqlRef对象
//
// 返回值:
//   - 返回 *xorm.Builder
//   - 异常
func (abs *AbstractModelBuilderFacade) processChainBuilder(
	req *pb.BuilderRequest, appendBuilder *xorm.Builder,
	builderRef *pb.DeepWrapper, ctx *ModelBuilderCtx) (*xorm.Builder, error) {
	// 上游确定是 model 构造器
	sqlRef := builderRef.Sql.GetModel()
	// 选择方言
	dialectBuilder, err := abs.builderSql.BuildDialect(ctx)
	if err != nil {
		return nil, err
	}

	// 链式构造器
	builder := ChainSqlBuilder(req, appendBuilder, dialectBuilder, sqlRef, ctx)

	return builder, nil
}

// -----------------------------------------------   facade: 通过模板构建sql  --------------------------------------------

// BuildSqlByTemplate : 根据模板语法构建sql
// 参数:
//   - request: sqlbuilder 请求体
//
// 返回值:
//   - 返回 sql
//   - 返回 错误信息
func BuildSqlByTemplate(request *pb.BuilderRequest) (string, error) {
	return RenderMasterTemplate(request.GetBuilders()[0].GetSql().GetTemplate())
}
