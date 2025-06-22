package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	defaultDir = "./testdata/config"
}

func TestScanFile(t *testing.T) {
	cfgFile := "testdata/config/config.toml"
	t.Run("test env file", func(t *testing.T) {
		t.Setenv(EnvConfigFile, cfgFile)
		scanFile()
		assert.Equal(t, cfgFile, configFile)
	})

	t.Run("test env mode", func(t *testing.T) {
		e := "dev"
		t.Setenv(EnvConfigFile, "")
		t.Setenv(EnvConfigMode, e)
		scanFile()
		assert.Equal(
			t,
			[]string{
				fmt.Sprintf("config.%s.toml", e),
				fmt.Sprintf("db.%s.toml", e),
			},
			files,
		)
	})
}

func TestLoadConfig(t *testing.T) {
	type Config struct {
		App App `toml:"app"`
	}
	t.Run("test load config with env file", func(t *testing.T) {
		cfg := Config{}
		cfgFile := "testdata/config/config.toml"
		files = []string{}
		t.Setenv(EnvConfigMode, "")
		t.Setenv(EnvConfigFile, cfgFile)
		LoadConfig(&cfg)
		t.Log(configFile, files, cfg.App)
		assert.Equal(t, "GinTpl3", cfg.App.Name)
		assert.Equal(t, uint16(9000), cfg.App.Port)
	})

	t.Run("test load config with env mode", func(t *testing.T) {
		cfg := Config{}
		configFile = ""
		e := "dev"
		t.Setenv(EnvConfigFile, "")
		t.Setenv(EnvConfigMode, e)
		LoadConfig(&cfg)
		t.Log(cfg)
		assert.Equal(t, "GinTpl", cfg.App.Name)
		assert.Equal(t, uint16(9001), cfg.App.Port)
	})

	t.Run("test load config with default dir", func(t *testing.T) {
		cfg := Config{}
		t.Setenv(EnvConfigFile, "")
		t.Setenv(EnvConfigMode, "")
		LoadConfig(&cfg)
		t.Log(files, configFile, cfg)
		assert.Equal(t, "GinTpl3", cfg.App.Name)
		assert.Equal(t, uint16(9000), cfg.App.Port)
		assert.Equal(t, "debug", cfg.App.Mode)
	})
}
