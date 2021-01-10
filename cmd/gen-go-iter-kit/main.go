package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/soven/go-iterate/internal/app"
	"github.com/soven/go-iterate/internal/app/go_text_templates"
	"github.com/soven/go-iterate/internal/app/multiple"
	"github.com/soven/go-iterate/internal/app/with_adding"
	"github.com/soven/go-iterate/internal/app/with_formatting"
	"github.com/soven/go-iterate/internal/app/with_preparing"
	"github.com/soven/go-iterate/internal/app/with_validating"
	"github.com/soven/go-iterate/internal/cli"
)

var allTemplatePaths = []string{
	"templates/iter.go.tmpl",
}

var isDevEnv = strings.ToLower(os.Getenv("ENV")) == "dev"

var (
	Version    = "v0.0.0"
	PackageDir = "."
)

func run() error {
	const versionFlag = "version"
	if len(os.Args) >= 2 && os.Args[1] == versionFlag {
		fmt.Println("Gen Go Iter KIT " + Version)
		return nil
	}

	templateGenerators := make([]app.IterGenerator, 0, len(allTemplatePaths))
	for _, templatePath := range allTemplatePaths {
		templateGenerators = append(templateGenerators,
			go_text_templates.MakeAsIsGenerator(handleTemplatePath(templatePath)))
	}

	var iterGen app.IterGenerator = preparing.WithBaseGenIterPreparation(
		validating.WithBaseGenIterValidation(
			multiple.IterGenerators(templateGenerators...)))

	if !isDevEnv {
		iterGen = adding.WithNoEditPrefixAdding(iterGen)
	}

	controller := cli.NewIterGenerator(formatting.WithFormatting(iterGen))
	err := controller.GenIter()
	if err != nil {
		return errors.Wrap(err, "controller gen iter")
	}

	return nil
}

func handleTemplatePath(source string) string {
	if isDevEnv {
		return source
	}
	return filepath.Join(PackageDir, source)
}

func main() {
	err := run()
	if err != nil {
		log.Fatalln("something went wrong:", err)
	}
}
