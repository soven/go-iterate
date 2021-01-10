package main

import (
	"log"

	"github.com/pkg/errors"

	"github.com/soven/go-iterate/internal/app"
	"github.com/soven/go-iterate/internal/app/go_replacer_assembler"
	"github.com/soven/go-iterate/internal/app/join"
	"github.com/soven/go-iterate/internal/cli"
)

var allResembledPaths = []string{
	"internal/testground/resembled/iter.go",
	"internal/testground/resembled/iter_check.go",
	"internal/testground/resembled/iter_convert.go",
	"internal/testground/resembled/iter_handle.go",
	"internal/testground/resembled/iter_multi.go",
	"internal/testground/resembled/iter_slice.go",
	"internal/testground/resembled/iter_util.go",
}

func run() error {
	templateGenerators := make([]app.TemplateAssembler, 0, len(allResembledPaths))
	for _, templatePath := range allResembledPaths {
		templateGenerators = append(templateGenerators,
			go_replacer_assembler.MakeAsIsGenerator(templatePath))
	}

	controller := cli.NewTemplateAssembler(
		join.TemplateAssemblers(templateGenerators...))
	err := controller.AssembleTemplate()
	if err != nil {
		return errors.Wrap(err, "controller assemble template")
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatalln("something went wrong:", err)
	}
}
