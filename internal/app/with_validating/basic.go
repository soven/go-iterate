package validating

import (
	"github.com/pkg/errors"
	"os"
)

type basic struct{}

func (basic) errIfEmptyStr(str string) error {
	if len(str) == 0 {
		return errors.New("empty")
	}
	return nil
}

func (basic) errIfFileError(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return errors.Wrap(err, "file error")
	}
	return nil
}

var basicInstance = basic{}
