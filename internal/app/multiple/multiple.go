package multiple

import (
	"github.com/pkg/errors"
	"github.com/soven/go-iterate/internal/app"
)

type DoubleGenIter struct {
	lhs, rhs app.IterGenerator
}

func newDoubleGenIter(lhs, rhs app.IterGenerator) DoubleGenIter {
	if lhs == nil {
		lhs = app.NoGenerateIter
	}
	if rhs == nil {
		rhs = app.NoGenerateIter
	}
	return DoubleGenIter{lhs: lhs, rhs: rhs}
}

// GenerateIter should generate code using path to package and context.
func (d DoubleGenIter) GenerateIter(packagePath string, ctx app.GenerateIterContext,
) (targetFilePaths []string, err error) {
	targetFilePathsLHS, err := d.lhs.GenerateIter(packagePath, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "lhs gen iter")
	}
	targetFilePathsRHS, err := d.rhs.GenerateIter(packagePath, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "rhs gen iter")
	}

	return append(targetFilePathsLHS, targetFilePathsRHS...), nil
}

func Multiple(gens ...app.IterGenerator) app.IterGenerator {
	var res app.IterGenerator = app.NoGenerateIter
	for i := len(gens) - 1; i >= 0; i-- {
		if gens[i] == nil {
			continue
		}
		res = newDoubleGenIter(gens[i], res)
	}

	return res
}
