package go_text_templates

import (
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
)

var parseFileFetcher = fetchTemplate(func(templatePath string) (t *template.Template, err error) {
	t, err = template.New(filepath.Base(templatePath)).ParseFiles(templatePath)
	if err != nil {
		return nil, errors.Wrap(err, "parse files")
	}
	return t, nil
})
