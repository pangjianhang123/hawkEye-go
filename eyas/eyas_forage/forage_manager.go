package eyas_forage

import (

	"time"
	"context"
	"github.com/ricky1122alonefe/hawkEye-go/module"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

// 任务管理器
type ForageManager struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
	watcher clientv3.Watcher
}

var (
	// 单例
	forageMgr *ForageManager
)

// 监听cron任务变化
func (forageMgr *ForageManager) watchForageJobs() (err error) {
	var (
		etcdScheduleGetResponse *clientv3.GetResponse
		etcdScheduleKvPair *mvccpb.KeyValue
		watchChanSchedule clientv3.WatchChan

		etcdNormalGetResponse *clientv3.GetResponse
		etcdNormalKvPair *mvccpb.KeyValue
		watchChanNormal  clientv3.WatchChan

		job *module.ScheduleJob
		watchStartRevision int64


		watchRespSchedule clientv3.WatchResponse
		watchRespNormal clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobName string
		jobEvent *module.JobEvent
	)

	//获取目前etcd中存储的所有的定时任务
	if etcdScheduleGetResponse, err = forageMgr.kv.Get(context.TODO(), module.SCHE_JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	//获取目前etcd中存储所有的普通任务
	if etcdNormalGetResponse,err  =  forageMgr.kv.Get(context.TODO(), module.SIMP_JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}


	// 当前有哪些定时任务
	for _, etcdScheduleKvPair = range etcdScheduleGetResponse.Kvs {
		//得到目前每一个定时任务
		if job, err = module.UnpackJob(etcdScheduleKvPair.Value); err == nil {
			jobEvent = module.BuildJobEvent(module.JOB_EVENT_SAVE, job)
			// 同步给scheduler(调度协程)
			forage_scheduler.PushJobEvent(jobEvent)
		}
	}
	for _,etcdNormalKvPair =range etcdNormalGetResponse.Kvs{
		//得到目前每一个普通任务
		if job, err = module.UnpackJob(etcdNormalKvPair.Value); err == nil {
			jobEvent = module.BuildJobEvent(module.JOB_EVENT_SAVE, job)
			// 同步给scheduler(调度协程)
			forage_scheduler.PushJobEvent(jobEvent)
		}
	}

	// 2, 从该revision向后监听变化事件
	go func() {

		watchStartRevision = etcdScheduleGetResponse.Header.Revision + 1
		// 监听/cron/jobs/目录的后续变化
		watchChanSchedule = forageMgr.watcher.Watch(context.TODO(), module.SCHE_JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		watchChanNormal = forageMgr.watcher.Watch(context.TODO(), module.SIMP_JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		// 处理监听事件
		for watchRespSchedule = range watchChanSchedule {
			for _, watchEvent = range watchRespSchedule.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 写入etcd动作
					if job, err = module.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					//构建一个创建事件
					jobEvent = module.BuildJobEvent(module.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE: // 删除etcd操作

					jobName = module.ExtractJobName(string(watchEvent.Kv.Key))

					job = &module.ScheduleJob{Name: jobName}

					//构建一个删除事件
					jobEvent = module.BuildJobEvent(module.JOB_EVENT_DELETE, job)
				}
				// 变化推给scheduler
				forage_scheduler.PushJobEvent(jobEvent)
			}
		}
		for watchRespNormal = range watchChanSchedule {
			for _, watchEvent = range watchRespNormal.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 写入etcd动作
					if job, err = module.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					//构建一个创建事件
					jobEvent = module.BuildJobEvent(module.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE: // 删除etcd操作

					jobName = module.ExtractJobName(string(watchEvent.Kv.Key))

					job = &module.ScheduleJob{Name: jobName}

					//构建一个删除事件
					jobEvent = module.BuildJobEvent(module.JOB_EVENT_DELETE, job)
				}
				// 变化推给scheduler
				forage_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return
}

// 监听取消通知
func (forageMgr *ForageManager) scheduleWatchKiller() {
	var (
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent *module.JobEvent
		jobName string
		job *module.ScheduleJob
	)
	//
	go func() { // 监听协程
		// 监听/cron/killer/目录的变化
		watchChan = forageMgr.watcher.Watch(context.TODO(), module.JOB_KILLER_DIR, clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 杀死任务事件
					jobName = module.ExtractKillerName(string(watchEvent.Kv.Key))
					job = &module.ScheduleJob{Name: jobName}
					jobEvent = module.BuildJobEvent(module.JOB_EVENT_KILL, job)
					// 事件推给scheduler
					forage_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE: // killer标记过期, 被自动删除
				}
			}
		}
	}()
}

// 初始化管理器
func InitJobMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
		watcher clientv3.Watcher
	)

	// 初始化配置
	config = clientv3.Config{
		Endpoints: G_config.EtcdEndpoints, // 集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, // 连接超时
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	// 赋值单例
	forageMgr = &ForageManager{
		client: client,
		kv: kv,
		lease: lease,
		watcher: watcher,
	}

	// 启动任务监听
	forageMgr.watchForageJobs()

	// 启动监听killer
	forageMgr.scheduleWatchKiller()

	return
}

// 创建任务执行锁
func (forageMgr *ForageManager) CreateJobLock(jobName string) (jobLock *JobLock){
	jobLock = InitJobLock(jobName, forageMgr.kv, forageMgr.lease)
	return
}