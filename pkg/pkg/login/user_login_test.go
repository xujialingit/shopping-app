package login

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github/xujialingit/shopping-app/pkg/cache"
	"github/xujialingit/shopping-app/pkg/pkg/login/model"
	"github/xujialingit/shopping-app/pkg/pkg/token"
	"testing"
	"time"
)

func TestRefreshTokenSystem_GenerateToken(t *testing.T) {
	cacheRepo, err := cache.New("test", &cache.RedisConf{
		Addr: "139.155.76.9:6379",
		Pass: "000415",
	})
	assert.NoError(t, err)

	cfg := &RefreshTokenConfig{
		Secret:          "test_secret",
		ExpireDuration:  time.Second * 2,
		RefreshDuration: time.Second * 10,
	}

	ctx := context.Background()
	userId := 1
	userName := "name1"
	system := NewByRefreshToken(cfg, cacheRepo)

	t.Run("测试有效时间", func(t *testing.T) {
		resp, err := system.GenerateToken(ctx, userId, userName)
		assert.NoError(t, err)
		refreshResp := resp.Token.(*model.LoginResponseByRefreshToekn)
		t.Logf("生成的token:%s", refreshResp.AccessToken)
		t.Logf("生成的刷新token:%s", refreshResp.RefreshToken)

		claims, err := token.New(cfg.Secret).JwtParse(refreshResp.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, claims.UserID, int64(userId))
		assert.Equal(t, claims.UserName, userName)

		time.Sleep(time.Second * 3)
		_, err = token.New(cfg.Secret).JwtParse(refreshResp.AccessToken)
		assert.Equal(t, token.ErrorTokenExpiredOrNotActive, err)
	})

	t.Run("测试刷新token", func(t *testing.T) {
		resp, err := system.GenerateToken(ctx, userId, userName)
		assert.NoError(t, err)

		refreshResp := resp.Token.(*model.LoginResponseByRefreshToekn)
		time.Sleep(time.Second * 3)

		newResp, err := system.RefreshToken(ctx, refreshResp.RefreshToken)
		assert.NoError(t, err)
		newRefreshToken := newResp.Token.(*model.LoginResponseByRefreshToekn)

		claims, err := token.New(cfg.Secret).JwtParse(newRefreshToken.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, claims.UserID, int64(userId))
		assert.Equal(t, claims.UserName, userName)
	})
}
