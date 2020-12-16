package go_text_templates

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/pkg/errors"

	"github.com/soven/go-iterate/internal/app"
)

type templateFetcher interface {
	FetchTemplate(templatePath string) (t *template.Template, err error)
}

type fetchTemplate func(templatePath string) (t *template.Template, err error)

func (f fetchTemplate) FetchTemplate(templatePath string) (t *template.Template, err error) {
	return f(templatePath)
}

var noFetchTemplate templateFetcher = fetchTemplate(func(_ string) (*template.Template, error) {
	return nil, errors.New("no fetch template")
})

type fileNameMaker interface {
	MakeFileName(templatePath string, ctx app.GenerateIterContext, t *template.Template) (string, error)
}

type makeFileName func(templatePath string, ctx app.GenerateIterContext, t *template.Template) (string, error)

func (m makeFileName) MakeFileName(templatePath string, ctx app.GenerateIterContext, t *template.Template,
) (string, error) {
	return m(templatePath, ctx, t)
}

var makeAlwaysNameIsTarget fileNameMaker = makeFileName(func(_ string, _ app.GenerateIterContext, _ *template.Template,
) (string, error) {
	const alwaysName = "target.go"
	return alwaysName, nil
})

type AsIsGenerator struct {
	templateFilePath string
	fetcher          templateFetcher
	fileNameMaker    fileNameMaker

	fetchOnce sync.Once
	t         *template.Template
	fetchErr  error
}

func makeAsIsGenerator(templateFilePath string, fetcher templateFetcher, filePathMaker fileNameMaker) *AsIsGenerator {
	if fetcher == nil {
		fetcher = noFetchTemplate
	}
	if filePathMaker == nil {
		filePathMaker = makeAlwaysNameIsTarget
	}
	return &AsIsGenerator{
		templateFilePath: templateFilePath,
		fetcher:          fetcher,
		fileNameMaker:    filePathMaker,
	}
}

// GenerateIter generates code using path to package and context.
func (g *AsIsGenerator) GenerateIter(packagePath string, ctx app.GenerateIterContext) ([]string, error) {
	g.fetchTemplate()
	if g.fetchErr != nil {
		// assert no error.
		// if err so it means source template is corrupted.
		// just print log (ignore print error)
		_, _ = fmt.Fprintf(os.Stderr, "fetch template: %v", g.fetchErr)
		return nil, nil
	}

	targetFileName, err := g.fileNameMaker.MakeFileName(g.templateFilePath, ctx, g.t)
	if err != nil {
		return nil, errors.Wrap(err, "make file name")
	}
	targetFilePath := filepath.Join(packagePath, targetFileName)
	fh, err := os.OpenFile(targetFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w\n", targetFilePath, err)
	}
	defer func() {
		// ignore error since that is not relevant.
		_ = fh.Close()
	}()

	err = g.t.Execute(fh, ctx)
	if err != nil {
		return nil, fmt.Errorf("execute template into file %s: %w", targetFilePath, err)
	}

	return []string{targetFilePath}, nil
}

func (g *AsIsGenerator) fetchTemplate() {
	g.fetchOnce.Do(func() {
		g.t, g.fetchErr = g.fetcher.FetchTemplate(g.templateFilePath)
	})
}

// MakeAsIsGenerator returns an instance of AsIsGenerator ref.
func MakeAsIsGenerator(templateFilePath string) *AsIsGenerator {
	return makeAsIsGenerator(templateFilePath, parseFileFetcher,
		prefixedWithGen(prefixedWithTypeTitle(templateExtTrimmingInstance)))
}
