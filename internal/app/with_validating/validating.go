package validating

import (
	"github.com/pkg/errors"
	"github.com/soven/go-iterate/internal/app"
)

type validator interface {
	ValidatePackagePath(string) error
	ValidateContext(app.GenerateIterContext) error
}

type alwaysErrValidator struct{}

func (alwaysErrValidator) ValidatePackagePath(_ string) error { return errors.New("error") }
func (alwaysErrValidator) ValidateContext(_ app.GenerateIterContext) error {
	return errors.New("error")
}

var alwaysErrValidatorInstance = alwaysErrValidator{}

// GenIterValidator is a wrapper over app.IterGenerator which validates input arguments,
// then calls decorated instance.
type GenIterValidator struct {
	decorated app.IterGenerator

	validator validator
}

func newGenIterValidator(decorated app.IterGenerator, validator validator) GenIterValidator {
	if decorated == nil {
		decorated = app.NoGenerateIter
	}
	if validator == nil {
		validator = alwaysErrValidatorInstance
	}
	return GenIterValidator{decorated: decorated, validator: validator}
}

// GenerateIter should generate code using path to package and context.
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

func WithBaseValidation(decorated app.IterGenerator) GenIterValidator {
	return newGenIterValidator(decorated, baseValidator{basic: basicInstance})
}
