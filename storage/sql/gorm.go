package sql

import (
	"context"
	"errors"
	"time"

	"github.com/ilaziness/gokit/config"
	"github.com/ilaziness/gokit/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

func GenerateDAO(cfg *config.DB, savePath string, hook func(*gen.Generator)) {
	InitGORM(cfg)
	g := gen.NewGenerator(gen.Config{
		OutPath:        savePath,
		Mode:           gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:  true,
		FieldCoverable: true,
	})
	g.UseDB(gormDB)
	hook(g)
	g.Execute()
}

type GormLogger struct {
	level logger.LogLevel
}

func NewGormLoggerDebug() *GormLogger {
	return &GormLogger{level: logger.Info}
}

func NewGormLoggerRelease() *GormLogger {
	return &GormLogger{level: logger.Error}
}

func (g *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l := *g
	l.level = level
	return &l
}

func (g *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= logger.Info {
		log.Info(ctx, msg, data...)
	}
}

func (g *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= logger.Warn {
		log.Warn(ctx, msg, data...)
	}
}

func (g *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= logger.Error {
		log.Error(ctx, msg, data...)
	}
}

func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if g.level <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil && g.level >= logger.Error && !errors.Is(err, gorm.ErrRecordNotFound) {
		g.Error(ctx, "file: %s, error: %v, sql: %v", utils.FileWithLineNum(), err, sql)
		return
	}

	if g.level == logger.Info {
		log.Info(ctx, "file:%s, cost: %s, sql: %s, rows: %d", utils.FileWithLineNum(), elapsed, sql, rows)
	}

	_, span := otel.GetTracerProvider().Tracer("gorm").Start(ctx, "gorm-sql")
	defer span.End()
	span.SetAttributes(attribute.String("sql", sql))
	span.SetAttributes(attribute.String("cost", elapsed.String()))
	span.SetAttributes(attribute.Int64("rows", rows))
	span.SetAttributes(attribute.String("file", utils.FileWithLineNum()))
}
