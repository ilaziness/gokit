package mysql

import (
	nativeSQL "database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/ilaziness/gokit/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	gormDB *gorm.DB
	sqlDB  *nativeSQL.DB
)

func GormDB() *gorm.DB {
	return gormDB
}

func GetDSN(cfg *config.DB) string {
	if cfg.DSN == "" {
		return buildDSN(cfg)
	}
	return cfg.DSN
}

func buildDSN(cfg *config.DB) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DbName)
}

// InitGORM 初始化MySQL连接
// dns refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
func InitGORM(cfg *config.DB) {
	db, err := gorm.Open(mysql.Open(GetDSN(cfg)), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	gormDB = db
}

// EntDriver 创建ent client驱动
func EntDriver(cfg *config.DB) *sql.Driver {
	drv, err := sql.Open("mysql", GetDSN(cfg))
	if err != nil {
		panic(err)
	}
	if err = drv.DB().Ping(); err != nil {
		panic(err)
	}
	if cfg.MaxIdleConns > 0 {
		drv.DB().SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		drv.DB().SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.ConnMaxLifeTime > 0 {
		drv.DB().SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeTime) * time.Second)
	}
	sqlDB = drv.DB()
	return drv
}

// EntNativeDB 获取env原始db对象
func EntNativeDB() *nativeSQL.DB {
	return sqlDB
}
