package go_text_templates

import (
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/soven/go-iterate/internal/app"
)

type templateExtTrimming struct{}

func (m templateExtTrimming) MakeFileName(templatePath string, _ app.GenerateIterContext,
	_ *template.Template) (string, error) {
	const templateExt = ".tmpl"
	templateName := filepath.Base(templatePath)
	return strings.TrimSuffix(templateName, templateExt), nil
}

var templateExtTrimmingInstance = templateExtTrimming{}

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
	return camelToSnakeCase(ctx.TitlePrefix) + "_" + fileName, nil
}

var (
	matchFirstCap = regexp.MustCompile(`(.)([A-Z][a-z]+)`)
	matchAllCap   = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func camelToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
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
