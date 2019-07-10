package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ricky1122alonefe/hawkEye-go/nest/route"
	"fmt"
)

func main() {
	fmt.Println("hello world")
	r:=route.InitRouter()
	gin.SetMode(gin.ReleaseMode)
	r.Run()
}