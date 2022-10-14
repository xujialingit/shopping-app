package login

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github/xujialingit/shopping-app/pkg/cache"
	"github/xujialingit/shopping-app/pkg/pkg/login/model"
	"github/xujialingit/shopping-app/pkg/pkg/token"
	"time"
)

type LoginTokenSystem interface {
	GenerateToken(ctx context.Context, userId int, userName string) (*model.LoginResponse, error)

	TokenCancel(ctx context.Context, token string) error

	ToeknCancelById(ctx context.Context, userId int, userName string) error

	RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, error)
}

func NewByRefreshToken(cfg *RefreshTokenConfig, repo cache.Repo) LoginTokenSystem {
	return &RefreshTokenSystem{cfg: cfg, cache: repo}
}

//刷新token配置
type RefreshTokenConfig struct {
	Secret          string        `json:"secret"`
	ExpireDuration  time.Duration `json:"expire_duration"`
	RefreshDuration time.Duration `json:"refresh_duration"`
}

type RefreshTokenSystem struct {
	cfg   *RefreshTokenConfig
	cache cache.Repo
}

//生成token
func (r *RefreshTokenSystem) GenerateToken(ctx context.Context, userId int, userName string) (*model.LoginResponse, error) {
	assessToken, err := token.New(r.cfg.Secret).JwtSign(int64(userId), userName, r.cfg.ExpireDuration)
	if err != nil {
		return nil, err
	}
	refreshToken := r.generateRefreshToken(r.cfg.Secret, userId, userName)
	userClaims := struct {
		UserId   int    `json:"user_id"`
		UserName string `json:"user_name"`
	}{
		UserId:   userId,
		UserName: userName,
	}

	userJson, _ := json.Marshal(userClaims)
	err = r.cache.Set(ctx, model.RedisRefreshTokenKeypRefix+refreshToken, string(userJson), r.cfg.RefreshDuration)

	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{Token: &model.LoginResponseByRefreshToekn{
		AccessToken:  assessToken,
		RefreshToken: refreshToken,
	}}, nil
}

func (r RefreshTokenSystem) TokenCancel(ctx context.Context, token string) error {
	_ = r.cache.Del(ctx, model.RedisRefreshTokenKeypRefix+token)
	return nil
}

func (r RefreshTokenSystem) ToeknCancelById(ctx context.Context, userId int, userName string) error {
	refreshToken := r.generateRefreshToken(r.cfg.Secret, userId, userName)
	return r.TokenCancel(ctx, refreshToken)
}

//刷新token
func (r RefreshTokenSystem) RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, error) {
	userJson, err := r.cache.Get(ctx, model.RedisRefreshTokenKeypRefix+refreshToken)

	if userJson == "" {
		return nil, errors.New("刷新Token无效")
	}
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, err
		}
	}

	userClaims := struct {
		UserId   int    `json:"user_id"`
		UserName string `json:"user_name"`
	}{}

	_ = json.Unmarshal([]byte(userJson), &userClaims)
	_ = r.TokenCancel(ctx, refreshToken)
	return r.GenerateToken(ctx, userClaims.UserId, userClaims.UserName)
}

//生成刷新token
func (r *RefreshTokenSystem) generateRefreshToken(secret string, userId int, userName string) string {
	hencrypt := hmac.New(md5.New, []byte(secret))
	hencrypt.Write([]byte(fmt.Sprintf("%v%d_%s", time.Now().Unix(), userId, userName)))
	return fmt.Sprintf("%x", hencrypt.Sum(nil))
}

type BlackListConfig struct {
	Secret         string        `json:"secret"`
	ExpireDuration time.Duration `json:"expire_duration"`
}

type BlackListSystem struct {
	cfg *BlackListConfig

	cache cache.Repo
}

func (r *BlackListSystem) blackListKey(userId int64, userName string) string {
	return fmt.Sprintf("%s_%d_%s", model.RedisBlackListKeyPrefix, userId, userName)
}

// CheckBlackList 验证时，需要验证是否在黑名单
func (r *BlackListSystem) CheckBlackList(ctx context.Context, accessToken string) (bool, error) {
	claim, err := token.New(r.cfg.Secret).JwtParseUnsafe(accessToken)
	if err != nil {
		return false, fmt.Errorf("this token is unvalid ")
	}
	key := r.blackListKey(claim.UserID, claim.UserName)

	a, err := r.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return a == "1", nil
}
