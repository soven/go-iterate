package validating

import (
	"github.com/pkg/errors"
	"github.com/soven/go-iterate/internal/app"
)

type genIterValidator interface {
	// ValidatePackagePath should do package path validation.
	ValidatePackagePath(string) error
	// ValidateContext does context validation.
	ValidateContext(app.GenerateIterContext) error
}

type alwaysErrGenIterValidator struct{}

func (alwaysErrGenIterValidator) ValidatePackagePath(_ string) error { return errors.New("error") }
func (alwaysErrGenIterValidator) ValidateContext(_ app.GenerateIterContext) error {
	return errors.New("error")
}

var alwaysErrGenIterValidatorInstance = alwaysErrGenIterValidator{}

// GenIterValidator is a wrapper over app.IterGenerator which validates input arguments,
// then calls decorated instance.
type GenIterValidator struct {
	decorated app.IterGenerator

	validator genIterValidator
}

func newGenIterValidator(decorated app.IterGenerator, validator genIterValidator) GenIterValidator {
	if decorated == nil {
		decorated = app.NoGenerateIter
	}
	if validator == nil {
		validator = alwaysErrGenIterValidatorInstance
	}
	return GenIterValidator{decorated: decorated, validator: validator}
}

// GenerateIter generates code using path to package and context.
func (v GenIterValidator) GenerateIter(packagePath string, ctx app.GenerateIterContext) ([]string, error) {
	err := v.validator.ValidatePackagePath(packagePath)
	if err != nil {
		return nil, errors.Wrap(err, "invalid package path")
	}

	err = v.validator.ValidateContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "invalid context")
	}

	// no wrapping since no additional context
	return v.decorated.GenerateIter(packagePath, ctx)
}

// WithBaseGenIterValidation decorates app.IterGenerator with validation logic.
func WithBaseGenIterValidation(decorated app.IterGenerator) GenIterValidator {
	return newGenIterValidator(decorated, basicGenIterValidatorInstance)
}
