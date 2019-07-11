package api_server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/apsdehal/go-logger"
	"github.com/gin-gonic/gin"
	"go.etcd.io/etcd/clientv3"

	"github.com/ricky1122alonefe/hawkEye-go/module"
	"github.com/ricky1122alonefe/hawkEye-go/nest/manager"
)

var (
	log     *logger.Logger
	log_err error
)



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
		err    error
		job    module.ScheduleJob
		oldJob *module.ScheduleJob
		req    module.SaveJobRequest
		resp   module.SaveJobResponse
	)

	reqBody, _ := ioutil.ReadAll(ctx.Request.Body)

	if err = json.Unmarshal(reqBody, req); err != nil {
		log.Critical(err.Error())
		resp.Code =http.StatusBadRequest
		resp.Msg = err.Error()
		ctx.JSON(400,resp)
		return
	}

	job.Name = req.Name
	job.Command = req.Command
	job.CronExpr = req.CronExpr

	if oldJob, err = manager.G_jobMgr.SaveJob(&job); err != nil {
		log.Critical(err.Error())
		log.Critical(err.Error())
		resp.Code =http.StatusBadRequest
		resp.Msg = err.Error()
		ctx.JSON(400,resp)
		return
	}

	resp.Code = http.StatusOK
	resp.Msg = "success"
	resp.ScheduleJob = *oldJob

	ctx.JSON(200, resp)
}

func SaveSimpleJob(ctx *gin.Context){
	var (
		err    error
		job    module.SimpleJob
		oldJob *module.SimpleJob
		req    module.SaveJobRequest
		resp   module.SaveNormalJobResponse
	)

	reqBody, _ := ioutil.ReadAll(ctx.Request.Body)

	if err = json.Unmarshal(reqBody, req); err != nil {
		log.Critical(err.Error())
		resp.Code =http.StatusBadRequest
		resp.Msg = err.Error()
		ctx.JSON(400,resp)
		return
	}

	job.Name = req.Name
	job.Command = req.Command
	//job.CronExpr = req.CronExpr

	if oldJob, err = manager.G_jobMgr.SaveNormalJob(&job); err != nil {
		log.Critical(err.Error())
		log.Critical(err.Error())
		resp.Code =http.StatusBadRequest
		resp.Msg = err.Error()
		ctx.JSON(400,resp)
		return
	}

	resp.Code = http.StatusOK
	resp.Msg = "success"
	resp.SimpleJob = *oldJob

	ctx.JSON(200, resp)
}

func AddJob(ctx *gin.Context) {

}

func DeleteJob(ctx *gin.Context) {
	var (
		err    error // interface{}
		name   string
		req    module.DeleteJobRequest
		oldJob *module.ScheduleJob
		resp   module.SaveJobResponse
	)
	reqBody, _ := ioutil.ReadAll(ctx.Request.Body)
	if err = json.Unmarshal(reqBody, req); err != nil {
		log.Critical(err.Error())
		resp.Code = http.StatusBadRequest
		resp.Msg = err.Error()
		ctx.JSON(400,resp)
		return
	}

	if oldJob, err = manager.G_jobMgr.DeleteJob(name); err != nil {
		log.Critical(err.Error())
		resp.Code = http.StatusBadRequest
		resp.Msg = err.Error()
		ctx.JSON(400, resp)
	}


	resp.Code = http.StatusOK
	resp.Msg = "success"
	resp.ScheduleJob = *oldJob
	ctx.JSON(200, resp)
}

func KillJob(ctx *gin.Context) {
	var (
		err  error
		name string
		resp module.KillJobResponse
		req  module.DeleteJobRequest
	)

	reqBody, _ := ioutil.ReadAll(ctx.Request.Body)
	if err = json.Unmarshal(reqBody, req); err != nil {
		log.Critical(err.Error())
	}

	if err = manager.G_jobMgr.KillJob(name); err != nil {
		resp.Code = http.StatusBadRequest
		resp.Msg = err.Error()
		ctx.JSON(400, resp)
	}

	// 正常应答

}

func GetEyasList(ctx *gin.Context) {
	var (
		jobList []*module.ScheduleJob
		simpleList []*module.SimpleJob
		resp    module.JobListResponse
		err     error
	)


	jobType :=ctx.Param("type")
	// 获取任务列表
	if jobList,simpleList, err = manager.G_jobMgr.ListJobs(jobType); err != nil {
		resp.Code = http.StatusOK
		resp.Msg = err.Error()
		ctx.JSON(400, resp)
	}

	resp.Code = http.StatusOK
	resp.Msg = "success"
	resp.List = jobList
	resp.ListSimple = simpleList
	ctx.JSON(200, resp)
	// 正常应答

}
