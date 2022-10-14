package token

import (
	"github.com/golang-jwt/jwt"
	"net/url"
	"time"
)

/*
继续jwt封装的token生成器
*/

type Token interface {
	i()
	JwtSign(userId int64, userName string, expireDuration time.Duration) (tokenString string, err error)
	JwtParseUnsafe(tokenString string) (*claims, error)
	JwtParse(tokenString string) (*claims, error)
	JwtParseFromAuthorizationHeader(token string) (*claims, error)
	UrlSign(timestamp int64, path string, method string, params url.Values) (tokenString string, err error)
}

type token struct {
	secret string
}

type claims struct {
	UserID   int64
	UserName string
	jwt.StandardClaims
}

func New(secret string) Token {
	return &token{
		secret: secret,
	}
}

func (t *token) i() {

}
