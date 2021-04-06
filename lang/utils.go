package lang

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
)

// CombinedCode get combined code that using import
func CombinedCode(basePath, filename string) (string, error) {
	filePath := path.Join(basePath, filename)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	reg := regexp.MustCompile("import '(.*?)'")
	newData := reg.ReplaceAllStringFunc(string(data), func(part string) string {
		find := reg.FindStringSubmatch(part)
		if len(find) == 0 {
			panic(fmt.Sprintf("import error %s", part))
		}
		tmpCode, err := CombinedCode(path.Dir(filePath), find[1])
		if err != nil {
			return fmt.Sprintf("echoln('import %s error %s from %s')", find[1], err.Error(), filename)
		}
		return tmpCode
	})
	return newData, nil
}
