package go_replacer_assembler

import (
	"strings"

	"github.com/soven/go-iterate/internal/app"
)

var baseReplacer = makeReplacer(func(ctx app.ToMacroReplaceContext) (*strings.Replacer, error) {
	var oldnew []string
	if len(ctx.TypeNameFrom) > 0 {
		oldnew = append(oldnew, ctx.TypeNameFrom, "{{ .TypeName }}")
	}
	if len(ctx.PackageNameFrom) > 0 {
		oldnew = append(oldnew, ctx.PackageNameFrom, "{{ .PackageName }}")
	}
	if len(ctx.ZeroTypeValueFrom) > 0 {
		oldnew = append(oldnew, ctx.ZeroTypeValueFrom, "{{ .ZeroTypeValue }}")
	}
	if len(ctx.TitlePrefixFrom) > 0 {
		oldnew = append(oldnew, ctx.TitlePrefixFrom, "{{ .TitlePrefix }}")
	}
	return strings.NewReplacer(oldnew...), nil
})
