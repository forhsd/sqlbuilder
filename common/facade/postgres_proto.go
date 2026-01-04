package facade

import (
	"github.com/forhsd/logger"
	xorm "xorm.io/builder"
)

// PostgresModelBuilderSql 功能接口具体实现
type PostgresModelBuilderSql struct {
	DefaultChainBuilder
}

// 类型实现检查
var _ ModelChainSqlBuilder = &PostgresModelBuilderSql{}

// BuildDialect pg 对于接口实现
// 参数:
//
// 返回值:
//   - 返回一个 xorm.Builder 对象
//   - 异常
func (p *PostgresModelBuilderSql) BuildDialect(ctx *ModelBuilderCtx) (*xorm.Builder, error) {
	logger.Debug("[模板方法模式]-[postgres] 构建方言...")
	return xorm.Dialect(xorm.POSTGRES), nil
}
