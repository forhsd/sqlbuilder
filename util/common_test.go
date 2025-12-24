package util

import (
	"fmt"
	"testing"
)

func TestPath(t *testing.T) {
	path := SourceCodeSubstringPath(`sqlbuilder.*?/`)
	fmt.Println(path)
}
