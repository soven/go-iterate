package go_replacer_assembler

import "io/ioutil"

var readFileLoader = loadSourceCode(func(sourceFilePath string) (body []byte, err error) {
	return ioutil.ReadFile(sourceFilePath)
})
