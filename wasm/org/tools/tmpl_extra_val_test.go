package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeArrValueNoReflect(t *testing.T) {
	orgArr1 := []any{"a", "b", "c"}
	finalStr1 := Normalize1Arr(orgArr1)
	assert.EqualValues(t, "'a','b','c'", finalStr1)

	orgArr2 := []any{1, 2, 3}
	finalStr2 := Normalize1Arr(orgArr2)
	assert.EqualValues(t, "1,2,3", finalStr2)

	orgArr3 := []any{true, false, true}
	finalStr3 := Normalize1Arr(orgArr3)
	assert.EqualValues(t, "true,false,true", finalStr3)

	finalStr4 := Normalize1Arr(nil)
	assert.EqualValues(t, "", finalStr4)

	orgArr5 := []string{"a", "b", "c"}
	finalStr5 := Normalize1Arr(orgArr5)
	assert.EqualValues(t, "'a','b','c'", finalStr5)

	orgArr6 := []int{1, 2, 3}
	finalStr6 := Normalize1Arr(orgArr6)
	assert.EqualValues(t, "1,2,3", finalStr6)

	orgArr7 := []bool{true, false, true}
	finalStr7 := Normalize1Arr(orgArr7)
	assert.EqualValues(t, "true,false,true", finalStr7)
}
