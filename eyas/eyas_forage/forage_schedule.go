package eyas_forage

import (

	"time"
	"fmt"

	"github.com/ricky1122alonefe/hawkEye-go/module"
)

// 任务调度
type Scheduler struct {
	jobEventChan chan *module.JobEvent	//  etcd任务事件队列
	jobPlanTable map[string]*module.JobSchedulePlan // 任务调度计划表
	jobNormalPlanTable map[string] *module.JobNormalPlan// 普通脚本调度计划表
	jobExecutingTable map[string]*module.JobExecuteInfo // 任务执行表
	jobExecutingNormalTable map[string]*module.JobNormalExecuteInfo //普通任务执行列表
	jobResultChan chan *module.JobExecuteResult	// 任务结果队列
	jobNormalChan chan *module.JobNormalExecuteResult// 普通任务执行队列
}

var (
	forage_scheduler *Scheduler
)

// 处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *module.JobEvent) {
	var (
		jobSchedulePlan *module.JobSchedulePlan
		jobExecuteInfo *module.JobExecuteInfo
		jobExecuting bool
		jobExisted bool
		err error
	)
	switch jobEvent.EventType {
	case module.JOB_EVENT_SAVE:	// 保存任务事件
		if jobSchedulePlan, err = module.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case module.JOB_EVENT_DELETE: // 删除任务事件
		if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	case module.JOB_EVENT_KILL: // 强杀任务事件
		// 取消掉Command执行, 判断任务是否在执行中
		if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc()	// 触发command杀死shell子进程, 任务得到退出
		}
	}
}
//处理普通任务事件
func (scheduler *Scheduler) handleNormalJobEvent(jobEvent *module.NormalJobEvent) {
	var (
		jobSchedulePlan *module.JobNormalPlan
		jobExecuteInfo *module.JobExecuteInfo
		jobExecuting bool
		jobExisted bool
		err error
	)
	switch jobEvent.EventType {
	case module.JOB_EVENT_SAVE:	// 保存任务事件
		if jobSchedulePlan, err = module.BuildjobNormalPlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobNormalPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case module.JOB_EVENT_DELETE: // 删除任务事件
		if jobSchedulePlan, jobExisted = scheduler.jobNormalPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	case module.JOB_EVENT_KILL: // 强杀任务事件
		// 取消掉Command执行, 判断任务是否在执行中
		if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc()	// 触发command杀死shell子进程, 任务得到退出
		}
	}
}

// 尝试执行任务
func (scheduler *Scheduler) TryStartScheduleJob(jobPlan *module.JobSchedulePlan) {
	// 调度 和 执行 是2件事情
	var (
		jobExecuteInfo *module.JobExecuteInfo
		jobExecuting bool
	)

	// 执行的任务可能运行很久, 1分钟会调度60次，但是只能执行1次, 防止并发！

	// 如果任务正在执行，跳过本次调度
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		// fmt.Println("尚未退出,跳过执行:", jobPlan.Job.Name)
		return
	}

	// 构建执行状态信息
	jobExecuteInfo = module.BuildJobExecuteInfo(jobPlan)

	// 保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	// 执行任务
	fmt.Println("执行任务:", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
	G_executor.ExecuteJob(jobExecuteInfo)
}

// 尝试普通任务
func (scheduler *Scheduler) TryStartNormalJob(jobPlan *module.JobNormalPlan) {
	// 调度 和 执行 是2件事情
	var (
		jobExecuteInfo *module.JobNormalExecuteInfo
		jobExecuting bool
	)

	// 执行的任务可能运行很久, 1分钟会调度60次，但是只能执行1次, 防止并发！

	// 如果任务正在执行，跳过本次调度
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingNormalTable[jobPlan.Job.Name]; jobExecuting {
		log.Info("task ==>>>> still running")
		return
	}

	// 构建执行状态信息
	jobExecuteInfo = module.BuildNormalJobExecuteInfo(jobPlan)

	// 保存执行状态
	scheduler.jobExecutingNormalTable[jobPlan.Job.Name] = jobExecuteInfo

	// 执行任务
	log.Info("执行任务:"+ jobExecuteInfo.Job.Name+jobExecuteInfo.PlanTime.String()+ jobExecuteInfo.RealTime.String())
	G_executor.ExecuteNormalJob(jobExecuteInfo)
}

// 重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule() (scheduleAfter time.Duration) {
	var (
		jobPlan *module.JobSchedulePlan
		now time.Time
		nearTime *time.Time
	)

	// 如果任务表为空话，随便睡眠多久
	if len(scheduler.jobPlanTable) == 0 {
		scheduleAfter = 1 * time.Second
		return
	}

	// 当前时间
	now = time.Now()

	// 遍历所有任务
	for _, jobPlan = range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			scheduler.TryStartScheduleJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now) // 更新下次执行时间
		}

		// 统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	// 下次调度间隔（最近要执行的任务调度时间 - 当前时间）
	scheduleAfter = (*nearTime).Sub(now)
	return
}

// 处理任务结果
func (scheduler *Scheduler) handleJobResult(result *module.JobExecuteResult) {
	var (
		jobLog *module.JobLog
	)
	// 删除执行状态
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)

	// 生成执行日志
	if result.Err != module.ERR_LOCK_ALREADY_REQUIRED {
		jobLog = &module.JobLog{
			JobName: result.ExecuteInfo.Job.Name,
			Command: result.ExecuteInfo.Job.Command,
			Output: string(result.Output),
			PlanTime: result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
			ScheduleTime: result.ExecuteInfo.RealTime.UnixNano() / 1000 / 1000,
			StartTime: result.StartTime.UnixNano() / 1000 / 1000,
			EndTime: result.EndTime.UnixNano() / 1000 / 1000,
		}
		if result.Err != nil {
			jobLog.Err = result.Err.Error()
		} else {
			jobLog.Err = ""
		}
		G_logSink.Append(jobLog)
	}

	// fmt.Println("任务执行完成:", result.ExecuteInfo.Job.Name, string(result.Output), result.Err)
}

// 调度协程
func (scheduler *Scheduler) scheduleLoop() {
	var (
		jobEvent *module.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobResult *module.JobExecuteResult
	)

	// 初始化一次(1秒)
	scheduleAfter = scheduler.TrySchedule()

	// 调度的延迟定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	// 定时任务common.Job
	for {
		select {
		case jobEvent = <- scheduler.jobEventChan:	//监听任务变化事件
			// 对内存中维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		case <- scheduleTimer.C:	// 最近的任务到期了
		case jobResult = <- scheduler.jobResultChan: // 监听任务执行结果
			scheduler.handleJobResult(jobResult)
		}
		// 调度一次任务
		scheduleAfter = scheduler.TrySchedule()
		// 重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}

// 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *module.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

// 初始化调度器
func InitScheduler() (err error) {
	forage_scheduler = &Scheduler{
		jobEventChan: make(chan *module.JobEvent, 1000),
		jobPlanTable: make(map[string]*module.JobSchedulePlan),
		jobExecutingTable: make(map[string]*module.JobExecuteInfo),
		jobResultChan: make(chan *module.JobExecuteResult, 1000),
	}
	// 启动调度协程
	go forage_scheduler.scheduleLoop()
	return
}

// 回传任务执行结果
func (scheduler *Scheduler) PushJobResult(jobResult *module.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}
//回传普通任务执行结果
func (scheduler *Scheduler) PushNormalJobResult(jobResult *module.JobNormalExecuteResult) {
	scheduler.jobNormalChan <- jobResult
}