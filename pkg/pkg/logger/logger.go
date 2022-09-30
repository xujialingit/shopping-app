package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultLevel = zapcore.InfoLevel

	DefaultTimeLayout = "2006-01-02 15:04:05"
)

type Option func(*option)

type option struct {
	level          zapcore.Level
	fields         map[string]string
	file           io.Writer
	timeLayout     string
	disableConsole bool
	printJson      bool
}

//设置option的日志level为zaocore.DebugLevel
func WithDebugLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.DebugLevel
	}
}

func WithInfoLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.InfoLevel
	}
}

//设置option的日志level为zaocore.WarnLevel
func WithWarnLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.WarnLevel
	}
}

//设置option的日志level为zaocore.ErrorLevel
func WithErrorLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.ErrorLevel
	}
}

//添加字段到log
func WithField(key, value string) Option {
	return func(opt *option) {
		opt.fields[key] = value
	}
}

//指定log的写入文件： os.Writer
func WithFileP(file string) Option {
	dir := filepath.Dir(file)
	if err := os.Mkdir(dir, 0766); err != nil {
		panic(err)
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0766)
	if err != nil {
		panic(err)
	}
	return func(opt *option) {
		opt.file = f
	}
}

//通过lumberjack指定文件自动切割和备份日志
func WithFileRotationP(file string) Option {
	dir := filepath.Dir(file)
	if err := os.Mkdir(dir, 0766); err != nil {
		panic(err)
	}
	return func(opt *option) {
		opt.file = &lumberjack.Logger{
			Filename:   file,
			MaxSize:    128, //单个文件最大尺寸,默认单位M
			MaxAge:     30,  //最长保留时间，单位day
			MaxBackups: 300, //最多多少个日志
			LocalTime:  true,
			Compress:   true, //是否压缩
		}
	}
}

//设置option timeLayout属性
func WithTimeLayout(timeLayout string) Option {
	return func(opt *option) {
		opt.timeLayout = timeLayout
	}
}

//是否控制台打印
func WithDisableConsole() Option {
	return func(opt *option) {
		opt.disableConsole = true
	}
}

//设置日志以json格式
func WithJsonFormat() Option {
	return func(o *option) {
		o.printJson = true
	}
}

func New(opts ...Option) (*zap.Logger, error) {
	opt := &option{
		level:  DefaultLevel,
		fields: make(map[string]string),
	}

	//修改option
	for _, f := range opts {
		f(opt)
	}

	timeLayout := DefaultTimeLayout
	if opt.timeLayout != "" {
		timeLayout = opt.timeLayout
	}

	encodeConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktract",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(time time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(time.Format(timeLayout))
		},
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	//json解码器
	jsonEncoder := zapcore.NewJSONEncoder(encodeConfig)
	if !opt.printJson {
		jsonEncoder = zapcore.NewConsoleEncoder(encodeConfig)
	}

	lowPrioity := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= opt.level && level < zapcore.ErrorLevel
	})

	highProioity := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= opt.level && level >= zap.ErrorLevel
	})

	stdout := zapcore.Lock(os.Stdout)
	stderr := zapcore.Lock(os.Stderr)

	core := zapcore.NewTee()

	if !opt.disableConsole {
		core = zapcore.NewTee(
			zapcore.NewCore(jsonEncoder, zapcore.NewMultiWriteSyncer(stdout), lowPrioity),
			zapcore.NewCore(jsonEncoder, zapcore.NewMultiWriteSyncer(stderr), highProioity),
		)
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(jsonEncoder, zapcore.AddSync(opt.file), zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return level >= opt.level
			})),
		)
	}
	logger := zap.New(core, zap.AddCaller(), zap.ErrorOutput(stderr))

	for key, value := range opt.fields {
		logger = logger.WithOptions(zap.Fields(zapcore.Field{
			Key:    key,
			Type:   zapcore.StringType,
			String: value,
		}))
	}
	return logger, nil
}

var _ Meta = (*meta)(nil)

type Meta interface {
	Key() string
	Value() interface{}
	meta()
}

type meta struct {
	key   string
	value interface{}
}

func (m meta) Key() string {
	return m.key
}

func (m meta) Value() interface{} {
	return m.value
}

func (m meta) meta() {
}

func NewMeta(key string, value interface{}) Meta {
	return &meta{key: key, value: value}
}

func WarpMeta(err error, metas ...Meta) (fields []zap.Field) {
	capacity := len(metas) + 1
	if err != nil {
		capacity++
	}
	fields = make([]zap.Field, 0, capacity)
	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	fields = append(fields, zap.Namespace("meta"))
	for _, meta := range metas {
		fields = append(fields, zap.Any(meta.Key(), meta.Value()))
	}
	return
}
