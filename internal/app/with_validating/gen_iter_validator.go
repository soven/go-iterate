package validating

import (
	"github.com/pkg/errors"
	"github.com/soven/go-iterate/internal/app"
)

type baseGenIterValidator struct {
	basic basic
}

// ValidatePackagePath does package path validation.
func (v baseGenIterValidator) ValidatePackagePath(path string) error {
	err := v.basic.errIfEmptyStr(path)
	if err != nil {
		// no wrapping since no additional context
		return err
	}

	return v.basic.errIfFileError(path)
}

// ValidateContext does context validation.
func (v baseGenIterValidator) ValidateContext(ctx app.GenerateIterContext) error {
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

func (v baseGenIterValidator) validateType(typ string) error {
	// no wrapping since no additional context
	return v.basic.errIfEmptyStr(typ)
}

func (v baseGenIterValidator) validatePackageName(pkg string) error {
	// no wrapping since no additional context
	return v.basic.errIfEmptyStr(pkg)
}

func (v baseGenIterValidator) validateTitlePrefix(_ string) error {
	// no validation yet
	return nil
}

func (v baseGenIterValidator) validateZeroTypeValue(zero string) error {
	// no wrapping since no additional context
	return v.basic.errIfEmptyStr(zero)
}

var basicGenIterValidatorInstance = baseGenIterValidator{basic: basicInstance}
