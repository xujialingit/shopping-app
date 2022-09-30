//封装Http响应
package response

//常用的返回结构
type JsonResponse struct {
	Code    int         `json:"code"` //状态码
	Message string      `json:"msg"`  //描述信息
	Data    interface{} `json:"data"` //返回数据
}

func NewResponse(payload ...interface{}) *JsonResponse {
	responseData := interface{}(nil)
	if len(payload) > 0 && payload[0] != nil {
		responseData = payload[0]
	} else {
		responseData = make(map[string]interface{})
	}

	return &JsonResponse{
		Code:    0,
		Message: "成功",
		Data:    responseData,
	}
}

//错误码定义
const (
	ServerError        = 10001
	TooManyRequests    = 10002
	AuthorizationError = 10003
	ParamBindError     = 10004
	TokenExpired       = 10005

	SendEamilCodeError = 10006
	ValidEmailCodeFail = 10007
	EmailIsExists      = 10008
	CreateUserError    = 10009

	UserNotRegistry    = 10010
	PwdError           = 10011
	RefreshTokenError  = 10012
	EmailCodeTypeError = 10013
	UserNotExits       = 10014
	ChangePWDFail      = 10015
)

func Text(code int) string {
	return codeText[code]
}
