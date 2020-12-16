package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/soven/go-iterate/internal/app"
)

type TemplateAssembler struct {
	app app.TemplateAssembler
}

func NewTemplateAssembler(gen app.TemplateAssembler) TemplateAssembler {
	if gen == nil {
		gen = app.NoAssembleTemplate
	}
	return TemplateAssembler{app: gen}
}

func (c TemplateAssembler) AssembleTemplate() error {
	path, ctx := parseAssembleTemplateFlags()
	generatedFiles, err := c.app.AssembleTemplate(path, ctx)
	if err != nil {
		// no wrapping since no additional context.
		_, errWrite := fmt.Fprintf(os.Stderr,
			"could not assemble template for the flags for path `%s` and context `%+v`: %v\n",
			path, ctx, err)
		if errWrite != nil {
			return errors.Wrap(errWrite, fmt.Sprintf("write to stderr `%v`", err))
		}
	}

	if len(generatedFiles) > 0 {
		_, errWrite := fmt.Fprintf(os.Stdout, "assembled:\n%s\n", strings.Join(generatedFiles, "\n"))
		if errWrite != nil {
			return errors.Wrap(errWrite, fmt.Sprintf("write result %+v to stdout", generatedFiles))
		}
	}
	return nil
}

func parseAssembleTemplateFlags() (templateDir string, ctx app.ToMacroReplaceContext) {
	flag.StringVar(&ctx.TypeNameFrom, "target", "", "Name of target type (required)")
	flag.StringVar(&ctx.PackageNameFrom, "package", "", "Name of package (required)")
	flag.StringVar(&templateDir, "path", "", "Path to the template dir (required)")
	flag.StringVar(&ctx.ZeroTypeValueFrom, "zero", "", "Zero value of target type (required)")
	flag.StringVar(&ctx.TitlePrefixFrom, "prefix", "", "Prefix of type title (optional)")
	flag.Parse()
	return templateDir, ctx
}
