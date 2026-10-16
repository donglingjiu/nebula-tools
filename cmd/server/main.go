package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"nebula-tools/internal/api"
	"nebula-tools/internal/buildinfo"
	"nebula-tools/internal/db"
)

func main() {
	port := flag.Int("port", 0, "listen port")
	flag.Parse()

	addr := ":8080"
	if *port > 0 {
		addr = fmt.Sprintf(":%d", *port)
	} else if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	if err := db.Init(); err != nil {
		log.Fatal("[nebula] 数据库初始化失败: ", err)
	}
	api.BootstrapAdmin()

	if s := os.Getenv("NEBULA_JWT_SECRET"); s != "" {
		api.SetJWTSecret([]byte(s))
	}

	b := buildinfo.Get()
	log.Printf("[nebula] %s | Go(build)=%s | bundled=%v", b.Version, b.Go, b.Bundled)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), api.Auth(), api.Usage())
	api.Router(r)
	api.StartAlarm()

	log.Printf("[nebula] 服务已启动 → http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
