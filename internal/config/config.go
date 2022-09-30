package config

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"time"
)

/*
config使用单例模式
暴露出一个Get方法获取全局配置
*/

var config = new(Config)

type Config struct {
	Mysql struct {
		Base struct {
			ConnMaxLifeTime time.Duration `toml:"connMaxLifeTime"`
			MaxIdleConne    int           `toml:"maxIdleConn"`
			MaxOpenConn     int           `toml:"maxOpenConn"`
			Addr            string        `toml:"addr"`
			Name            string        `toml:"name"`
			Pass            string        `toml:"pass"`
			User            string        `toml:"user"`
		} `toml:"base"`
	} `toml:"mysql"`

	Jwt struct {
		Secret          string        `toml:"secret"`
		ExpireDuration  time.Duration `toml:"expireDuration"`
		RefreshDuration time.Duration `json:"refreshDuration"`
	}

	Redis struct {
		Addr         string `toml:"addr"`
		Db           int    `toml:"db"`
		Pass         string `toml:"pass"`
		MaxRetries   int    `toml:"maxRetries"`
		MinIdleConns int    `toml:"minIdleConns"`
		PoolSize     int    `toml:"poolSize"`
	} `toml:"redis"`

	Log struct {
		LogPath    string `toml:"logPath"`
		Level      string `toml:"level"`
		Stdout     bool   `toml:"stdout"`
		JsonFormat bool   `toml:"jsonFormat"`
	} `toml:"log"`

	Server struct {
		ServerName string `toml:"serverName"`
		Host       string `toml:"host"`
		Pprof      bool   `toml:"pprof"`
	} `toml:"server"`

	Email struct {
		QQ struct {
			SmtpHost string `toml:"smtpHost"`
			SmtpPort string `toml:"smtpPort"`
			Sender   string `toml:"sender"`
			Secret   string `toml:"secret"`
		} `toml:"qq"`
	} `toml:"email"`
}

//序列化
func (c *Config) ToJSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func InitConfig() {
	viper.SetConfigName("cfg")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./internal/config")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(config); err != nil {
		panic(err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if err := viper.Unmarshal(config); err != nil {
			panic(err)
		}
	})
}

func Get() Config {
	return *config
}
