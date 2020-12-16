package preparing

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/soven/go-iterate/internal/app"
)

type genIterPreparer interface {
	// PreparePackagePath should prepare package path.
	PreparePackagePath(string) (string, error)
	// PrepareContext should prepare context.
	PrepareContext(app.GenerateIterContext) (app.GenerateIterContext, error)
}

type noGenIterPreparer struct{}

func (noGenIterPreparer) PreparePackagePath(v string) (string, error) { return v, nil }
func (noGenIterPreparer) PrepareContext(ctx app.GenerateIterContext) (app.GenerateIterContext, error) {
	return ctx, nil
}

var noGenIterPreparerInstance = noGenIterPreparer{}

// GenIterValidator is a wrapper over app.IterGenerator which validates input arguments,
// then calls decorated instance.
type GenIterPreparing struct {
	decorated app.IterGenerator

	preparer genIterPreparer
}

func newGenIterPreparing(decorated app.IterGenerator, preparer genIterPreparer) GenIterPreparing {
	if decorated == nil {
		decorated = app.NoGenerateIter
	}
	if preparer == nil {
		preparer = noGenIterPreparerInstance
	}
	return GenIterPreparing{decorated: decorated, preparer: preparer}
}

// GenerateIter should generate code using path to package and context.
func (p GenIterPreparing) GenerateIter(packagePath string, ctx app.GenerateIterContext) ([]string, error) {
	packagePath, err := p.preparer.PreparePackagePath(packagePath)
	if err != nil {
		return nil, errors.Wrap(err, "prepare package path")
	}

	ctx, err = p.preparer.PrepareContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "prepare context")
	}

	targetFilePaths, err := p.decorated.GenerateIter(packagePath, ctx)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("generate iter for prepared pacakge path `%s` and context `%+v`",
			packagePath, ctx))
	}

	return targetFilePaths, nil
}

// WithBaseGenIterPreparation decorates app.IterGenerator with preparing context logic.
func WithBaseGenIterPreparation(decorated app.IterGenerator) GenIterPreparing {
	return newGenIterPreparing(decorated, basePreparerInstance)
}
