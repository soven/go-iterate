package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/soven/go-iterate/internal/app"
)

type IterGenerator struct {
	app app.IterGenerator
}

func NewIterGenerator(gen app.IterGenerator) IterGenerator {
	if gen == nil {
		gen = app.NoGenerateIter
	}
	return IterGenerator{app: gen}
}

func (c IterGenerator) GenIter() error {
	path, ctx := parseFlags()
	generatedFiles, err := c.app.GenerateIter(path, ctx)
	if err != nil {
		// no wrapping since no additional context.
		_, errWrite := fmt.Fprintf(os.Stderr,
			"could not generate iter for the flags for path `%s` and context `%+v`: %v\n",
			path, ctx, err)
		if errWrite != nil {
			return errors.Wrap(errWrite, fmt.Sprintf("write to stderr `%v`", err))
		}
	}

	_, errWrite := fmt.Fprintf(os.Stdout, "generated:\n%s\n", strings.Join(generatedFiles, "\n"))
	if errWrite != nil {
		return errors.Wrap(errWrite, fmt.Sprintf("write result %+v to stdout", generatedFiles))
	}
	return nil
}
