//继续gin的封装
package core

import (
	"errors"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cors "github.com/rs/cors/wrapper/gin"
	swaggerFiles "github.com/swaggo/files"
	response2 "github/xujialingit/shopping-app/pkg/pkg/response"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"net/http"
	"runtime/debug"
	"time"

	ginSwagger "github.com/swaggo/gin-swagger"
)

//最大并发数量
const _MaxBurstSize = 100000

type Option func(*option)

type option struct {
	disablePProf      bool
	diableSwagger     bool
	disablePrometheus bool
	panicNotify       OnPanicNotify
	//recordMetrics     RecordMetrics
	enableCors bool //是否支持跨域
	enableRate bool
}

//发生panic时通知用
type OnPanicNotify func(ctx Context, err interface{}, stackInfo string)

//WithDisablePProf 禁用 pprof
func WithDisablePProf() Option {
	return func(opt *option) {
		opt.disablePProf = true
	}
}

//WithDisableSwagger 禁用Swagger
func WithDisableSwagger() Option {
	return func(opt *option) {
		opt.diableSwagger = true
	}
}

//WithDisablePrometheus 禁用prometheus
func WithDisablePrometheus() Option {
	return func(opt *option) {
		opt.disablePrometheus = true
	}
}

//WithPanicNotify 设置panic时的通知回调
func WithPanicNotify(notify OnPanicNotify) Option {
	return func(o *option) {
		o.panicNotify = notify
	}
}

//线路追踪相关
//func WithRecordMetrics(record RecordMetrics) Option {
//	return func(opt *option) {
//		opt.recordMetrics = record
//	}
//}

//开启跨域访问
func WithEnableCors() Option {
	return func(opt *option) {
		opt.enableCors = true
	}
}

func WithEnableRate() Option {
	return func(opt *option) {
		opt.enableRate = true
	}
}

//WarpAuthHandler 用来处理 Auth 的入口，在之后的handler中只需 ctx.UserID() ctx.UserName()即可
//handler 是真正的处理者
func WarpAuthHandler(handler func(Context) (userID int64, userName string, err response2.Error)) HandlerFunc {
	return func(ctx Context) {
		userID, userName, err := handler(ctx)

		if err != nil {
			ctx.AbortWithError(err)
			return
		}
		ctx.setUserID(userID)
		ctx.setUserName(userName)
	}
}

type RouteGroup interface {
	Group(string, ...HandlerFunc) RouteGroup
	Use(...HandlerFunc)
	IRoutes
}

var _ IRoutes = (*router)(nil)

type IRoutes interface {
	Any(string, ...HandlerFunc)
	GET(string, ...HandlerFunc)
	POST(string, ...HandlerFunc)
	DELETE(string, ...HandlerFunc)
	PATCH(string, ...HandlerFunc)
	PUT(string, ...HandlerFunc)
	OPTIONS(string, ...HandlerFunc)
	HEAD(string, ...HandlerFunc)
}

type router struct {
	group *gin.RouterGroup
}

func (r *router) Group(s string, handlerFunc ...HandlerFunc) RouteGroup {
	group := r.group.Group(s, warpHandlers(handlerFunc...)...)
	return &router{
		group: group,
	}
}

func (r *router) Use(handlerFunc ...HandlerFunc) {
	r.group.Use(warpHandlers(handlerFunc...)...)
}

func (r *router) Any(s string, handlerFunc ...HandlerFunc) {
	r.group.Any(s, warpHandlers(handlerFunc...)...)
}

func (r *router) GET(s string, handlerFunc ...HandlerFunc) {
	r.group.GET(s, warpHandlers(handlerFunc...)...)
}

func (r *router) POST(s string, handlerFunc ...HandlerFunc) {
	r.group.POST(s, warpHandlers(handlerFunc...)...)
}

func (r *router) DELETE(s string, handlerFunc ...HandlerFunc) {
	r.group.DELETE(s, warpHandlers(handlerFunc...)...)
}

func (r *router) PATCH(s string, handlerFunc ...HandlerFunc) {
	r.group.PATCH(s, warpHandlers(handlerFunc...)...)
}

