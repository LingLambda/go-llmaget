package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"llmaget/config"
	"llmaget/handlers"
	"llmaget/services"
)

func main() {
	log.Println("🚀 FF14 石之家服务启动...")

	// 加载配置
	state := config.GetState()
	state.Load()

	// 创建服务
	ff14Svc := services.NewFF14Service()

	// 首次执行数据获取
	go func() {
		if err := ff14Svc.SaveMyBaseInfo(); err != nil {
			log.Printf("⚠️ 首次数据获取失败: %v", err)
		}
		// 启动定时任务
		startScheduler(ff14Svc)
	}()

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 引擎
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return log.Prefix() + param.TimeStamp.Format("2006/01/02 15:04:05") +
				" | " + param.Method +
				" | " + param.Path +
				" | " + param.StatusCodeColor() +
				param.ResetColor() + "\n"
		},
	}))

	// CORS 中间件
	r.Use(corsMiddleware())

	// 注册路由
	handler := handlers.NewHandler(ff14Svc)
	handler.RegisterRoutes(r)

	// 打印启动信息
	log.Printf("🌐 HTTP服务器启动在 %s", config.ServerPort)

	// 启动服务器
	if err := r.Run(config.ServerPort); err != nil {
		log.Fatalf("❌ HTTP服务器启动失败: %v", err)
	}
}

// startScheduler 启动定时任务
func startScheduler(ff14Svc *services.FF14Service) {
	// 定时获取基础信息
	go func() {
		ticker := time.NewTicker(config.FetchInterval)
		defer ticker.Stop()

		log.Printf("⏰ 基础信息定时任务启动，每 %v 执行一次", config.FetchInterval)

		for range ticker.C {
			log.Println("⏰ 获取基础信息任务触发...")
			if err := ff14Svc.SaveMyBaseInfo(); err != nil {
				log.Printf("❌ 获取基础信息失败: %v", err)
			}
		}
	}()

	// 定时签到
	go func() {
		ticker := time.NewTicker(config.SignInterval)
		defer ticker.Stop()

		log.Printf("⏰ 每日签到任务启动，每 %v 执行一次", config.SignInterval)

		for range ticker.C {
			log.Println("⏰ 签到任务触发...")
			if resp, err := ff14Svc.SignAndGetSignReward(); err != nil {
				log.Printf("❌ 签到并领取奖励失败: %v, %s", err, string(resp))
			}
		}
	}()

}

// corsMiddleware CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
