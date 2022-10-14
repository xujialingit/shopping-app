package token

import (
	"net/url"
	"testing"
	"time"
)

const secret = "i1ydX9RtHyuJTrw7frcu"

//通过用户ID和用户名生成
func BenchmarkNew(b *testing.B) {
	b.ResetTimer()
	token := New(secret)
	for i := 0; i < b.N; i++ {
		tokenStrng, _ := token.JwtSign(123456789, "xujilainTest", 24*time.Hour)
		token.JwtParse(tokenStrng)
	}
}

func TestToken_UrlSign(t *testing.T) {
	urlPath := "/echo"
	method := "post"
	params := url.Values{}
	params.Add("a", "a1")
	params.Add("b", "b1")
	params.Add("c", "c1")

	tokenString, err := New(secret).UrlSign(time.Now().Unix(), urlPath, method, params)
	if err != nil {
		t.Error("s失败", err)
		return
	}
	t.Log(tokenString)
}

func TestToken_JwtSign(t *testing.T) {
	tokenString, _ := New(secret).JwtSign(123, "test_user", 2*time.Second)
	t.Log(tokenString)

	time.Sleep(3 * time.Second)
	user, err := New(secret).JwtParse(tokenString)
	if err != nil {
		t.Error(err)
	}
	t.Log(user)
}
