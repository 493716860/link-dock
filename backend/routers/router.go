package routers

import (
	"linkdock/controllers"

	"github.com/beego/beego/v2/server/web"
)

func init() {
	ns := web.NewNamespace("/api",
		web.NSRouter("/auth/login", &controllers.AuthController{}, "post:Login"),
		web.NSRouter("/auth/logout", &controllers.AuthController{}, "post:Logout"),
		web.NSRouter("/auth/register", &controllers.AuthController{}, "post:Register"),
		web.NSRouter("/auth/me", &controllers.AuthController{}, "get:Me"),
		web.NSRouter("/auth/check-username", &controllers.AuthController{}, "post:CheckUsername"),
		web.NSRouter("/auth/recovery-question", &controllers.AuthController{}, "post:GetRecoveryQuestion"),
		web.NSRouter("/auth/reset-password", &controllers.AuthController{}, "post:ResetPassword"),
		web.NSRouter("/auth/change-password", &controllers.AuthController{}, "post:ChangePassword"),

		web.NSRouter("/admin/seed", &controllers.DataController{}, "get:GetSeedData;put:SaveSeedData"),

		web.NSRouter("/public/data", &controllers.DataController{}, "get:GetPublicData"),
		web.NSRouter("/user/data", &controllers.DataController{}, "get:GetUserData"),

		web.NSRouter("/sites/fetch-metadata", &controllers.DataController{}, "post:FetchMetadata"),
		web.NSRouter("/categories", &controllers.DataController{}, "post:CreateCategory"),
		web.NSRouter("/categories/reorder", &controllers.DataController{}, "post:UpdateCategoriesOrder"),
		web.NSRouter("/categories/:id", &controllers.DataController{}, "put:UpdateCategory;delete:DeleteCategory"),
		web.NSRouter("/sites", &controllers.DataController{}, "post:CreateSite"),
		web.NSRouter("/sites/reorder", &controllers.DataController{}, "post:UpdateSitesOrder"),
		web.NSRouter("/sites/:id/visit", &controllers.DataController{}, "post:TrackSiteVisit"),
		web.NSRouter("/sites/:id/state", &controllers.DataController{}, "post:UpdateSiteState"),
		web.NSRouter("/sites/:id", &controllers.DataController{}, "put:UpdateSite;delete:DeleteSite"),
	)
	web.AddNamespace(ns)
}
