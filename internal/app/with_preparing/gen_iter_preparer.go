package preparing

import (
	"path/filepath"
	"strings"

	"github.com/soven/go-iterate/internal/app"
)

type baseGenIterPreparer struct{}

// PreparePackagePath prepares package path.
func (baseGenIterPreparer) PreparePackagePath(path string) (string, error) {
	if len(path) == 0 {
		return filepath.Dir(path), nil
	}
	return path, nil
}

// PrepareContext prepares context.
func (baseGenIterPreparer) PrepareContext(ctx app.GenerateIterContext) (app.GenerateIterContext, error) {
	ctx.TitlePrefix = strings.Title(ctx.TitlePrefix)
	return ctx, nil
}

var basePreparerInstance = baseGenIterPreparer{}
