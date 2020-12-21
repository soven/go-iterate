package go_replacer_assembler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/soven/go-iterate/internal/app"
)

type sourceCodeLoader interface {
	LoadSourceCode(sourceFilePath string) (body []byte, err error)
}

type loadSourceCode func(sourceFilePath string) (body []byte, err error)

func (l loadSourceCode) LoadSourceCode(sourceFilePath string) (body []byte, err error) {
	return l(sourceFilePath)
}

var noLoadSourceCode sourceCodeLoader = loadSourceCode(func(_ string) ([]byte, error) {
	return nil, errors.New("no load source code")
})

type fileNameMaker interface {
	MakeFileName(sourceCodePath string, ctx app.ToMacroReplaceContext) (string, error)
}

type makeFileName func(sourceCodePath string, ctx app.ToMacroReplaceContext) (string, error)

func (m makeFileName) MakeFileName(sourceCodePath string, ctx app.ToMacroReplaceContext,
) (string, error) {
	return m(sourceCodePath, ctx)
}

var makeAlwaysNameIsTarget fileNameMaker = makeFileName(func(_ string, _ app.ToMacroReplaceContext,
) (string, error) {
	const alwaysName = "target.go"
	return alwaysName, nil
})

type replacerMaker interface {
	// MakeReplacer should always return not nil replacer if error is nil.
	MakeReplacer(ctx app.ToMacroReplaceContext) (*strings.Replacer, error)
}

type makeReplacer func(ctx app.ToMacroReplaceContext) (*strings.Replacer, error)

func (m makeReplacer) MakeReplacer(ctx app.ToMacroReplaceContext) (*strings.Replacer, error) {
	return m(ctx)
}

var noReplace replacerMaker = makeReplacer(func(_ app.ToMacroReplaceContext) (*strings.Replacer, error) {
	return strings.NewReplacer(), nil
})

type AsIsGenerator struct {
	sourceCodePath string
	loader         sourceCodeLoader
	fileNameMaker  fileNameMaker
	replacerMaker  replacerMaker

	loadOnce   sync.Once
	sourceCode []byte
	loadErr    error
}

func makeAsIsGenerator(sourceCodePath string, loader sourceCodeLoader, fileNameMaker fileNameMaker,
	replacerMaker replacerMaker,
) *AsIsGenerator {
	if loader == nil {
		loader = noLoadSourceCode
	}
	if fileNameMaker == nil {
		fileNameMaker = makeAlwaysNameIsTarget
	}
	if replacerMaker == nil {
		replacerMaker = noReplace
	}
	return &AsIsGenerator{
		sourceCodePath: sourceCodePath,
		loader:         loader,
		fileNameMaker:  fileNameMaker,
		replacerMaker:  replacerMaker,
	}
}

// AssembleTemplate assembles template from code using replace context.
func (g *AsIsGenerator) AssembleTemplate(templateDir string, ctx app.ToMacroReplaceContext,
) (targetFilePaths []string, err error) {
	g.loadSourceCode()
	if g.loadErr != nil {
		// assert no error.
		// just print log (ignore print error)
		_, _ = fmt.Fprintf(os.Stderr, "load source code: %v\n", g.loadErr)
		return nil, nil
	}

	targetFileName, err := g.fileNameMaker.MakeFileName(g.sourceCodePath, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "make file name")
	}
	targetFilePath := filepath.Join(templateDir, targetFileName)

	replacer, err := g.replacerMaker.MakeReplacer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "make replacer")
	}

	fh, err := os.OpenFile(targetFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", targetFilePath, err)
	}
	defer func() {
		// ignore error since that is not relevant.
		_ = fh.Close()
	}()

	_, err = replacer.WriteString(fh, string(g.sourceCode))
	if err != nil {
		return nil, fmt.Errorf("replace with writing to %s: %w",
			targetFilePath, err)
	}

	return []string{targetFilePath}, nil
}

func (g *AsIsGenerator) loadSourceCode() {
	g.loadOnce.Do(func() {
		g.sourceCode, g.loadErr = g.loader.LoadSourceCode(g.sourceCodePath)
	})
}

// MakeAsIsGenerator returns an instance of AsIsGenerator ref.
func MakeAsIsGenerator(sourceCodePath string) *AsIsGenerator {
	return makeAsIsGenerator(sourceCodePath, readFileLoader,
		templateExtAddingInstance, baseReplacer)
}
