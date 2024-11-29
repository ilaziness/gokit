package mysql

import (
	"github.com/ilaziness/gokit/config"
	"gorm.io/gen"
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
