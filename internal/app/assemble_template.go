package app

// ToMacroReplaceContext is a context for macro replacing.
type ToMacroReplaceContext struct {
	// TypeNameFrom is a type name which will be replaced
	// to the type macro. Note: if TypeNameFrom equals to TitlePrefixFrom,
	// so first TitlePrefixFrom replaced, then TypeNameFrom.
	TypeNameFrom string
	// PackageNameFrom is a type name which will be replaced
	// to the package macro.
	PackageNameFrom string
	// TitlePrefixFrom is a type name which will be replaced
	// to the prefix macro. Note: if TypeNameFrom equals to TitlePrefixFrom,
	// so first TitlePrefixFrom replaced, then TypeNameFrom.
	TitlePrefixFrom string
	// ZeroTypeValue is a type name which will be replaced
	// to the zero value type macro.
	ZeroTypeValueFrom string
}

// TemplateAssembler is a use case.
type TemplateAssembler interface {
	// AssembleTemplate should assemble template from code using replace context.
	AssembleTemplate(templateDir string, ctx ToMacroReplaceContext) (targetFilePaths []string, err error)
}

// AssembleTemplate is a shortcut implementation
// of TemplateAssembler based on a function
type AssembleTemplate func(templateDir string, ctx ToMacroReplaceContext) (targetFilePaths []string, err error)

// AssembleTemplate assembles template from code using replace context.
func (g AssembleTemplate) AssembleTemplate(templateDir string, ctx ToMacroReplaceContext,
) (targetFilePaths []string, err error) {
	return g(templateDir, ctx)
}

// NoAssembleTemplate is a zero value for TemplateAssembler.
// It does nothing.
var NoAssembleTemplate TemplateAssembler = AssembleTemplate(func(_ string, _ ToMacroReplaceContext) ([]string, error) {
	return nil, nil
})
