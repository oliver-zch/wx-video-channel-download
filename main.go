package main

import (
	_ "embed"
	"fmt"
	"log"

	"wx_channels_web/internal/api"
	"wx_channels_web/internal/config"
)

//go:embed web/index.html
var indexHTML []byte

func main() {
	cfg := config.Load()
	srv := api.NewServer(cfg, indexHTML)

	addr := fmt.Sprintf("%s:%d", cfg.API.Hostname, cfg.API.Port)
	log.Printf("视频号下载服务启动: http://%s", addr)
	if err := srv.Run(addr); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
