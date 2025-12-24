package clause

import (
	"bytes"
	"fmt"
	"strings"
)

type SimpleUpdate struct {
	Driver Driver
	Schema string
	Table  string
	Sets   map[string]interface{}
	Where  map[string]interface{}
}

var (
	string2LiteralSafeMapOnJson = map[Driver]string{
		DriverPostgres: PGTableColumn,
		DriverMysql:    MysqlTableTableColumn,
		DriverDoris:    MysqlTableTableColumn,
	}
)

const (
	leftEqRightFormat = " %s = %s "
)

func (info *SimpleUpdate) BuildSimpleSql() string {
	strBuf := bytes.NewBufferString("")

	// 关键字1
	strBuf.WriteString("update")

	str2Format := string2LiteralSafeMapOnJson[info.Driver]
	strBuf.WriteString(fmt.Sprintf(str2Format, info.Schema, info.Table))

	// 关键字2
	strBuf.WriteString(" set ")

	// 构建 set
	var tempSetArr []string
	for field, value := range info.Sets {
		left := updateSetByDriver(info, field, str2Format)
		tempSetArr = append(tempSetArr, fmt.Sprintf(leftEqRightFormat, left, fmt.Sprintf("%v", value)))
	}

	finalSetStr := strings.Join(tempSetArr, ",")
	strBuf.WriteString(finalSetStr)

	// 关键字3
	strBuf.WriteString(" where ")

	// 构建 where
	var tempWhereArr []string
	for field, value := range info.Where {
		left := fmt.Sprintf(str2Format, info.Table, fmt.Sprintf("%v", field))
		tempWhereArr = append(tempWhereArr, fmt.Sprintf(leftEqRightFormat, left, fmt.Sprintf("%v", value)))
	}

	finalWhereStr := strings.Join(tempWhereArr, " and ")
	strBuf.WriteString(finalWhereStr)
	strBuf.WriteString(";")

	return strBuf.String()
}

func updateSetByDriver(info *SimpleUpdate, field string, str2Format string) string {
	// pg 的 update-set 不能有表名
	if info.Driver == DriverPostgres {
		return fmt.Sprintf(PGStringLiteralSafe, fmt.Sprintf("%v", field))
	}
	return fmt.Sprintf(str2Format, info.Table, fmt.Sprintf("%v", field))
}
