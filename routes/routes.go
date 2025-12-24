package routes

import (
	"github.com/forhsd/sqlbuilder/controller"
	"github.com/forhsd/sqlbuilder/demo"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 不需要v1版本的公共服务的路由放外面
	// r.POST("/auth", api.GetAuth)

	apiV1 := router.Group("/api/v1")
	apiDemo := router.Group("/api/demo")

	{
		{
			// golang 练习
			apiDemo.GET("/database/:id", demo.GetDatabase)
			// curl -X GET 'localhost:7077/api/demo/sqlbuilder/gen?limit=10' | jq
			apiDemo.GET("/sqlbuilder/gen", demo.GetGenerateSql)
			// sudo apt install jq
			// curl -X POST localhost:7077/api/demo/sqlbuilder/product -d '{"category":""}' -H "Content-Type: application/json" -s | jq
			// curl -X POST localhost:7077/api/demo/sqlbuilder/product -d '{"category":"Doohickey"}' -H "Content-Type: application/json" -s | jq
			apiDemo.POST("/sqlbuilder/product", demo.GetGenProduct)
		}

		// 暴露接口

		// http协议生成sql
		apiV1.POST("/sqlbuilder", controller.GenerateHBQL)
		// proto协议生成sql
		apiV1.POST("/sqlbuilder/proto", controller.GenerateHBQLByProto)
		// 解析主从模板
		apiV1.POST("/sqlbuilder/template-analyze", controller.AnalyzeTemplatesByProto)
		// 解析附加模板
		apiV1.POST("/sqlbuilder/addition-analyze", controller.AnalyzeAdditionByProto)
	}
}
