package manager

import (
	"time"
	"encoding/json"
	"context"
	"os"


	//"go.etcd.io/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"github.com/apsdehal/go-logger"

	"github.com/ricky1122alonefe/hawkEye-go/module"
)

const (
	ALL = "all"
	SCHEDULE = "schedule"
	SIMPLE = "simple"
)

var (
	log     *logger.Logger
	log_err error
)

// 任务管理器
type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}

var (
	// 单例
	G_jobMgr *JobMgr
)

func init() {
	log, log_err = logger.New("test", 1, os.Stdout)
}
// 初始化管理器
func InitJobMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
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

	// 赋值单例
	G_jobMgr = &JobMgr{
		client: client,
		kv: kv,
		lease: lease,
	}
	return
}

// 保存cron任务
func (jobMgr *JobMgr) SaveJob(job *module.ScheduleJob) (oldJob *module.ScheduleJob, err error) {
	// 把任务保存到/cron/jobs/任务名 -> json
	var (
		jobKey string
		jobValue []byte
		putResp *clientv3.PutResponse
		oldJobObj module.ScheduleJob
	)

	// etcd的保存key
	jobKey = module.SCHE_JOB_SAVE_DIR +job.Name
	// 任务信息json
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	// 保存到etcd
	if putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}
	// 如果是更新, 那么返回旧值
	if putResp.PrevKv != nil {
		// 对旧值做一个反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

//普通脚本任务保存与添加
func (jobMgr *JobMgr)SaveNormalJob(job *module.SimpleJob)(oldJob *module.SimpleJob,err error){
	var (
		jobKey string
		jobValue []byte
		putResp *clientv3.PutResponse
		oldJobObj module.SimpleJob
	)

	// etcd的保存key
	jobKey = module.SIMP_JOB_SAVE_DIR +job.Name
	// 任务信息json
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	// 保存到etcd
	if putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}
	// 如果是更新, 那么返回旧值
	if putResp.PrevKv != nil {
		// 对旧值做一个反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}
// 删除任务
func (jobMgr *JobMgr) DeleteJob(name string) (oldJob *module.ScheduleJob, err error) {
	var (
		jobKey string
		delResp *clientv3.DeleteResponse
		oldJobObj module.ScheduleJob
	)

	// etcd中保存任务的key
	jobKey = module.SCHE_JOB_SAVE_DIR + name

	// 从etcd中删除它
	if delResp, err = jobMgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}

	// 返回被删除的任务信息
	if len(delResp.PrevKvs) != 0 {
		// 解析一下旧值, 返回它
		if err =json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

// 列举任务
func (jobMgr *JobMgr) ListJobs(jobType string) (jobscheList []*module.ScheduleJob,normalList []*module.SimpleJob, err error) {
	var (
		dirKey string
		getResp *clientv3.GetResponse
		//kvPair *mvccpb.KeyValue
		//job *module.ScheduleJob
		jobList = make([]*module.ScheduleJob, 0)
		simpleList = make([]*module.SimpleJob,0)
	)

	// 任务保存的目录


	// 获取目录下所有任务信息
	if getResp, err = jobMgr.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return
	}

	switch jobType{

	case ALL:

		dirKey = module.ALL_JOB_SAVE_DIR
		if getResp,err = GetJobsByTypes(dirKey,jobMgr);err!=nil{
			log.Critical(err.Error())
		}
		jobList,_,_  = GetJobs(getResp,SCHEDULE)
		_,simpleList,_=GetJobs(getResp,SIMPLE)


	case SCHEDULE:

		dirKey = module.SCHE_JOB_SAVE_DIR

		if getResp,err = GetJobsByTypes(dirKey,jobMgr);err!=nil{
			log.Critical(err.Error())
		}
		jobList,simpleList,_  = GetJobs(getResp,jobType)

	case SIMPLE:

		dirKey = module.SIMP_JOB_SAVE_DIR

		if getResp,err = GetJobsByTypes(dirKey,jobMgr);err!=nil{
			log.Critical(err.Error())
		}
		jobList,simpleList,_  = GetJobs(getResp,jobType)
	}


	return jobList,simpleList,err
}



func GetJobsByTypes(jobType string,jbMgr *JobMgr)( *clientv3.GetResponse, error){
	var getResp *clientv3.GetResponse
	var err error
	if getResp, err = jbMgr.kv.Get(context.TODO(), jobType, clientv3.WithPrefix()); err != nil {
		return nil,err
	}
	return getResp,nil
}

func GetJobs(jbs *clientv3.GetResponse,jobType string)(jobscheList []*module.ScheduleJob,normalList []*module.SimpleJob, err error){

	var (
	jobList = make([]*module.ScheduleJob, 0)
	simpleList = make([]*module.SimpleJob,0)
	job *module.ScheduleJob
	sJob *module.SimpleJob
	)

	switch jobType{

	case SCHEDULE:
		for _, kvPair := range jbs.Kvs {
			job = &module.ScheduleJob{}
			if err =json.Unmarshal(kvPair.Value, job); err != nil {
				err = nil
				continue
			}
			jobList = append(jobList, job)
		}
	case SIMPLE:
		for _, kvPair := range jbs.Kvs {
			sJob = &module.SimpleJob{}
			if err =json.Unmarshal(kvPair.Value, job); err != nil {
				err = nil
				continue
			}
			simpleList = append(simpleList, sJob)
		}
	}
	return jobList,simpleList,nil
}

// 杀死任务
func (jobMgr *JobMgr) KillJob(name string) (err error) {
	// 更新一下key=/cron/killer/任务名
	var (
		killerKey string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId clientv3.LeaseID
	)

	// 通知worker杀死对应任务
	killerKey = module.JOB_KILLER_DIR + name

	// 让worker监听到一次put操作, 创建一个租约让其稍后自动过期即可
	if leaseGrantResp, err = jobMgr.lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	// 租约ID
	leaseId = leaseGrantResp.ID

	// 设置killer标记
	if _, err = jobMgr.kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}