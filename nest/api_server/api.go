package api_server

import (
	"github.com/gin-gonic/gin"

	"go.etcd.io/etcd/clientv3"

	"encoding/json"
	"github.com/ricky1122alonefe/hawkEye-go/module"
	"github.com/ricky1122alonefe/hawkEye-go/nest/manager"
	"io/ioutil"

	"github.com/apsdehal/go-logger"
	"os"
	"net/http"
)

var log *logger.Logger
var log_err error

func init() {
	log, log_err = logger.New("test", 1, os.Stdout)
}

type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

func SaveJob(ctx *gin.Context) {
	var (
		err     error
		job     module.ScheduleJob
		oldJob  *module.ScheduleJob
		req module.SaveJobRequest
		resp module.SaveJobResponse
	)

	reqBody, _ := ioutil.ReadAll(ctx.Request.Body)
	if err = json.Unmarshal(reqBody, req); err != nil {
		log.Critical(err.Error())
	}
	job.Name = req.Name
	job.Command = req.Command
	job.CronExpr = req.CronExpr

	if oldJob, err = manager.G_jobMgr.SaveJob(&job); err != nil {
		log.Critical(err.Error())
	}

	resp.Code = http.StatusOK
	resp.Msg = "success"
	resp.ScheduleJob = *oldJob

	ctx.JSON(200,resp)
}

func AddJob(ctx *gin.Context) {

}

func DeleteJob(ctx *gin.Context) {
	var (
		err error	// interface{}
		name string
		req module.DeleteJobRequest
		oldJob *module.ScheduleJob
		resp module.SaveJobResponse

	)
	reqBody, _ := ioutil.ReadAll(ctx.Request.Body)
	if err = json.Unmarshal(reqBody, req); err != nil {
		log.Critical(err.Error())
	}
	// POST:   a=1&b=2&c=3

	// 去删除任务
	if oldJob, err = manager.G_jobMgr.DeleteJob(name); err != nil {
		log.Critical(err.Error())
		resp.Code = http.StatusBadRequest
		resp.Msg = err.Error()
		ctx.JSON(400,resp)
	}

	// 正常应答
	resp.Code= http.StatusOK
	resp.Msg = "success"
	resp.ScheduleJob = *oldJob
	ctx.JSON(200,resp)
}

func KillJob(ctx *gin.Context) {
	var (
		err error
		name string
		resp module.KillJobResponse
		req module.DeleteJobRequest
	)

	reqBody, _ := ioutil.ReadAll(ctx.Request.Body)
	if err = json.Unmarshal(reqBody, req); err != nil {
		log.Critical(err.Error())
	}

	if err = manager.G_jobMgr.KillJob(name); err != nil {
		resp.Code = http.StatusBadRequest
		resp.Msg = err.Error()
		ctx.JSON(400,resp)
	}

	// 正常应答

}

func GetEyasList(ctx *gin.Context) {
	var (
		jobList []*module.ScheduleJob
		resp module.JobListResponse
		err error
	)

	// 获取任务列表
	if jobList, err = manager.G_jobMgr.ListJobs(); err != nil {
		resp.Code = http.StatusOK
		resp.Msg = err.Error()
		ctx.JSON(400,resp)
	}

	resp.Code = http.StatusOK
	resp.Msg = "success"
	resp.List = jobList
	ctx.JSON(200,resp)
	// 正常应答

}
