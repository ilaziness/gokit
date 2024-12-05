package generator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/ilaziness/gokit/log"
)

// msyql orm:gorm, ent, sqlx

type GenType struct {
	projectName      string
	modname          string
	name             []string
	outputPath       string
	mysql            string
	redis            bool
	rocketMQProducer bool
	rocketMQConsumer bool
	otelTrace        bool
	cache            bool
	nacosConfig      bool
	nacosNaming      bool
}

var tplDir string
var Gen *GenType

func init() {
	// 获取当前文件的路径
	_, currentFile, _, _ := runtime.Caller(0)
	// 获取当前文件所在的目录
	currentDir := filepath.Dir(currentFile)
	// 拼接模板目录的路径
	tplDir = filepath.Join(currentDir, "tpl")
}

// NewGen 创建生成器, modname指定go mod名称
func NewGen(modname string, ops ...Option) *GenType {
	projectname := path.Base(modname)
	g := &GenType{
		projectName: projectname,
		modname:     modname,
		outputPath:  "./gen_output/" + projectname,
	}
	for _, opt := range ops {
		opt(g)
	}

	return g
}

type Option func(*GenType)

// WithName 应用名称，一个项目包含多个子应用传多个
func WithName(name ...string) Option {
	return func(g *GenType) {
		g.name = append(g.name, name...)
	}
}

// WithMysql 启用MySQL，指定orm：gorm, ent, sqlx
func WithMysql(mysql string) Option {
	return func(g *GenType) {
		g.mysql = mysql
	}
}

// WithRedis 启用redis
func WithRedis() Option {
	return func(g *GenType) {
		g.redis = true
	}
}

// WithRocketMQProducer 启用rocket mq生产者
func WithRocketMQProducer() Option {
	return func(g *GenType) {
		g.rocketMQProducer = true
	}
}

// WithRocketMQConsumer 启用rocket mq消费者
func WithRocketMQConsumer() Option {
	return func(g *GenType) {
		g.rocketMQConsumer = true
	}
}

// WithOtelTrace 启用链路追踪
func WithOtelTrace() Option {
	return func(g *GenType) {
		g.otelTrace = true
	}
}

// WithCache 启用redis cache
func WithCache() Option {
	return func(g *GenType) {
		g.cache = true
	}
}

// WithNacosConfig 启用nacos配置中心
func WithNacosConfig() Option {
	return func(g *GenType) {
		g.nacosConfig = true
	}
}

// WithNacosNaming 启用nacos服务注册
func WithNacosNaming() Option {
	return func(g *GenType) {
		g.nacosNaming = true
	}
}

// Generate 生成项目
func (g *GenType) Generate() {
	if err := g.touchDir(g.outputPath); err != nil {
		log.Logger.Error(err)
		return
	}
	if err := g.genMod(); err != nil {
		log.Logger.Error(err)
		return
	}
	for _, name := range g.name {
		if err := g.genConfig(name); err != nil {
			log.Logger.Error(err)
			return
		}
	}

}

func (g *GenType) genMod() error {
	data := map[string]string{
		"modname": g.modname,
	}
	return g.renderTpl("go.mod.tpl", data, g.outputPath+"/go.mod")
}

// touchDir 目录不存在时创建
func (*GenType) touchDir(dir string) error {
	var err error
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return err
}

func (g *GenType) genConfig(name string) error {
	tpl := tplDir + "/config.tpl"
	dir := g.outputPath + "/config/" + name
	output := dir + "/config.toml"
	if err := g.touchDir(dir); err != nil {
		return fmt.Errorf("touch dir failed: %v", err)
	}
	templ, err := template.ParseFiles(tpl)
	if err != nil {
		return fmt.Errorf("parse template failed: %v", err)
	}
	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}

	return templ.Execute(file, g.getData(name))
}

func (g *GenType) renderTpl(tpl string, data any, output string) error {
	templ, err := template.ParseFiles(fmt.Sprintf("%s/%s", tplDir, tpl))
	if err != nil {
		return fmt.Errorf("parse template failed: %v", err)
	}
	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}

	return templ.Execute(file, data)
}

type TplData struct {
	Name             string
	Otel             bool
	OtelTrace        bool
	Mysql            bool
	MysqlORM         string
	Redis            bool
	Cache            bool
	Nacos            bool
	NacosConfig      bool
	NacosNaming      bool
	RocketMQ         bool
	RocketMQProducer bool
	RocketMQConsumer bool
}

func (g *GenType) getData(name string) TplData {
	data := TplData{
		Name:             name,
		MysqlORM:         g.mysql,
		Redis:            g.redis,
		Cache:            g.cache,
		RocketMQ:         g.rocketMQConsumer || g.rocketMQProducer,
		RocketMQProducer: g.rocketMQProducer,
		RocketMQConsumer: g.rocketMQConsumer,
		Nacos:            g.nacosConfig || g.nacosNaming,
		NacosConfig:      g.nacosConfig,
		NacosNaming:      g.nacosNaming,
	}

	if g.mysql != "" {
		data.Mysql = true
	}
	if g.otelTrace {
		data.Otel = true
		data.OtelTrace = true
	}

	return data
}
