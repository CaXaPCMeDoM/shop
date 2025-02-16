package api

import (
	"avito-tech-winter-2025/internal/config"
	"avito-tech-winter-2025/internal/http/auth"
	"avito-tech-winter-2025/internal/http/handler"
	"avito-tech-winter-2025/internal/middleware"
	"avito-tech-winter-2025/internal/storage/postgres"
	"avito-tech-winter-2025/pkg/hash"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	DB       *postgres.Storage
	TokenMgr *auth.Manager
	Hasher   *hash.SHA1
	Cfg      *config.Config
}

func SetupRouter(deps *Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	authHandler := handler.NewAuthHandler(deps.DB, deps.TokenMgr, deps.Hasher, deps.Cfg)
	mainHandler := handler.NewHandler(deps.DB)

	router.POST("/api/auth", authHandler.Login)

	authGroup := router.Group("/api")
	authGroup.Use(middleware.AuthMiddleware(deps.TokenMgr))

	authGroup.GET("/info", mainHandler.GetInfo)
	authGroup.POST("/sendCoin", mainHandler.SendCoin)
	authGroup.GET("/buy/:item", mainHandler.BuyItem)

	return router
}
