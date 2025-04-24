// Package config provide confg file read and unmarshal
package config

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// EnvConfigFile 配置文件环境变量名称
// 指定了文件则只会使用指定的文件
const EnvConfigFile = "ENV_CONFIG_FILE"

// EnvConfigMode 配置文件环境模式
// 会成为查找配置文件文件名的一部分
const EnvConfigMode = "ENV_CONFIG_MODE"

var (
	// 配置文件列表
	files = make([]string, 0)
	// 配置文件目录
	defaultDir  = "./config"
	defaultType = "toml"
	configFile  = ""
	// env dev => dev.toml test => test.toml prod不会改变
	env = ""
)

// scanFile 查找配置文件
// 优先使用环境变量 EnvConfigFile 指定的文件，文件保存到 configFile
// 没有指定文件使用是 defaultType 后缀的所有文件，文件列表保存到 files
func scanFile() {
	env = os.Getenv(EnvConfigMode)
	// 优先环境变量指定的配置文件
	cfgFile := os.Getenv(EnvConfigFile)
	if cfgFile != "" {
		configFile = cfgFile
		return
	}
	var err error
	fileSuffix := defaultType
	if env != "" {
		fileSuffix = fmt.Sprintf("%s.%s", env, defaultType)
	}
	dirfs := os.DirFS(defaultDir)
	files, err = fs.Glob(dirfs, fmt.Sprintf("*.%s", fileSuffix))
	if err != nil {
		panic(err)
	}
}

// LoadConfig 读取解析配置文件
func LoadConfig[T any](c T, dir ...string) {
	if len(dir) > 0 {
		defaultDir = defaultDir + "/" + dir[0]
	}
	scanFile()
	v := viper.New()
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			panic(err)
		}
	}
	for _, file := range files {
		v.SetConfigFile(fmt.Sprintf("%s/%s", defaultDir, file))
		if err := v.MergeInConfig(); err != nil {
			panic(err)
		}
	}
	rootDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	v.Set("app.root_dir", rootDir)
	v.OnConfigChange(func(e fsnotify.Event) {
		slog.Info("Config file changed:" + e.Name)
		if err = v.Unmarshal(&c); err != nil {
			panic(err)
		}
	})

	if err = v.Unmarshal(&c); err != nil {
		panic(err)
	}
}
