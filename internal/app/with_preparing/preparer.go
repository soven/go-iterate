package preparing

import (
	"path/filepath"
	"strings"

	"github.com/soven/go-iterate/internal/app"
)

type basePreparer struct{}

func (basePreparer) PreparePackagePath(path string) (string, error) {
	if len(path) == 0 {
		return filepath.Dir(path), nil
	}
	return path, nil
}

func (basePreparer) PrepareContext(ctx app.GenerateIterContext) (app.GenerateIterContext, error) {
	ctx.TitlePrefix = strings.Title(strings.ToLower(ctx.TitlePrefix))
	ctx.PackageName = strings.ToLower(ctx.PackageName)
	return ctx, nil
}

var basePreparerInstance = basePreparer{}
