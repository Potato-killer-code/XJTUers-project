package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"smart-cabinet/internal/config"
	"smart-cabinet/internal/db"
	"smart-cabinet/internal/handler"
	"smart-cabinet/internal/mcu"
	"smart-cabinet/internal/service"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("===== 智能外卖柜后端服务 =====")

	// 1. 加载配置
	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	log.Printf("配置加载成功: HTTP=%s, MCU=%s", cfg.Server.Addr(), cfg.MCU.Addr())

	// 2. 连接数据库
	database, err := db.New(cfg.Database)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close()
	log.Println("数据库连接成功")

	// 3. 启动 MCU TCP 服务
	mcuServer := mcu.NewServer(cfg.MCU)
	if err := mcuServer.Start(); err != nil {
		log.Fatalf("MCU 服务启动失败: %v", err)
	}
	defer mcuServer.Stop()

	// 4. 初始化业务服务
	svc := service.NewCabinetService(database, mcuServer)

	// 5. 注册 HTTP 路由
	mux := http.NewServeMux()
	hdl := handler.NewCabinetHandler(svc)
	hdl.RegisterRoutes(mux)

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// 静态文件服务（Vue 前端构建产物）
	staticDir := cfg.Server.StaticDir
	if staticDir == "" {
		staticDir = "web/dist"
	}
	fs := http.FileServer(http.Dir(staticDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// API 路径不走静态文件
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			http.NotFound(w, r)
			return
		}
		// 检查静态文件是否存在，不存在则 fallback 到 index.html（SPA 路由）
		path := staticDir + r.URL.Path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, staticDir+"/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	// 6. 启动 HTTP 服务
	httpServer := &http.Server{
		Addr:    cfg.Server.Addr(),
		Handler: mux,
	}

	go func() {
		log.Printf("HTTP 服务启动: http://%s", cfg.Server.Addr())
		log.Printf("  API 文档: api/api.md")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 服务异常: %v", err)
		}
	}()

	// 7. 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务...")
	httpServer.Close()
	mcuServer.Stop()
	log.Println("服务已安全退出")
}
