package main

import (
	"flag"
	"fmt"

	"qianhao-backend/internal/config"
	"qianhao-backend/internal/handler"
	"qianhao-backend/internal/svc"

	cronJobs "qianhao-backend/internal/cron" //  引入 cron 包

	"github.com/robfig/cron/v3" // 引入库
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/qianhao-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	// 新增：启动定时任务 ---
	cRunner := cron.New(cron.WithSeconds()) // 支持秒级控制
	// 初始化 Job
	expireJob := cronJobs.NewExpireOrderJob(ctx)

	// 添加任务：每 5 分钟执行一次 (你可以根据需求调整)
	// Cron 表达式: "秒 分 时 日 月 周"
	// "0 */5 * * * *" 代表每5分钟
	// 为了测试方便，你可以先改成 "*/30 * * * * *" (每30秒一次)
	cRunner.AddFunc("0 */10 * * * *", expireJob.Run)

	cRunner.Start()
	defer cRunner.Stop()

	// 👇 2. 插入这段全局错误处理代码
	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		return 200, map[string]interface{}{
			"code": 500, // 或者你可以定义具体的错误码
			"msg":  err.Error(),
			"data": nil,
		}
	})
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
