package api_server

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"

	"github.com/ricky1122alonefe/hawkEye-go/module"
)

func KeepAlive(ctx *gin.Context) {
	var (
		req module.KeepAliveRequest
	)
	reqBody, _ := ioutil.ReadAll(ctx.Request.Body)
	if err := json.Unmarshal(reqBody, req); err != nil {
		log.Warning(err.Error())
	}
	log.Info("reciveing msg from " + req.Addr + req.Msg + req.TimeStamp)
}
