package main

import (
	"log"

	"github.com/Jennifer-ra/accounting/services/go-api/internal/config"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/database"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/handler"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/middleware"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
		if cfg.SessionSecret == "change-me-in-production" {
			log.Fatal("SESSION_TOKEN_SECRET must be changed in production")
		}
		if cfg.WeChatMockLogin {
			log.Fatal("WECHAT_MOCK_LOGIN must be disabled in production")
		}
	}

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	authSvc := service.NewAuthService(cfg, db)
	appSvc := service.NewAccountingService(db)
	api := handler.New(authSvc, appSvc)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.CORS(cfg))
	api.Register(router)

	log.Printf("accounting api listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
