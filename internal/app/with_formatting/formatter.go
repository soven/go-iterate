// TODO the logic into go text templates
package formatting

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

type goFormat struct{}

// FormatFile formats file by the given path.
func (goFormat) FormatFile(filePath string) error {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "read file")
	}

	result, err := format.Source(fileBytes)
	if err != nil {
		// assert no error.
		// if err so it means source template is corrupted.
		// just print log (ignore print error)
		_, _ = fmt.Fprintf(os.Stderr, "go format: %v", err)
		return nil
	}

	err = ioutil.WriteFile(filePath, result, 0664)
	if err != nil {
		return errors.Wrap(err, "write file")
	}

	return nil
}

var goFormatInstance = goFormat{}
