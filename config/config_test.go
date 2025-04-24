package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	defaultDir = "./testdata/config"
}

func TestScanFile(t *testing.T) {
	cfgFile := fmt.Sprintf("testdata/config/config.toml")
	t.Run("test env file", func(t *testing.T) {
		if err := os.Setenv(EnvConfigFile, cfgFile); err != nil {
			t.Error(err)
		}
		scanFile()
		assert.Equal(t, cfgFile, configFile)
	})

	t.Run("test env mode", func(t *testing.T) {
		e := "dev"
		_ = os.Setenv(EnvConfigFile, "")
		if err := os.Setenv(EnvConfigMode, e); err != nil {
			t.Error(err)
		}
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
		cfgFile := fmt.Sprintf("testdata/config/config.toml")
		files = []string{}
		_ = os.Setenv(EnvConfigMode, "")
		if err := os.Setenv(EnvConfigFile, cfgFile); err != nil {
			t.Error(err)
		}
		LoadConfig(&cfg)
		t.Log(configFile, files, cfg.App)
		assert.Equal(t, "GinTpl3", cfg.App.Name)
		assert.Equal(t, uint16(9000), cfg.App.Port)
	})

	t.Run("test load config with env mode", func(t *testing.T) {
		cfg := Config{}
		configFile = ""
		e := "dev"
		if err := os.Setenv(EnvConfigFile, ""); err != nil {
			t.Error(err)
		}
		if err := os.Setenv(EnvConfigMode, e); err != nil {
			t.Error(err)
		}
		LoadConfig(&cfg)
		t.Log(cfg)
		assert.Equal(t, "GinTpl", cfg.App.Name)
		assert.Equal(t, uint16(9001), cfg.App.Port)
	})

	t.Run("test load config with default dir", func(t *testing.T) {
		cfg := Config{}
		if err := os.Setenv(EnvConfigFile, ""); err != nil {
			t.Error(err)
		}
		if err := os.Setenv(EnvConfigMode, ""); err != nil {
			t.Error(err)
		}
		LoadConfig(&cfg)
		t.Log(files, configFile, cfg)
		assert.Equal(t, "GinTpl3", cfg.App.Name)
		assert.Equal(t, uint16(9000), cfg.App.Port)
		assert.Equal(t, "debug", cfg.App.Mode)
	})
}
