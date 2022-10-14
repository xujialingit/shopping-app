package token

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func (t *token) UrlSign(timestamp int64, path string, method string, params url.Values) (tokenString string, err error) {
	//合法的 methods
	methods := map[string]bool{
		"get":     true,
		"post":    true,
		"put":     true,
		"path":    true,
		"delete":  true,
		"head":    true,
		"options": true,
	}

	methodName := strings.ToLower(method)
	if !methods[methodName] {
		err = errors.New("UrlSign: method参数错误")
		return
	}

	//自带的sortBy key
	sortParamsEncode := params.Encode()

	//加密规则：path + method + sortParamsEncode + secret
	encryptStr := fmt.Sprintf("%s%s%s%d%s", path, methodName, sortParamsEncode, timestamp, t.secret)
	s := md5.New()
	s.Write([]byte(encryptStr))
	md5Str := hex.EncodeToString(s.Sum(nil))

	tokenString = md5Str
	return
}
