package api

import (
	"errors"
	"github/xujialingit/shopping-app/internal/config"
	"github/xujialingit/shopping-app/pkg/cache"
	"github/xujialingit/shopping-app/pkg/db"
	"go.uber.org/zap"
	"net/http"
)

//Server类包含了Server集成的插件
type Server struct {
	Logger     *zap.Logger //单例logger，全局只有一个logger，通过传递公用
	DB         db.Repo
	HttpServer *http.Server
	Cache      cache.Repo
}

func NewApiServer(logger *zap.Logger) (*Server, error) {
	if logger == nil {
		return nil, errors.New("logger参数不能为空")
	}
	s := &Server{}
	s.Logger = logger

	cfg := config.Get()

	dbRepo, err := db.New(&db.DBConfig{
		User:            cfg.Mysql.Base.User,
		Pass:            cfg.Mysql.Base.Pass,
		Addr:            cfg.Mysql.Base.Addr,
		Name:            cfg.Mysql.Base.Name,
		MaxOpenConn:     cfg.Mysql.Base.MaxOpenConn,
		MaxIdleConn:     cfg.Mysql.Base.MaxIdleConne,
		ConnMaxLifeTime: cfg.Mysql.Base.ConnMaxLifeTime,
		ServerName:      cfg.Server.ServerName,
	})

	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}
	s.DB = dbRepo

	//redis缓存
	cacheRepo, err := cache.New(cfg.Server.ServerName, &cache.RedisConf{
		Addr:         cfg.Redis.Addr,
		Pass:         cfg.Redis.Pass,
		Db:           cfg.Redis.Db,
		MaxRetries:   cfg.Redis.MaxRetries,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	if err != nil {
		logger.Fatal("redis服务开启失败！", zap.Error(err))
	}
	s.Cache = cacheRepo

	return s, nil
}
