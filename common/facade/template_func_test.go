package facade

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeArrValue(t *testing.T) {
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

	finalStr5 := DefaultValue(16, 10)
	assert.EqualValues(t, 16, finalStr5)

	finalStr6 := DefaultValue(0, 10)
	assert.EqualValues(t, 10, finalStr6)

	finalStr7 := DefaultValue("", "aasd")
	assert.EqualValues(t, "aasd", finalStr7)

	finalStr8 := DefaultValue([]any{1, 2, 3}, "aasd")
	assert.EqualValues(t, []any{1, 2, 3}, finalStr8)

	finalStr9 := DefaultValue(nil, []any{"a", "b", "c"})
	assert.EqualValues(t, []any{"a", "b", "c"}, finalStr9)
}
