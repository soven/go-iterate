package multiple

import (
	"github.com/pkg/errors"
	"github.com/soven/go-iterate/internal/app"
)

// DoubleAssembleTemplate is an implementation of app.TemplateAssembler
// for two generators.
type DoubleAssembleTemplate struct {
	lhs, rhs app.TemplateAssembler
}

func newDoubleAssembleTemplate(lhs, rhs app.TemplateAssembler) DoubleAssembleTemplate {
	if lhs == nil {
		lhs = app.NoAssembleTemplate
	}
	if rhs == nil {
		rhs = app.NoAssembleTemplate
	}
	return DoubleAssembleTemplate{lhs: lhs, rhs: rhs}
}

// AssembleTemplate assembles template from code using replace context.
func (d DoubleAssembleTemplate) AssembleTemplate(templateDir string, ctx app.ToMacroReplaceContext,
) (targetFilePaths []string, err error) {
	targetFilePathsLHS, err := d.lhs.AssembleTemplate(templateDir, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "lhs gen iter")
	}
	targetFilePathsRHS, err := d.rhs.AssembleTemplate(templateDir, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "rhs gen iter")
	}

	return append(targetFilePathsLHS, targetFilePathsRHS...), nil
}

// TemplateAssemblers joins few generators to one.
func TemplateAssemblers(gens ...app.TemplateAssembler) app.TemplateAssembler {
	var res = app.NoAssembleTemplate
	for i := len(gens) - 1; i >= 0; i-- {
		if gens[i] == nil {
			continue
		}
		res = newDoubleAssembleTemplate(gens[i], res)
	}

	return res
}
