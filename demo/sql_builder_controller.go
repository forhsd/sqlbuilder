package demo

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/forhsd/logger"
	"github.com/forhsd/sqlbuilder/common"
	"github.com/gin-gonic/gin"
)

func GetGenerateSql(c *gin.Context) {
	// url路径传参 => id := c.Param("id")
	// url问号传参 => key := c.Query("key")
	// body传参 => if err := c.ShouldBindJSON(&req); err != nil { return }

	limitNStr := c.Query("limit")

	limitN, limitErr := strconv.Atoi(limitNStr)
	if limitErr != nil {
		fmt.Println("入参不是int类型字符串:", limitErr)
		return
	}

	sql, err := GetNativeSql(c, limitN)
	logger.Debug("原生sql: \r\n", sql)

	if err != nil {
		resp := common.Response{
			Code: http.StatusInternalServerError,
			Msg:  "生成sql失败",
			Data: nil,
		}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	result := map[string]string{"sql": sql}

	resp := common.Response{
		Code: http.StatusOK,
		Msg:  "生成sql成功",
		Data: result,
	}

	c.JSON(http.StatusOK, resp)
}

type ProductRequest struct {
	Category string `json:"category"`
}

type ProductResponse struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

func GetGenProduct(c *gin.Context) {

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := GetProductPieResult(c, req.Category)

	if err != nil {
		resp := common.Response{
			Code: http.StatusInternalServerError,
			Msg:  "product生成sql失败",
			Data: nil,
		}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	result := map[string]any{"rows": data}

	resp := common.Response{
		Code: http.StatusOK,
		Msg:  "product生成sql成功",
		Data: result,
	}

	c.JSON(http.StatusOK, resp)

}
