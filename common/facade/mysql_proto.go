package facade

import (
	"github.com/forhsd/logger"
	xorm "xorm.io/builder"
)

// MysqlModelBuilderSql 功能接口具体实现
type MysqlModelBuilderSql struct {
	DefaultChainBuilder
}

// 类型实现检查
var _ ModelChainSqlBuilder = &MysqlModelBuilderSql{}

// BuildDialect mysql 对于接口实现
// 参数:
//
// 返回值:
//   - 返回一个 xorm.Builder 对象
//   - 异常
func (m *MysqlModelBuilderSql) BuildDialect(ctx *ModelBuilderCtx) (*xorm.Builder, error) {
	logger.Debug("[模板方法模式]-[mysql] 构建方言...")
	// xorm.Dialect(DialectDriverMap[ctx.Driver]), nil
	return xorm.Dialect(xorm.MYSQL), nil
}
