package module

type KeepAliveRequest struct {
	Addr      string `json:"addr"`
	Msg       string `json:"msg"`
	TimeStamp string `json:"timeStamp"`
}
