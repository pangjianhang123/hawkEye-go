package hooks

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

type WebHook struct {
	WebHookID      int       ` json:"web_hook_id"`
	RepositoryName string    `json:"repository_name"`
	BranchName     string    `json:"branch_name"`
	Tag            string    `json:"tag"`
	Shell          string    `json:"shell"`
	Status         int       `json:"status"`
	Key            string    `json:"key"`
	Secure         string    `json:"secure"`
	LastExecTime   time.Time `json:"last_exec_time"`
	CreateTime     time.Time `json:"create_time"`
	HookType       string    `json:"hook_type"`
	CreateAt       int       `json:"create_at"`
}

// 生成32位MD5
func Md5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}
