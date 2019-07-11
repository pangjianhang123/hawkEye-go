package module

type SaveJobRequest struct{
	Name string `json:"name"`	//  任务名
	Command string	`json:"command"` // shell命令
	CronExpr string	`json:"cronExpr"`	// cron表达式
}


type DeleteJobRequest struct{
	JobName string `json:"job_name"`
}