package app

// GenerateIterContext is a context for iter generator.
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

// IterGenerator is a use case.
type IterGenerator interface {
	// GenerateIter should generate code using path to package and context.
	GenerateIter(packagePath string, ctx GenerateIterContext) (targetFilePaths []string, err error)
}

// GenerateIter is a shortcut implementation
// of IterGenerator based on a function
type GenerateIter func(packagePath string, ctx GenerateIterContext) (targetFilePaths []string, err error)

// GenerateIter should generate code using path to package and context.
func (g GenerateIter) GenerateIter(packagePath string, ctx GenerateIterContext) (targetFilePaths []string, err error) {
	return g(packagePath, ctx)
}

// NoGenerateIter is a zero value for IterGenerator.
// It does nothing.
var NoGenerateIter IterGenerator = GenerateIter(func(_ string, _ GenerateIterContext) ([]string, error) {
	return nil, nil
})
