package demo

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Result struct {
	AreaCode    string  `json:"areaCode"`
	TotalAmount float64 `json:"totalAmount"`
}

/**
SELECT
	"ht_order"."id" AS "tid",
	SUM(amount) AS total_amount
FROM "public"."ht_order" "ho"
WHERE "ho"."gmt_time" > '2024-07-22 10:00:00'
AND "ho"."gmt_time" < '2024-07-23 10:00:00'
GROUP BY "ho"."id"
ORDER BY "ho"."total_amount" DESC
LIMIT 100
*/

func GetNativeSql(c *gin.Context, limitN int) (string, error) {
	// 从 Gin 上下文中获取 db 对象
	db := c.MustGet("db").(*gorm.DB)

	sql := Gen(db, limitN)

	return sql, nil
}

func GetProductPieResult(c *gin.Context, category string) ([]map[string]any, error) {
	// 从 Gin 上下文中获取 db 对象
	db := c.MustGet("db").(*gorm.DB)

	data := GenProduct(db, category)

	return data, nil
}
