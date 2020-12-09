package validating

import (
	"github.com/pkg/errors"
	"github.com/soven/go-iterate/internal/app"
)

type baseValidator struct {
	basic basic
}

func (v baseValidator) ValidatePackagePath(path string) error {
	err := v.basic.errIfEmptyStr(path)
	if err != nil {
		// no wrapping since no additional context
		return err
	}

	return v.basic.errIfFileError(path)
}

func (v baseValidator) ValidateContext(ctx app.GenerateIterContext) error {
	err := v.validateType(ctx.TypeName)
	if err != nil {
		return errors.Wrap(err, "type name")
	}

	err = v.validatePackageName(ctx.PackageName)
	if err != nil {
		return errors.Wrap(err, "package name")
	}

	err = v.validateTitlePrefix(ctx.TitlePrefix)
	if err != nil {
		return errors.Wrap(err, "title prefix")
	}

	err = v.validateZeroTypeValue(ctx.ZeroTypeValue)
	if err != nil {
		return errors.Wrap(err, "zero type value")
	}

	return nil
}

func (v baseValidator) validateType(typ string) error {
	// no wrapping since no additional context
	return v.basic.errIfEmptyStr(typ)
}

func (v baseValidator) validatePackageName(pkg string) error {
	// no wrapping since no additional context
	return v.basic.errIfEmptyStr(pkg)
}

func (v baseValidator) validateTitlePrefix(_ string) error {
	// no validation yet
	return nil
}

func (v baseValidator) validateZeroTypeValue(zero string) error {
	// no wrapping since no additional context
	return v.basic.errIfEmptyStr(zero)
}

var basicValidatorInstance = baseValidator{}
