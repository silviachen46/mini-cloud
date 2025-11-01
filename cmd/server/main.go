package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/silviachen46/mini-cloud/internal/api"
	"github.com/silviachen46/mini-cloud/internal/meta"
	"github.com/silviachen46/mini-cloud/internal/storage"
)

func main() {
	// 1) 依赖初始化
	db, err := meta.Open("mini.db") // 简化 DSN，更通用
	if err != nil { log.Fatal("meta open:", err) }
	if err := meta.Migrate(db); err != nil { log.Fatal("meta migrate:", err) }

	stor := storage.NewFS("./data")
	if err := stor.Init(); err != nil { log.Fatal("storage init:", err) }

	// 2) 路由
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 注册对象 API（PUT/GET/DELETE/HEAD）
	api.RegisterBasic(r, db, stor)

	// 打印已注册路由
	for _, rt := range r.Routes() {
		log.Printf("[route] %-6s %s\n", rt.Method, rt.Path)
	}

	// 3) 启动
	log.Println("listening on :8080")
	if err := r.Run(":8080"); err != nil { log.Fatal("gin run error:", err) }
}
