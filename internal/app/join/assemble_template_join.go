package join

import (
	"github.com/pkg/errors"
	"github.com/soven/go-iterate/internal/app"
	"io"
	"io/ioutil"
	"os"
	"regexp"
)

// GlueAssembleTemplate is an implementation of app.TemplateAssembler
// which joins the slave's results to the first master's result file and remove slave's file.
// It also deletes package name heading line from slave file.
// If first master results is empty list, so it does nothing.
type GlueAssembleTemplate struct {
	master, slave app.TemplateAssembler
}

func newGlueAssembleTemplate(master, slave app.TemplateAssembler) GlueAssembleTemplate {
	if master == nil {
		master = app.NoAssembleTemplate
	}
	if slave == nil {
		slave = app.NoAssembleTemplate
	}
	return GlueAssembleTemplate{master: master, slave: slave}
}

// AssembleTemplate assembles template from code using replace context.
func (d GlueAssembleTemplate) AssembleTemplate(templateDir string, ctx app.ToMacroReplaceContext,
) (targetFilePaths []string, err error) {
	targetFilePathsMaster, err := d.master.AssembleTemplate(templateDir, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "master gen iter")
	}
	if len(targetFilePathsMaster) == 0 {
		return nil, nil
	}
	targetFilePathMaster := targetFilePathsMaster[0]

	targetFileMaster, err := os.OpenFile(targetFilePathMaster, os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		return nil, errors.Wrap(err, "open master file")
	}

	for _, filePath := range targetFilePathsMaster[1:] {
		err := d.glueFile(targetFileMaster, filePath)
		if err != nil {
			return nil, errors.Wrapf(err, "glue one of masters file %s", filePath)
		}
	}

	targetFilePathsSlave, err := d.slave.AssembleTemplate(templateDir, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "slave gen iter")
	}

	for _, filePath := range targetFilePathsSlave {
		err := d.glueFile(targetFileMaster, filePath)
		if err != nil {
			return nil, errors.Wrapf(err, "glue one of slaves file %s", filePath)
		}
	}

	return []string{targetFilePathMaster}, nil
}

func (d GlueAssembleTemplate) glueFile(target io.Writer, filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "read file content")
	}
	err = d.glue(target, content)
	if err != nil {
		return errors.Wrap(err, "glue")
	}

	// remove glued file (any way ignore error)
	_ = os.Remove(filePath)

	return nil
}

func (d GlueAssembleTemplate) glue(target io.Writer, content []byte) error {
	if len(content) == 0 {
		return nil
	}

	// write line feed.
	_, err := target.Write([]byte{'\n'})
	if err != nil {
		return errors.Wrap(err, "write line feed to target")
	}
	// write content
	_, err = target.Write(cleanHeadLine(content))
	if err != nil {
		return errors.Wrap(err, "write content to target")
	}
	return nil
}

const headLineRegexpPattern = `^package\s+[\w\s\.{\\}]+\n+(?:import\s\"[\w+\-\.\/]+\")?`

var headLineRegexp = regexp.MustCompile(headLineRegexpPattern)

func cleanHeadLine(content []byte) []byte {
	// replace package line to empty bytes
	return headLineRegexp.ReplaceAll(content, []byte{})
}

// TemplateAssemblers joins few generators to one.
func TemplateAssemblers(gens ...app.TemplateAssembler) app.TemplateAssembler {
	var res = app.NoAssembleTemplate
	for i := len(gens) - 1; i >= 0; i-- {
		if gens[i] == nil {
			continue
		}
		res = newGlueAssembleTemplate(gens[i], res)
	}

	return res
}
