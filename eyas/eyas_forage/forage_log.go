package eyas_forage

import (

	"github.com/ricky1122alonefe/hawkEye-go/module"
	"github.com/ricky1122alonefe/hawkEye-go/eyas/config"
	"context"
	"os"
	"gopkg.in/mgo.v2"

	"github.com/apsdehal/go-logger"
	"time"
)
var (
	log     *logger.Logger
	log_err error
)

var (
	G_config *config.Config
)

func init() {
	config.InitConfig("")
	log, log_err = logger.New("test", 1, os.Stdout)
}

// mongodb存储日志
type LogSink struct {
	client *mgo.Session
	logCollection *mgo.Collection
	logChan chan *module.JobLog
	autoCommitChan chan *module.LogBatch
}

var (
	G_logSink *LogSink
)

// 批量写入日志
func (logSink *LogSink) saveLogs(batch *module.LogBatch) {
	logSink.logCollection.Insert(context.TODO(), batch.Logs)
}

// 日志存储协程
func (logSink *LogSink) writeLoop() {
	var (
		log *module.JobLog
		logBatch *module.LogBatch // 当前的批次
		commitTimer *time.Timer
		timeoutBatch *module.LogBatch // 超时批次
	)

	for {
		select {
		case log = <- logSink.logChan:
			if logBatch == nil {
				logBatch = &module.LogBatch{}
				// 让这个批次超时自动提交(给1秒的时间）
				commitTimer = time.AfterFunc(
					time.Duration(G_config.JobLogCommitTimeout) * time.Millisecond,
					func(batch *module.LogBatch) func() {
						return func() {
							logSink.autoCommitChan <- batch
						}
					}(logBatch),
				)
			}

			// 把新日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)

			// 如果批次满了, 就立即发送
			if len(logBatch.Logs) >= G_config.JobLogBatchSize {
				// 发送日志
				logSink.saveLogs(logBatch)
				// 清空logBatch
				logBatch = nil
				// 取消定时器
				commitTimer.Stop()
			}
		case timeoutBatch = <- logSink.autoCommitChan: // 过期的批次
			// 判断过期批次是否仍旧是当前的批次
			if timeoutBatch != logBatch {
				continue // 跳过已经被提交的批次
			}
			// 把批次写入到mongo中
			logSink.saveLogs(timeoutBatch)
			// 清空logBatch
			logBatch = nil
		}
	}
}

func InitLogSink() (err error) {
	var (
		client *mgo.Session

	)

	// 建立mongodb连接
	if client,err= mgo.Dial(G_config.MongodbUri);err!=nil{
		log.Critical(err.Error())
		return
	}

	//   选择db和collection
	G_logSink = &LogSink{
		client: client,
		logCollection: client.DB("cron").C("log"),
		logChan: make(chan *module.JobLog, 1000),
		autoCommitChan: make(chan *module.LogBatch, 1000),
	}

	// 启动一个mongodb处理协程
	go G_logSink.writeLoop()
	return
}

// 发送日志
func (logSink *LogSink) Append(jobLog *module.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
		// 队列满了就丢弃
	}
}