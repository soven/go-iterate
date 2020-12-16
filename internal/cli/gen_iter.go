package cli

import (
	"flag"
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
	path, ctx := parseGenIterFlags()
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

	if len(generatedFiles) > 0 {
		_, errWrite := fmt.Fprintf(os.Stdout, "generated:\n%s\n", strings.Join(generatedFiles, "\n"))
		if errWrite != nil {
			return errors.Wrap(errWrite, fmt.Sprintf("write result %+v to stdout", generatedFiles))
		}
	}
	return nil
}

func parseGenIterFlags() (packagePath string, ctx app.GenerateIterContext) {
	flag.StringVar(&ctx.TypeName, "target", "", "Name of target type (required)")
	flag.StringVar(&ctx.PackageName, "package", "", "Name of package (required)")
	flag.StringVar(&packagePath, "path", "", "Path to the package (required)")
	flag.StringVar(&ctx.ZeroTypeValue, "zero", "", "Zero value of target type (required)")
	flag.StringVar(&ctx.TitlePrefix, "prefix", "", "Prefix of type title (optional)")
	flag.Parse()
	return packagePath, ctx
}