func (r *router) PUT(s string, handlerFunc ...HandlerFunc) {
	r.group.PUT(s, warpHandlers(handlerFunc...)...)
}

func (r *router) OPTIONS(s string, handlerFunc ...HandlerFunc) {
	r.group.OPTIONS(s, warpHandlers(handlerFunc...)...)
}

func (r *router) HEAD(s string, handlerFunc ...HandlerFunc) {
	r.group.HEAD(s, warpHandlers(handlerFunc...)...)
}

//把自己定义handler 在gin.HandlerFunc中调用
func warpHandlers(handlers ...HandlerFunc) []gin.HandlerFunc {
	funcs := make([]gin.HandlerFunc, len(handlers))

	for i, handler := range handlers {
		hd := handler
		funcs[i] = func(c *gin.Context) {
			ctx := newContext(c)
			defer releaseContext(ctx)
			hd(ctx)
		}
	}
	return funcs
}

//封装一层 gin.Engin
type Engine interface {
	http.Handler
	Group(relativePath string) RouteGroup
}

type engine struct {
	e         *gin.Engine
	baseGroup *gin.RouterGroup //全局basePath
}

func (engin *engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	engin.e.ServeHTTP(writer, request)
}

func (engin *engine) Group(relativePath string) RouteGroup {
	return &router{
		group: engin.baseGroup.Group(relativePath),
	}
}

func New(serverName string, logger *zap.Logger, options ...Option) (Engine, error) {
	if logger == nil {
		return nil, errors.New("初始话Engine时logger为必要参数")
	}

	gin.SetMode(gin.DebugMode)
	mux := &engine{
		e: gin.New(),
	}
	//全部url以 serverName开头： /serverName/test
	basePath := "/" + serverName
	mux.baseGroup = mux.e.Group(basePath)

	//读取配置
	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	//????
	if !opt.disablePProf {
		pprof.RouteRegister(mux.baseGroup)
	}

	//swagger 注册swagger
	if !opt.diableSwagger {
		mux.baseGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	//????
	if !opt.disablePrometheus {
		mux.baseGroup.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	//跨域
	if opt.enableCors {
		mux.baseGroup.Use(cors.New(cors.Options{
			AllowedOrigins: []string{"*"}, //跨域的域名
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:     []string{"*"},
			AllowCredentials:   true,
			OptionsPassthrough: true,
		}))
	}

	//注册全局 Recover and logger
	mux.baseGroup.Use(func(ctx *gin.Context) {
		c := newContext(ctx)
		defer releaseContext(c)

		//注入logger到ctx
		c.setLogger(logger)

		defer func() {
			if err := recover(); err != nil {
				stackInfo := string(debug.Stack())
				logger.Error("got panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", stackInfo))

				if notify := opt.panicNotify; notify != nil {
					//notify 中不能再panic错误
					notify(c, err, stackInfo)
				}
			}
		}()

		ctx.Next()
	})

	//???
	if opt.enableRate {
		limiter := rate.NewLimiter(rate.Every(time.Second*1), _MaxBurstSize)
		mux.baseGroup.Use(func(ctx *gin.Context) {
			context := newContext(ctx)

			defer releaseContext(context)

			if !limiter.Allow() {
				context.AbortWithError(response2.NewErrorAutoMsg(
					http.StatusTooManyRequests,
					response2.TooManyRequests,
				))
				return
			}

			ctx.Next()
		})
	}

	system := mux.Group("/system")
	{
		//健康检查
		system.GET("/health", func(ctx Context) {
			resp := &struct {
				Timestamp time.Time `json:"timestamp"`
				Host      string    `json:"host"`
				Status    string    `json:"status"`
			}{
				Timestamp: time.Now(),
				Host:      ctx.RequestContext().Request.Host,
				Status:    "ok",
			}
			ctx.Payload(resp)
		})
	}

	// 注册全局 Telemetry
	//openTelemetry := NewOpenTelemetry(opt.recordMetrics)
	//mux.baseGroup.Use(func(c *gin.Context) {
	//	ctx := newContext(c)
	//	defer releaseContext(ctx)
	//
	//	openTelemetry.Telemetry(ctx, serverName)
	//})
	return mux, nil
}
