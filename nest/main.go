package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ricky1122alonefe/hawkEye-go/nest/route"
)

func main() {
	r:=route.InitRouter()
	gin.SetMode(gin.ReleaseMode)
	r.Run()
}