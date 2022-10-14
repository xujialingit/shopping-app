package token

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"strings"
	"time"
)

/*
通过id和username生成token
*/

//生成token
func (t *token) JwtSign(userId int64, userName string, expireDuration time.Duration) (tokenString string, err error) {
	claimas := claims{
		UserID:   userId,
		UserName: userName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expireDuration).Unix(),
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),
		},
	}

	tokenString, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claimas).SignedString([]byte(t.secret))
	return
}

func (t *token) JwtParseUnsafe(tokenString string) (*claims, error) {
	tokenClaims, _, err := new(jwt.Parser).ParseUnverified(tokenString, &claims{})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*claims); ok {
			return claims, nil
		}
	}
	return nil, err
}

var (
	ErrorTokenCannotParse        = errors.New("token解密失败")
	ErrorTokenExpiredOrNotActive = errors.New("token过期或无效")
)

func (t *token) JwtParse(tokenString string) (*claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.secret), nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*claims); ok && tokenClaims.Valid {
			return claims, nil
		} else {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					return nil, ErrorTokenCannotParse
				} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					return nil, ErrorTokenExpiredOrNotActive
				} else {
					return nil, ErrorTokenCannotParse
				}
			} else {
				return nil, ErrorTokenCannotParse
			}
		}
	}
	return nil, ErrorTokenCannotParse
}

func (t *token) JwtParseFromAuthorizationHeader(token string) (*claims, error) {
	tokenString := stripBearerPrfixFromTokenString(token)
	return t.JwtParse(tokenString)
}

//切割生成的toekn前面的‘BEARER’
func stripBearerPrfixFromTokenString(tok string) string {
	if len(tok) > 6 && strings.ToUpper(tok[0:7]) == "BEARER" {
		return tok[7:]
	}
	return tok
}
