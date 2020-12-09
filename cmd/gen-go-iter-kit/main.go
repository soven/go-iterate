package main

import (
	formatting "github.com/soven/go-iterate/internal/app/with_formatting"
	"github.com/soven/go-iterate/internal/app/with_preparing"
	"github.com/soven/go-iterate/internal/app/with_validating"
	"log"

	"github.com/pkg/errors"

	"github.com/soven/go-iterate/internal/app/go_text_templates"
	"github.com/soven/go-iterate/internal/cli"
)

func run() error {
	controller := cli.NewIterGenerator(
		formatting.WithFormatting(
			preparing.WithBasePreparation(
				validating.WithBaseValidation(
					go_text_templates.MakeAsIsGenerator("templates/iter_slice.go.tmpl")))))
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
