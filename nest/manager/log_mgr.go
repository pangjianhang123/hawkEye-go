package manager

import (
	"gopkg.in/mgo.v2"


	"github.com/ricky1122alonefe/hawkEye-go/module"
	"log"

)

// mongodb日志管理
type LogMgr struct {
	session *mgo.Session
	logCollection *mgo.Collection
}

var (
	G_logMgr *LogMgr
)

func InitLogMgr() (err error) {
	var (
		session *mgo.Session
	)

	// 建立mongodb连接

	if session,err= mgo.Dial(G_config.MongodbUri);err!=nil{
		log.Println(err.Error())
	}
	G_logMgr = &LogMgr{
		session: session,
		logCollection: session.DB("cron").C("log"),
	}
	return
}

// 查看任务日志
func (logMgr *LogMgr) ListLog(name string, skip int, limit int) (logArr []*module.JobLog, err error){
	var (
		filter *module.JobLogFilter
		logSort *module.SortLogByStartTime
		jobLog *module.JobLog
	)

	// len(logArr)
	logArr = make([]*module.JobLog, 0)

	// 过滤条件
	filter = &module.JobLogFilter{JobName: name}

	// 按照任务开始时间倒排
	logSort = &module.SortLogByStartTime{SortOrder: -1}

	if err=logMgr.logCollection.Find(filter).All(&logArr);err!=nil{
		return
	}
	// 查询

	return
}