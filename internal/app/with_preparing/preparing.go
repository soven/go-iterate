package preparing

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/soven/go-iterate/internal/app"
)

type preparer interface {
	PreparePackagePath(string) (string, error)
	PrepareContext(app.GenerateIterContext) (app.GenerateIterContext, error)
}

type noPreparer struct{}

func (noPreparer) PreparePackagePath(v string) (string, error) { return v, nil }
func (noPreparer) PrepareContext(ctx app.GenerateIterContext) (app.GenerateIterContext, error) {
	return ctx, nil
}

var noPreparerInstance = noPreparer{}

// GenIterValidator is a wrapper over app.IterGenerator which validates input arguments,
// then calls decorated instance.
type GenIterPreparing struct {
	decorated app.IterGenerator

	preparer preparer
}

func newGenIterPreparing(decorated app.IterGenerator, preparer preparer) GenIterPreparing {
	if decorated == nil {
		decorated = app.NoGenerateIter
	}
	if preparer == nil {
		preparer = noPreparerInstance
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

func WithBasePreparation(decorated app.IterGenerator) GenIterPreparing {
	return newGenIterPreparing(decorated, basePreparerInstance)
}
