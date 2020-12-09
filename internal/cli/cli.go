package cli

import (
	"flag"

	"github.com/soven/go-iterate/internal/app"
)

func parseFlags() (packagePath string, ctx app.GenerateIterContext) {
	flag.StringVar(&ctx.TypeName, "target", "", "Name of target type (required)")
	flag.StringVar(&ctx.PackageName, "package", "", "Name of package (required)")
	flag.StringVar(&packagePath, "path", "", "Path to the package (required)")
	flag.StringVar(&ctx.ZeroTypeValue, "zero", "", "Zero value of target type (required)")
	flag.StringVar(&ctx.TitlePrefix, "prefix", "", "Prefix of type title (optional)")
	flag.Parse()
	return packagePath, ctx
}
