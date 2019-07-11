package api_server

import(

	"github.com/gin-gonic/gin"

	"go.etcd.io/etcd/clientv3"
)


type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}

func SaveJob(ctx *gin.Context){
	var (
		err error
		postJob string
		job common.Job
		oldJob *common.Job
		bytes []byte
	)
}

func AddJob(ctx *gin.Context){

}

func DeleteJob(ctx *gin.Context){

}

func KillJob(ctx *gin.Context){

}

func GetEyasList(ctx *gin.Context){

}
