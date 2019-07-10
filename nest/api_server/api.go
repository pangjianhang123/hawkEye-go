package api_server

import(
	//"github.com/apsdehal/go-logger"
	"github.com/gin-gonic/gin"

	"go.etcd.io/etcd/clientv3"
)


type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}

func SaveJob(ctx *gin.Context){

}

func AddJob(ctx *gin.Context){

}

func DeleteJob(ctx *gin.Context){

}

func KillJob(ctx *gin.Context){

}
