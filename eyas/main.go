package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/ricky1122alonefe/hawkEye-go/eyas/config"
	"github.com/ricky1122alonefe/hawkEye-go/eyas/eyas_forage"
)

var (
	confFile string // 配置文件路径
)

// 解析命令行参数
func initArgs() {
	// worker -config ./worker.json
	// worker -h
	flag.StringVar(&confFile, "config", "./eyas.json", "eyas.json")
	flag.Parse()
}

// 初始化线程数量
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)

	// 初始化命令行参数
	initArgs()

	// 初始化线程
	initEnv()

	// 加载配置
	if err = config.InitConfig(confFile); err != nil {
		goto ERR
	}

	// 服务注册
	if err = eyas_forage.InitRegister(); err != nil {
		goto ERR
	}

	// 启动日志协程
	if err = eyas_forage.InitLogSink(); err != nil {
		goto ERR
	}

	// 启动执行器
	if err = eyas_forage.InitExecutor(); err != nil {
		goto ERR
	}

	// 启动调度器
	if err = eyas_forage.InitScheduler(); err != nil {
		goto ERR
	}

	// 初始化任务管理器
	if err = eyas_forage.InitJobMgr(); err != nil {
		goto ERR
	}

	// 正常退出
	for {
		time.Sleep(1 * time.Second)
	}

	return

ERR:
	fmt.Println(err)
}
