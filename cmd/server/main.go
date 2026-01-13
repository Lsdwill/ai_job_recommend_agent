package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"qd-sc/internal/api/handler"
	"qd-sc/internal/api/middleware"
	"qd-sc/internal/client"
	"qd-sc/internal/config"
	"qd-sc/internal/service"

	"github.com/gin-gonic/gin"
)

var (
	serverWg sync.WaitGroup
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		if _, err2 := os.Stat("config.local.yaml"); err2 == nil {
			cfg, err = config.Load("config.local.yaml")
		}
		if err != nil {
			log.Fatalf("加载配置失败: %v", err)
		}
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Printf("性能配置: GOMAXPROCS=%d, MaxGoroutines=%d, PoolSize=%d, QueueSize=%d",
		runtime.NumCPU(),
		cfg.Performance.MaxGoroutines,
		cfg.Performance.GoroutinePoolSize,
		cfg.Performance.TaskQueueSize)

	llmClient := client.NewLLMClient(cfg)
	amapClient := client.NewAmapClient(cfg)
	jobClient := client.NewJobClient(cfg)
	ocrClient := client.NewOCRClient(cfg)

	locationService := service.NewLocationService(cfg, amapClient)
	jobService := service.NewJobService(cfg, jobClient)
	policyService := service.NewPolicyService(cfg)
	chatService := service.NewChatService(cfg, llmClient, ocrClient, locationService, jobService, policyService)

	chatHandler := handler.NewChatHandler(chatService)
	healthHandler := handler.NewHealthHandler()
	metricsHandler := handler.NewMetricsHandler()

	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(200, 50))
	if cfg.Performance.EnableMetrics == nil || *cfg.Performance.EnableMetrics {
		router.Use(middleware.Metrics())
	}

	router.GET("/health", healthHandler.Check)
	if cfg.Performance.EnableMetrics == nil || *cfg.Performance.EnableMetrics {
		router.GET("/metrics", metricsHandler.GetMetrics)
	}
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": cfg.City.SystemName + " API",
			"version": "1.0.0",
			"endpoints": []string{
				"POST /v1/chat/completions",
				"GET /health",
				"GET /metrics (性能指标)",
				"GET /debug/pprof/* (性能分析)",
			},
		})
	})

	if cfg.Performance.EnablePprof == nil || *cfg.Performance.EnablePprof {
		debugGroup := router.Group("/debug/pprof")
		{
			debugGroup.GET("/", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			debugGroup.GET("/:name", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			debugGroup.POST("/:name", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		}
	}

	v1 := router.Group("/v1")
	{
		v1.POST("/chat/completions", chatHandler.ChatCompletions)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
		ReadHeaderTimeout: 10 * time.Second,
	}

	serverWg.Add(1)
	go func() {
		defer serverWg.Done()
		log.Printf("服务器启动在 %s", addr)
		log.Printf("OpenAI兼容端点: POST http://%s/v1/chat/completions", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	} else {
		log.Println("HTTP服务器已优雅关闭")
	}

	log.Println("清理资源完成")

	done := make(chan struct{})
	go func() {
		serverWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("所有goroutine已完成")
	case <-time.After(5 * time.Second):
		log.Println("等待超时，强制退出")
	}

	log.Println("服务器已完全关闭")
}
