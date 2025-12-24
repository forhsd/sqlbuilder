package demo

import (
	"net/http"

	"github.com/forhsd/logger"

	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"

	"github.com/forhsd/sqlbuilder/common"
)

func GetDatabase(c *gin.Context) {
	id := com.StrTo(c.Param("id")).MustInt()
	logger.Debug("################", id)

	if id <= 0 {
		resp := common.Response{
			Code: http.StatusBadRequest,
			Msg:  "参数错误，id必须大于0",
			Data: nil,
		}
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	database, err := GetOneDatabase(c, id)
	if err != nil {
		resp := common.Response{
			Code: http.StatusInternalServerError,
			Msg:  "获取数据库失败",
			Data: nil,
		}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := common.Response{
		Code: http.StatusOK,
		Msg:  "获取数据库成功",
		Data: database,
	}
	c.JSON(http.StatusOK, resp)
}
