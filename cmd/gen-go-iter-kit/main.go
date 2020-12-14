package main

import (
	"github.com/soven/go-iterate/internal/app"
	"github.com/soven/go-iterate/internal/app/multiple"
	formatting "github.com/soven/go-iterate/internal/app/with_formatting"
	"github.com/soven/go-iterate/internal/app/with_preparing"
	"github.com/soven/go-iterate/internal/app/with_validating"
	"log"

	"github.com/pkg/errors"

	"github.com/soven/go-iterate/internal/app/go_text_templates"
	"github.com/soven/go-iterate/internal/cli"
)

var allTemplatePaths = []string{
	"templates/iter.go.tmpl",
	"templates/iter_check.go.tmpl",
	"templates/iter_convert.go.tmpl",
	"templates/iter_handle.go.tmpl",
	"templates/iter_multi.go.tmpl",
	"templates/iter_slice.go.tmpl",
	"templates/iter_util.go.tmpl",
}

func run() error {
	templateGenerators := make([]app.IterGenerator, 0, len(allTemplatePaths))
	for _, templatePath := range allTemplatePaths {
		templateGenerators = append(templateGenerators,
			go_text_templates.MakeAsIsGenerator(templatePath))
	}

	controller := cli.NewIterGenerator(
		formatting.WithFormatting(
			preparing.WithBasePreparation(
				validating.WithBaseValidation(
					multiple.Multiple(templateGenerators...)))))
	err := controller.GenIter()
	if err != nil {
		return errors.Wrap(err, "gen iter")
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatalln("something went wrong:", err)
	}
}
