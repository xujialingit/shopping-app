package main

import (
	"github/xujialingit/shopping-app/internal/config"
	"github/xujialingit/shopping-app/pkg/pkg/logger"
)

func main() {
	//初始化config
	config.InitConfig()

	//日志配置
	loggerOptions := findLogConfigOption()
	logger, err := logger.New(loggerOptions...)
	if err != nil {
		panic(err)
	}
	defer func() {
		logger.Sync()
	}()

	//服务

}

func findLogConfigOption() []logger.Option {
	c := config.Get()
	result := make([]logger.Option, 0)

	if c.Log.JsonFormat {
		result = append(result, logger.WithJsonFormat())
	}

	switch c.Log.Level {
	case "DEBUG":
		result = append(result, logger.WithDebugLevel())
	case "INFO":
		result = append(result, logger.WithInfoLevel())
	case "WARN":
		result = append(result, logger.WithWarnLevel())
	case "ERROR":
		result = append(result, logger.WithErrorLevel())
	}
	result = append(result, logger.WithDisableConsole())
	result = append(result, logger.WithFileRotationP(c.Log.LogPath))
	return result
}
