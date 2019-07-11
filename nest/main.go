package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ricky1122alonefe/hawkEye-go/nest/route"
	"github.com/ricky1122alonefe/hawkEye-go/nest/manager"
)

func main() {
	r:=route.InitRouter()
	gin.SetMode(gin.ReleaseMode)
	manager.InitConfig("")
	r.Run()
}
