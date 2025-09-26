package config

import (
	"embed"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"gitee.com/quant1x/gox/api"
	"gitee.com/quant1x/gox/logger"
	"gitee.com/quant1x/gox/util/homedir"
	"gitee.com/quant1x/pkg/yaml"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"xquant/conf"
	"xquant/pkg/utils"
)

var (
	configOnce   sync.Once
	xquantConfig atomic.Value
)

var ConfEnv = utils.GetEnvWithDefault("MAAS_CONF_ENV", "local")

func GetConfFileName() string {
	return fmt.Sprintf("conf.%s.yaml", env.ConfEnv)
}

func Load() *XquantConfig {
	configOnce.Do(func() {
		c := &XquantConfig{}
		if err := conf.Load(GetConfFileName(), c); err != nil {
			panic(fmt.Errorf("error to load conf file: %s", err))
		}

		afterLoad(c)
		xquantConfig.Store(c)
	})

	return xquantConfig.Load().(*XquantConfig)
}

const (
	TraceLevel string = "trace"
	DebugLevel string = "debug"
	InfoLevel  string = "info"
	WarnLevel  string = "warn"
	ErrorLevel string = "error"
	FatalLevel string = "fatal"
)

func getIntLevel(l string) hlog.Level {
	switch l {
	case TraceLevel:
		return hlog.LevelTrace
	case DebugLevel:
		return hlog.LevelDebug
	case InfoLevel:
		return hlog.LevelInfo
	case WarnLevel:
		return hlog.LevelWarn
	case ErrorLevel:
		return hlog.LevelError
	case FatalLevel:
		return hlog.LevelFatal
	default:
		return hlog.LevelInfo
	}
}

var CurrentLogLevel = hlog.LevelInfo

func afterLoad(c *XquantConfig) {
	if logLevel := utils.GetEnv("XQUANT_LOG_LEVEL"); logLevel != "" {
		CurrentLogLevel = getIntLevel(logLevel)
		hlog.SetLevel(CurrentLogLevel)
	} else {
		CurrentLogLevel = getIntLevel(c.LogLevel)
		hlog.SetLevel(CurrentLogLevel)
	}

	hlog.Debugf("[SUCCESS_LOAD_CONFIG] config_file=%s: config=%s", GetConfFileName(), utils.MustMarshalIndent(c))
}

const (
	// ResourcesPath 资源路径
	ResourcesPath = "resources"
)

//go:embed resources/*
var resources embed.FS

// 配置常量
const (
	// 配置文件名
	configFilename = "quant1x.yaml"
)

var (
	quant1XConfigFilename = "" // 配置文件完整路径
	listConfigFile        = []string{
		"~/." + configFilename,
		"~/.quant1x/" + configFilename,
		"~/runtime/etc/" + configFilename,
	}
)

// Quant1XConfig Quant1X基础配置
type Quant1XConfig struct {
	BaseDir string           `yaml:"basedir"` // 基础路径
	Data    DataParameter    `yaml:"data"`    // 数据源
	Runtime RuntimeParameter `yaml:"runtime"` // 运行时参数
	Trader  TraderParameter  `yaml:"trader"`  // 预览交易参数
}

// GetConfigFilename 获取配置文件路径
func GetConfigFilename() string {
	return quant1XConfigFilename
}

var (
	// GlobalConfig engine配置信息
	GlobalConfig Quant1XConfig
)

// LoadConfig 加载配置文件
func LoadConfig() (config Quant1XConfig, found bool) {
	for _, v := range listConfigFile {
		filename, err := homedir.Expand(v)
		if err != nil {
			continue
		}
		if !api.FileExist(filename) {
			continue
		}
		err = parseYamlConfig(filename, &config)
		if err != nil {
			logger.Fatalf("%+v", err)
		}
		found = true
		break
	}
	return
}

// ReadConfig 读取配置文件
func ReadConfig(rootPath string) (config Quant1XConfig) {
	target := GetConfigFilename()
	// 如果文件不存在, 导出默认配置
	if !api.FileExist(target) {
		target = rootPath + "/" + configFilename
		target, _ = homedir.Expand(target)
		filename := fmt.Sprintf("%s/%s", ResourcesPath, configFilename)
		_ = api.Export(resources, filename, target)
	}
	err := parseYamlConfig(target, &config)
	if err != nil {
		logger.Fatalf("%+v", err)
	}
	return
}

func parseYamlConfig(filename string, config *Quant1XConfig) error {
	if api.FileExist(filename) {
		dataBytes, err := os.ReadFile(filename)
		if err != nil {
			logger.Error(err)
			return err
		}
		err = yaml.Unmarshal(dataBytes, config)
		if err != nil {
			logger.Error(err)
			return err
		}
		config.BaseDir = strings.TrimSpace(config.BaseDir)
		if len(config.BaseDir) > 0 {
			quant1XConfigFilename = filename
		}
	}
	return nil
}
