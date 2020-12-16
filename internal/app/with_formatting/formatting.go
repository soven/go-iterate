package formatting

import (
	"github.com/pkg/errors"

	"github.com/soven/go-iterate/internal/app"
)

type formatter interface {
	// FormatFile should format file by the given path.
	FormatFile(filePath string) error
}

type FormatFile func(filePath string) error

func (f FormatFile) FormatFile(filePath string) error { return f(filePath) }

var noFormatter = FormatFile(func(filePath string) error { return nil })

type GenIterFormatting struct {
	decorated app.IterGenerator

	formatter formatter
}

func newGenIterFormatting(decorated app.IterGenerator, formatter formatter) GenIterFormatting {
	if decorated == nil {
		decorated = app.NoGenerateIter
	}
	if formatter == nil {
		formatter = noFormatter
	}
	return GenIterFormatting{decorated: decorated, formatter: formatter}
}

// GenerateIter should generate code using path to package and context.
func (f GenIterFormatting) GenerateIter(packagePath string, ctx app.GenerateIterContext) ([]string, error) {
	targetFilePaths, err := f.decorated.GenerateIter(packagePath, ctx)
	if err != nil {
		// no wrapping since no additional context
		return nil, err
	}

	for _, filePath := range targetFilePaths {
		err := f.formatter.FormatFile(filePath)
		if err != nil {
			return nil, errors.Wrapf(err, "format file `%s`", filePath)
		}
	}

	return targetFilePaths, nil
}

// WithFormatting decorates app.IterGenerator with formatting logic.
func WithFormatting(decorated app.IterGenerator) GenIterFormatting {
	return newGenIterFormatting(decorated, goFormatInstance)
}
