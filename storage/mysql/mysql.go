package mysql

import (
	nativeSQL "database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/hook"
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	gormDB *gorm.DB
	sqlDB  *nativeSQL.DB
	sqlxDB *sqlx.DB
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
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=True", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DbName)
}

// InitGORM 初始化MySQL连接
// dns refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
func InitGORM(cfg *config.DB) {
	l := NewGormLoggerRelease()
	if cfg.Debug {
		l = NewGormLoggerDebug()
	}
	db, err := gorm.Open(mysql.Open(GetDSN(cfg)), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 l,
	})
	if err != nil {
		panic(err)
	}
	gormDB = db
	gDB, err := gormDB.DB()
	if err != nil {
		panic(err)
	}
	if cfg.MaxIdleConns > 0 {
		gDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		gDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.ConnMaxLifeTime > 0 {
		gDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeTime) * time.Second)
	}
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

func InitSqlx(cfg *config.DB) {
	db, err := sqlx.Connect("mysql", GetDSN(cfg))
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	sqlxDB = db
	if cfg.MaxIdleConns > 0 {
		sqlxDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		sqlxDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.ConnMaxLifeTime > 0 {
		sqlxDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeTime) * time.Second)
	}

	hook.Exit.Register(func() {
		_ = sqlxDB.Close()
	})
}

func SqlxDB() *sqlx.DB {
	return sqlxDB
}
