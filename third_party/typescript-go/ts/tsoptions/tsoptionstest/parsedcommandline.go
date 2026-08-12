package tsoptionstest

import (
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/tsoptions"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/tspath"
	"gotest.tools/v3/assert"
)

func GetParsedCommandLine(t assert.TestingT, jsonText string, files map[string]string, currentDirectory string, useCaseSensitiveFileNames bool) *tsoptions.ParsedCommandLine {
	host := NewVFSParseConfigHost(files, currentDirectory, useCaseSensitiveFileNames)
	configFileName := tspath.CombinePaths(currentDirectory, "tsconfig.json")
	tsconfigSourceFile := tsoptions.NewTsconfigSourceFileFromFilePath(configFileName, tspath.ToPath(configFileName, currentDirectory, useCaseSensitiveFileNames), jsonText)
	return tsoptions.ParseJsonSourceFileConfigFileContent(tsconfigSourceFile, host, currentDirectory, nil, nil, configFileName, nil, nil, nil)
}
