package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/filter/cors"

	"linkdock/models"
	_ "linkdock/routers"
)

func main() {
	fmt.Println(">>> LinkDock Backend is starting...")
	// 自动建表 sync db
	orm.RunSyncdb("default", false, true)
	models.EnsureSiteSortOrderColumn()
	models.EnsureUserRecoveryColumns()
	models.EnsureSiteEnhancementColumns()
	fmt.Println(">>> Database sync completed.")

	// 配置静态资源路径，将前端 build 后的 dist 目录内容放在 backend/static 下
	web.SetStaticPath("/", "static")

	// SPA 路由支持：如果访问的不是 /api 开头的路径且文件不存在，则返回 index.html
	web.InsertFilter("/*", web.BeforeRouter, func(ctx *context.Context) {
		if strings.HasPrefix(ctx.Input.URL(), "/api") {
			return
		}
		// 如果路径包含点（可能是 .js, .css, .png 等静态文件），由 SetStaticPath 处理
		if strings.Contains(ctx.Input.URL(), ".") {
			return
		}
		// 其他所有路由（如 /login, /categories 等前端路由）转发到 index.html
		http.ServeFile(ctx.ResponseWriter, ctx.Request, "static/index.html")
	})

	// 加入跨域过滤器，按照标准实践配置
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"}, // 允许所有来源，如果是生产环境建议配置具体域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	}))

	port, _ := web.AppConfig.Int("httpport")
	fmt.Printf(">>> LinkDock Backend is running on port %d...\n", port)
	web.Run()
}
