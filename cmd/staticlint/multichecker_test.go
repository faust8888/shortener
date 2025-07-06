// staticlint_test.go
package staticlint

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestForbiddenOSExit(t *testing.T) {
	testdata := analysistest.TestData()
	analyzer := forbiddenOSExitAnalyzer
	analysistest.Run(t, testdata, analyzer, "./...")
}
