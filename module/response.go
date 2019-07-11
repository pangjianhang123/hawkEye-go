package module

type SaveJobResponse struct{
	Code int64 `json:"code"`
	Msg string `json:"msg"`
	ScheduleJob
}


type JobListResponse struct{
	Code int64 `json:"code"`
	Msg string `json:"msg"`
	List []*ScheduleJob `json:"list"`
}

type KillJobResponse struct{
	Code int64 `json:"code"`
	Msg string `json:"msg"`
}