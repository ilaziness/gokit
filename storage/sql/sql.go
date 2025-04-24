package sql

import (
	nativeSQL "database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/hook"
	// _ "github.com/jackc/pgx/v5/stdlib"  // ent pgx驱动
	"github.com/jmoiron/sqlx"
	// _ "github.com/lib/pq"  //sqlx的驱动
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	gormDB     *gorm.DB
	entSQLDB   *nativeSQL.DB
	sqlxDB     *sqlx.DB
	gormDriver = map[string]func(string) gorm.Dialector{
		"mysql":    mysql.Open,
		"postgres": postgres.Open,
		"sqlite3":  sqlite.Open,
	}
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
	driver := gormDriver[cfg.Dialect]
	if driver == nil {
		panic("error db dialect")
	}
	db, err := gorm.Open(driver(GetDSN(cfg)), &gorm.Config{
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
func EntDriver(cfg *config.DB) *entsql.Driver {
	if cfg.Dialect == "pgx" {
		db, err := nativeSQL.Open("pgx", GetDSN(cfg))
		if err != nil {
			panic(err)
		}
		if err = db.Ping(); err != nil {
			panic(err)
		}
		return entsql.OpenDB(dialect.Postgres, db)
	}
	drv, err := entsql.Open(cfg.Dialect, GetDSN(cfg))
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
	entSQLDB = drv.DB()
	return drv
}

// EntNativeDB 获取ent原始db对象
func EntNativeDB() *nativeSQL.DB {
	return entSQLDB
}

func InitSqlx(cfg *config.DB) {
	// Connect连接并ping
	db, err := sqlx.Connect(cfg.Dialect, GetDSN(cfg))
	if err != nil {
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
