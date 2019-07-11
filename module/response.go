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
	ListSimple []*SimpleJob `json:"listSimple"`
}

type KillJobResponse struct{
	Code int64 `json:"code"`
	Msg string `json:"msg"`
}


type SaveNormalJobResponse struct{
	Code int64 `json:"code"`
	Msg string `json:"msg"`
	SimpleJob
}