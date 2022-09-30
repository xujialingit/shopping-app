package logger

import (
	"errors"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestJsonLogger(t *testing.T) {
	logger, err := New(
		WithField("defined_key", "definde_value"),
		WithFileRotationP("./log/log.log"),
		WithDisableConsole(),
		WithDebugLevel(),
	)

	if err != nil {
		t.Fatal(err)
	}

	defer logger.Sync()
	err = errors.New("pkg error")
	logger.Error("err occuse", WarpMeta(nil, NewMeta("para1", "values1"), NewMeta("para2", "values2"))...)
	logger.Error("登录失败", WarpMeta(err, NewMeta("用户名", "978216492@qq.com"), NewMeta("密码", "123456"))...)
	logger.Info("infoLogger", zapcore.Field{Key: "测试info", String: "testInfo"})
	logger.Fatal("infoLogger", zapcore.Field{Key: "测试info", String: "testInfo"})
}
