package go_text_templates

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/soven/go-iterate/internal/app"
)

type templateExtTrimmed struct{}

func (m templateExtTrimmed) MakeFileName(templatePath string, _ app.GenerateIterContext,
	_ *template.Template) (string, error) {
	const templateExt = ".tmpl"
	templateName := filepath.Base(templatePath)
	return strings.TrimSuffix(templateName, templateExt), nil
}

var templateExtTrimmedInstance = templateExtTrimmed{}

type addTitlePrefixed struct {
	base fileNameMaker
}

func prefixedWithTypeTitle(base fileNameMaker) addTitlePrefixed {
	if base == nil {
		base = makeAlwaysNameIsTarget
	}
	return addTitlePrefixed{base: base}
}

func (m addTitlePrefixed) MakeFileName(templatePath string, ctx app.GenerateIterContext,
	t *template.Template) (string, error) {
	fileName, err := m.base.MakeFileName(templatePath, ctx, t)
	if err != nil {
		// no wrapping since no additional context
		return "", err
	}

	if len(ctx.TitlePrefix) == 0 {
		return fileName, nil
	}
	return strings.ToLower(ctx.TitlePrefix) + "_" + fileName, nil
}

type customPrefixed struct {
	base   fileNameMaker
	prefix string
}

func newPrefixed(base fileNameMaker, prefix string) customPrefixed {
	if base == nil {
		base = makeAlwaysNameIsTarget
	}
	return customPrefixed{base: base, prefix: prefix}
}

func (m customPrefixed) MakeFileName(templatePath string, ctx app.GenerateIterContext,
	t *template.Template) (string, error) {
	fileName, err := m.base.MakeFileName(templatePath, ctx, t)
	if err != nil {
		// no wrapping since no additional context
		return "", err
	}

	return m.prefix + fileName, nil
}

func prefixedWithGen(base fileNameMaker) customPrefixed {
	const prefix = "gen_"
	return newPrefixed(base, prefix)
}
