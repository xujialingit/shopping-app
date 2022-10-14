package model

import "errors"

type LoginResponse struct {
	Token interface{} `json:"token""`
}

const (
	RedisRefreshTokenKeypRefix = "sx:refresh"
	RedisBlackListKeyPrefix    = "sk:black_list"
)

type LoginResponseByRefreshToekn struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginResponseByBlackList struct {
	AccessToken string `json:"access_token"`
}

var (
	RefreshTokenExpired = errors.New("刷新token超时")
)
