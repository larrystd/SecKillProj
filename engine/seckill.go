package engine

import (
	"SecKill/api"
	"SecKill/conf"
	"SecKill/middleware/jwt"
	"SecKill/model"
	"encoding/gob"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

const SessionHeaderKey = "Authorization"

func SeckillEngine() *gin.Engine {
	router := gin.Default()

	config, err := conf.GetAppConfig()
	if err != nil {
		panic("failed to load redisService config" + err.Error())
	}
	store, _ := redis.NewStore(config.App.Redis.MaxIdle, config.App.Redis.Network,
		config.App.Redis.Address, config.App.Redis.Password, []byte("seckill"))
	router.Use(sessions.Sessions(SessionHeaderKey, store))
	gob.Register(&model.User{})

	userRouter := router.Group("/api/users")
	userRouter.POST("", api.RegisterUser)
	userRouter.Use(jwt.JWTAuth())
	{
		userRouter.PATCH("/:username/coupons/fetch/:name", api.FetchCoupon)
		userRouter.GET("/:username/coupons/list", api.ListCoupons)
		userRouter.POST("/:username/coupons/add", api.AddCoupon)
	}

	authRouter := router.Group("/api/auth")
	{
		authRouter.POST("", api.LoginAuth)
		authRouter.POST("/logout", api.Logout)
	}

	api.RunSecKillConsumer()

	return router
}
