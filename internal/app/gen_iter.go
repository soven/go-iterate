package app

type GenerateIterContext struct {
	// TypeName sets the name of the target type to generate iterator for.
	TypeName string
	// PackageName sets the package name of the package where
	// the code will be generated.
	PackageName string
	// TitlePrefix sets the prefix of titles in generated iterator types.
	TitlePrefix string
	// ZeroTypeValue sets zero value of the target type.
	ZeroTypeValue string
}

type IterGenerator interface {
	// GenerateIter should generate code using path to package and context.
	GenerateIter(packagePath string, ctx GenerateIterContext) (targetFilePaths []string, err error)
}

type GenerateIter func(packagePath string, ctx GenerateIterContext) (targetFilePaths []string, err error)

func (g GenerateIter) GenerateIter(packagePath string, ctx GenerateIterContext) (targetFilePaths []string, err error) {
	return g(packagePath, ctx)
}

var NoGenerateIter = GenerateIter(func(_ string, _ GenerateIterContext) ([]string, error) { return nil, nil })
