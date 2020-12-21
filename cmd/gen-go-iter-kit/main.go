package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/soven/go-iterate/internal/app"
	"github.com/soven/go-iterate/internal/app/multiple"
	formatting "github.com/soven/go-iterate/internal/app/with_formatting"
	"github.com/soven/go-iterate/internal/app/with_preparing"
	"github.com/soven/go-iterate/internal/app/with_validating"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/soven/go-iterate/internal/app/go_text_templates"
	"github.com/soven/go-iterate/internal/cli"
)

var isDevEnv bool

func init() {
	isDevEnv = strings.ToLower(os.Getenv("ENV")) == "dev"
}

var allTemplatePaths = []string{
	"templates/iter.go.tmpl",
	"templates/iter_check.go.tmpl",
	"templates/iter_convert.go.tmpl",
	"templates/iter_handle.go.tmpl",
	"templates/iter_multi.go.tmpl",
	"templates/iter_slice.go.tmpl",
	"templates/iter_util.go.tmpl",
}

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

	controller := cli.NewIterGenerator(
		formatting.WithFormatting(
			preparing.WithBaseGenIterPreparation(
				validating.WithBaseGenIterValidation(
					multiple.IterGenerators(templateGenerators...)))))
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
