package clause

import (
	"fmt"
	"strings"
	"testing"
)

var (
	newUserId       = "8878"
	newEnterpriseId = "4782"

	oldUserId       = "5403"
	oldEnterpriseId = "4330"

	// 新用户id和企业
	newSets = map[string]interface{}{
		"user_id":       newUserId,
		"enterprise_id": newEnterpriseId,
	}
	// 旧用户id和企业
	oldWhere = map[string]interface{}{
		"user_id":       oldUserId,
		"enterprise_id": oldEnterpriseId,
	}

	schema = "holder_bi"

	// 占位符依次: 企业, 用户
	FillDataFormat   = `'nc_base_fill_%s_%s'`
	ImportDataFormat = `'nc_base_import_%s_%s'`

	StringValueFormat = `'%s'`
)

func TestBuildSimpleSql(t *testing.T) {
	info := SimpleUpdate{
		Driver: DriverMysql,
		Schema: schema,
		Table:  "api_ds",
		Sets:   newSets,
		Where:  oldWhere,
	}

	sql := info.BuildSimpleSql()
	fmt.Println("生成SQL: \r\n", sql)
}

func TestBuildBatchSimpleSql(t *testing.T) {
	tables := []string{
		"card", // 要单独处理, 因为动态sql和etl混在一起
		"chart_filter_config",
		"collection",
		"custom_sorting_config",
		"data_fill_submittal",
		"data_import",
		"dataset_bookmark",
		"dataset_drill_config",
		"dynamic_template_dataset",
		"export_task_info",
		"public_share",
		"report_view_base",
		"fill_field_map",
		"view_global_chart",
		"view_global_config",
		"view_global_filter",
		"view_global_table",
	}

	//otherAddTables := []string{
	//	"data_fill_submittal",
	//	"data_import",
	//	"dynamic_template_dataset",
	//	"report_view_base",
	//	"public_share",
	//}

	var sqlBatch []string
	for _, tableItem := range tables {
		newSetsCp := copyMap(newSets)
		// 旧用户名，如果一样可省略
		//if slices.Contains(otherAddTables, tableItem) {
		//	newSetsCp["user_name"] = `'ABC'`
		//}

		item := SimpleUpdate{
			Driver: DriverMysql,
			Schema: schema,
			Table:  tableItem,
			Sets:   newSetsCp,
			Where:  oldWhere,
		}
		sql := item.BuildSimpleSql()
		sqlBatch = append(sqlBatch, sql)
	}

	sqlBatchStr := strings.Join(sqlBatch, "\r\n")
	fmt.Println("生成SQL:\r\n", sqlBatchStr)
}

func copyMap[K comparable, V any](src map[K]V) map[K]V {
	dst := make(map[K]V, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func TestBuildNCSimpleSql(t *testing.T) {
	var sqlBatch []string

	info1 := SimpleUpdate{
		Driver: DriverPostgres,
		Schema: "public",
		Table:  "nc_base_users_v2",
		Sets: map[string]interface{}{
			"fk_user_id": fmt.Sprintf(StringValueFormat, newUserId),
		},
		// 旧信息
		Where: map[string]interface{}{
			// // 占位符依次: 企业, 用户
			"base_id":    fmt.Sprintf(FillDataFormat, oldEnterpriseId, oldUserId),
			"fk_user_id": fmt.Sprintf(StringValueFormat, oldUserId),
		},
	}

	sql1 := info1.BuildSimpleSql()
	sqlBatch = append(sqlBatch, sql1)

	info2 := SimpleUpdate{
		Driver: DriverPostgres,
		Schema: "public",
		Table:  "nc_base_users_v2",
		Sets: map[string]interface{}{
			"fk_user_id": fmt.Sprintf(StringValueFormat, newUserId),
		},
		// 旧信息
		Where: map[string]interface{}{
			// // 占位符依次: 企业, 用户
			"base_id":    fmt.Sprintf(ImportDataFormat, oldEnterpriseId, oldUserId),
			"fk_user_id": fmt.Sprintf(StringValueFormat, oldUserId),
		},
	}

	sql2 := info2.BuildSimpleSql()
	sqlBatch = append(sqlBatch, sql2)

	sqlBatchStr := strings.Join(sqlBatch, "\r\n")
	fmt.Println("生成SQL: \r\n", sqlBatchStr)
}
