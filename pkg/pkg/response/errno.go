package response

import (
	"encoding/json"
	"net/http"
)

var _ Error = (*err)(nil)

type Error interface {
	error
	WithErr(err error) Error
	GetBusinessCode() int
	GetHttpCode() int
	GetMsg() string
	GetErr() error
	ToString() string
}

type err struct {
	HttpCode     int
	BusinessCode int
	Message      string
	Err          error
}

func NewError(httpcode, businessCode int, msg string) Error {
	return &err{
		HttpCode:     httpcode,
		BusinessCode: businessCode,
		Message:      msg,
	}
}

//新建 httpCode 为statusOk的err
func NewErrorWithStatusOk(businessCode int, msg string) Error {
	return &err{
		HttpCode:     http.StatusOK,
		BusinessCode: businessCode,
		Message:      msg,
	}
}

//通过状态码自动去找msg
func NewErrorWithStatusOkAutoMsg(businessCode int) Error {
	return &err{
		HttpCode:     http.StatusOK,
		BusinessCode: businessCode,
		Message:      Text(businessCode),
	}
}

func NewErrorAutoMsg(httpCode, businessCode int) Error {
	return &err{
		HttpCode:     httpCode,
		BusinessCode: businessCode,
		Message:      Text(businessCode),
	}
}

func (e err) Error() string {
	return e.ToString()
}

func (e err) WithErr(err error) Error {
	e.Err = err
	return e
}

func (e err) GetBusinessCode() int {
	return e.BusinessCode
}

func (e err) GetHttpCode() int {
	return e.HttpCode
}

func (e err) GetMsg() string {
	return e.Message
}

func (e err) GetErr() error {
	return e.Err
}

//返回json格式的错误
func (e err) ToString() string {
	err := &struct {
		HttpCode     int    `json:"http_code"`
		BusinessCode int    `json:"business_code"`
		Message      string `json:"msg"`
	}{
		HttpCode:     e.HttpCode,
		BusinessCode: e.BusinessCode,
		Message:      e.Message,
	}

	raw, _ := json.Marshal(err)
	return string(raw)
}
