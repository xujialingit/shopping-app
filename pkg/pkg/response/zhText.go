package response

var codeText = map[int]string{
	ServerError:        "服务器错误",
	TooManyRequests:    "请求发送过多",
	AuthorizationError: "token验证失败",
	ParamBindError:     "参数无效",
	TokenExpired:       "token过期",
	SendEamilCodeError: "发送验证码失败",
	ValidEmailCodeFail: "验证码错误",
	EmailIsExists:      "邮箱已被注册",
	CreateUserError:    "注册失败",
	UserNotRegistry:    "用户不存在",
	PwdError:           "账号密码错误",
	RefreshTokenError:  "刷新token失败",
	EmailCodeTypeError: "验证码类型错误",
	UserNotExits:       "用户不存在",
	ChangePWDFail:      "修改密码失败",
}
