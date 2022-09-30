//基于gin的context封装
package core

import (
	"bytes"
	stdContext "context"
	"github.com/gin-gonic/gin/binding"
	"github.com/spf13/cast"
	"io/ioutil"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type HandlerFunc func(c Context)

//内存缓存
const (
	_Logger     = "_looger_"
	_Response   = "_response_"
	_UserId     = "_user_id_"
	_UserName   = "_user_name_"
	_DisableLog = "_disable_log_"
)

var contextPool = &sync.Pool{
	New: func() interface{} {
		return new(context)
	},
}

//创建一个context
func newContext(ctx *gin.Context) Context {
	context := contextPool.Get().(*context)
	context.ctx = ctx
	return context
}

//释放context
func releaseContext(ctx Context) {
	c := ctx.(*context)
	c.ctx = nil
	contextPool.Put(c)
}

var _ Context = (*context)(nil)

//Context 封装gin.Context
type Context interface {
	//ShouldBindForm 同时反序列化 querystring 和 postForm
	//当querystring 和 postform存在相同字段时，postForm优先
	//tag: `form:”xxx“`
	ShouldBindForm(obj interface{}) error

	//反序列化 postJson
	//tag: `json:"xxx"`
	ShouldBindJSON(obj interface{}) error

	//shouldBindURL 反序列化 path参数(/user/:name)
	//tag `url:"xxx"`
	ShouldBindURL(obj interface{}) error

	//Header 获取Header对象
	Header() http.Header

	//GetHeader 通过key获取Header
	GetHeader(key string) string

	//SetHeader 设置Header
	SetHeader(key, value string)

	//URL unescape后的url
	URI() string

	//RequestData 获取请求体(可多次读取)
	RequestData() []byte

	//Redirect 重定向
	Rediect(code int, location string)

	//Payload 正确返回
	Payload(payload interface{})
	getResponse() interface{}

	//Html 返回界面
	HTML(name string, obj interface{})

	//AbortWithError 错误返回
	AbortWithError(err error)

	//Logger 获取Logger对象
	Logger() *zap.Logger
	setLogger(logger *zap.Logger)

	DisableLog(flag bool)
	getDisableLog() bool

	//UserId() 获取 UserID
	UserID() int64
	setUserID(userID int64)

	//UserName 获取 UserName
	UserName() string
	setUserName(userName string)

	//RequestContext 获取Gin的context
	RequestContext() *gin.Context

	//SvcContext 给下层用的context
	SvcContext() SvcContext
}

type context struct {
	ctx *gin.Context
}

//实现接口
func (c *context) ShouldBindForm(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.Form)
}

func (c *context) ShouldBindJSON(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.JSON)
}

func (c *context) ShouldBindURL(obj interface{}) error {
	return c.ctx.ShouldBindUri(obj)
}

func (c *context) Header() http.Header {
	header := c.ctx.Request.Header
	clone := make(http.Header, len(header))
	for k, v := range header {
		value := make([]string, len(v))
		copy(value, v)
		clone[k] = value
	}
	return clone
}

func (c *context) GetHeader(key string) string {
	return c.ctx.GetHeader(key)
}

func (c *context) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *context) URI() string {
	uri, _ := url.QueryUnescape(c.ctx.Request.URL.RequestURI())
	return uri
}

func (c *context) RequestData() []byte {
	rawData, _ := c.ctx.GetRawData()
	c.ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(rawData))
	return rawData
}

func (c *context) Rediect(code int, location string) {
	c.ctx.Redirect(code, location)
}

func (c *context) Payload(payload interface{}) {
	resp := response2.NewResponse(payload)

	if _, exist := c.ctx.Get(_Response); exist {
		return
	}
	c.ctx.JSON(http.StatusOK, resp)
	c.ctx.Set(_Response, resp)
}

func (c *context) getResponse() interface{} {
	if resp, ok := c.ctx.Get(_Response); ok != false {
		return resp
	}
	return nil
}

func (c *context) HTML(name string, obj interface{}) {
	c.ctx.HTML(200, name+".html", obj)
}

func (c *context) AbortWithError(err error) {
	if err != nil {
		errResp := response2.NewErrorAutoMsg(http.StatusInternalServerError, response2.ServerError)
		if v, ok := err.(response2.Error); ok {
			errResp = v
		} else {
			_ = errResp.WithErr(err)
		}

		//打印日志
		if errResp.GetErr() != nil {
			c.Logger().Error("服务错误...", zap.Error(errResp.GetErr()))
		}

		httpCode := errResp.GetHttpCode()
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}

		resp := response2.NewResponse()
		resp.Code = errResp.GetBusinessCode()
		resp.Message = errResp.GetMsg()

		c.ctx.AbortWithStatus(httpCode)
		c.ctx.Set(_Response, resp)
		c.ctx.JSON(httpCode, resp)
	}
}

func (c *context) Logger() *zap.Logger {
	logger, ok := c.ctx.Get(_Logger)
	if !ok {
		return nil
	}
	return logger.(*zap.Logger)
}

func (c *context) setLogger(logger *zap.Logger) {
	c.ctx.Set(_Logger, logger)
}

func (c *context) DisableLog(flag bool) {
	c.ctx.Set(_DisableLog, flag)
}

func (c *context) getDisableLog() bool {
	val, ok := c.ctx.Get(_DisableLog)
	if !ok {
		return false
	}
	return val.(bool)
}

func (c context) UserID() int64 {
	val, ok := c.ctx.Get(_UserId)
	if !ok {
		return 0
	}
	return val.(int64)
}

func (c context) setUserID(userID int64) {
	c.ctx.Set(_UserId, userID)
}

func (c context) UserName() string {
	val, ok := c.ctx.Get(_UserName)
	if !ok {
		return ""
	}
	return val.(string)
}

func (c context) setUserName(userName string) {
	c.ctx.Set(_UserName, userName)
}

func (c context) RequestContext() *gin.Context {
	return c.ctx
}

func (c context) SvcContext() SvcContext {
	ctx := c.RequestContext().Request.Context()

	//用户信息设置进去
	ctx = stdContext.WithValue(ctx, _UserId, c.UserID())
	ctx = stdContext.WithValue(ctx, _UserName, c.UserName())

	return &svcContext{
		ctx:    ctx,
		logger: c.Logger(),
	}
}

//svcContext 传给下层用的Context,精简去掉Request、Response等信息
//只保留一下信息
type SvcContext interface {
	UserID() int64
	UserName() string
	Context() stdContext.Context
	Logger() *zap.Logger
}

type svcContext struct {
	ctx    stdContext.Context
	logger *zap.Logger
}

func (s *svcContext) UserID() int64 {
	return cast.ToInt64(s.ctx.Value(_UserId))
}

func (s svcContext) UserName() string {
	return cast.ToString(s.ctx.Value(_UserName))
}

func (s svcContext) Context() stdContext.Context {
	return s.ctx
}

func (s svcContext) Logger() *zap.Logger {
	return s.logger
}
