package go_replacer_assembler

import (
	"path/filepath"

	"github.com/soven/go-iterate/internal/app"
)

type templateExtAdding struct{}

func (m templateExtAdding) MakeFileName(sourceCodePath string, _ app.ToMacroReplaceContext) (string, error) {
	const templateExt = ".tmpl"
	templateName := filepath.Base(sourceCodePath)
	return templateName + templateExt, nil
}

var templateExtAddingInstance = templateExtAdding{}
