package route

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	api "github.com/ricky1122alonefe/hawkEye-go/nest/api_server"

)

func InitRouter() *gin.Engine {
	r := gin.New()
	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json") //The url pointing to API definition
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	r.POST("/api/v1/save",api.SaveJob)
	r.POST("/api/v1/add",api.AddJob)
	r.POST("/api/v1/delete",api.DeleteJob)
	r.POST("/api/v1/kill",api.KillJob)
	return r
}