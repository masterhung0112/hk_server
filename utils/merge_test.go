package utils_test

import (
	"github.com/masterhung0112/go_server/utils"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestMergeStruct struct {
	suite.Suite
}

// Test merging maps alone. This isolates the complexity of merging maps from merging maps recursively in
// a struct/ptr/etc.
// Remember that for our purposes, "merging" means replacing base with patch if patch is /anything/ other than nil.

func TestMergeTestSuite(t *testing.T) {
	suite.Run(t, &TestMergeStruct{})
}

func (s *TestMergeStruct) TestMergeMapsWherePatchIsLonger() {
	m1 := map[string]int{"this": 1, "is": 2, "a map": 3}
	m2 := map[string]int{"this": 1, "is": 3, "a secnd map": 3, "another key": 4}

	expected := map[string]int{"this": 1, "is": 3, "a second map": 3, "another key": 4}
	merged, err := mergeStringIntMap(m1, m2)
	s.NoError(err)

	s.Equal(expected, merged)
}

func (s *TestMergeStruct) TestMergeMapsWhereBaseIsLonger() {
	m1 := map[string]int{"this": 1, "is": 2, "a map": 3, "with": 4, "more keys": -12}
	m2 := map[string]int{"this": 1, "is": 3, "a second map": 3}
	expected := map[string]int{"this": 1, "is": 3, "a second map": 3}

	merged, err := mergeStringIntMap(m1, m2)
	s.NoError(err)

	s.Equal(expected, merged)
}

func (s *TestMergeStruct) TestMergeMapsWhereBaseIsEmpty() {
	m1 := make(map[string]int)
	m2 := map[string]int{"this": 1, "is": 3, "a second map": 3, "another key": 4}

	expected := map[string]int{"this": 1, "is": 3, "a second map": 3, "another key": 4}
	merged, err := mergeStringIntMap(m1, m2)
	s.NoError(err)

	s.Equal(expected, merged)
}

func (s *TestMergeStruct) TestMergeMapsWherePatchIsEmpty() {
	m1 := map[string]int{"this": 1, "is": 3, "a map": 3, "another key": 4}
	var m2 map[string]int
	expected := map[string]int{"this": 1, "is": 3, "a map": 3, "another key": 4}

	merged, err := mergeStringIntMap(m1, m2)
	s.NoError(err)

	s.Equal(expected, merged)
}

func (s *TestMergeStruct) TestMergeMapStringIntPtrPatchWithDifferentKeysAndValues() {
	m1 := map[string]*int{"this": newInt(1), "is": newInt(3), "a key": newInt(3)}
	m2 := map[string]*int{"this": newInt(2), "is": newInt(3), "a key": newInt(4)}
	expected := map[string]*int{"this": newInt(2), "is": newInt(3), "a key": newInt(4)}

	merged, err := mergeStringPtrIntMap(m1, m2)
	s.NoError(err)

	s.Equal(expected, merged)
}

func (s *TestMergeStruct) TestMergeMapStringIntPtrPatchHasNilKeys() {
	//  -- doesn't matter, maps overwrite completely
	m1 := map[string]*int{"this": newInt(1), "is": newInt(3), "a key": newInt(3)}
	m2 := map[string]*int{"this": newInt(1), "is": nil, "a key": newInt(3)}
	expected := map[string]*int{"this": newInt(1), "is": nil, "a key": newInt(3)}

	merged, err := mergeStringPtrIntMap(m1, m2)
	s.NoError(err)

	s.Equal(expected, merged)
}

func (s *TestMergeStruct) TestMergeMapStringIntPtrPatchHasNilVals() {
	//  overwrite base with patch
	m1 := map[string]*int{"this": newInt(1), "is": nil, "base key": newInt(4)}
	m2 := map[string]*int{"this": newInt(1), "is": newInt(3), "a key": newInt(3)}
	expected := map[string]*int{"this": newInt(1), "is": newInt(3), "a key": newInt(3)}

	merged, err := mergeStringPtrIntMap(m1, m2)
	s.NoError(err)

	s.Equal(expected, merged)
}

func (s *TestMergeStruct) TestMergeMapStringIntPtrPatchHasNilValsPatchNil() {
	// patch is nil, so keep base
	m1 := map[string]*int{"this": newInt(1), "is": nil, "base key": newInt(4)}
	var m2 map[string]*int
	expected := map[string]*int{"this": newInt(1), "is": nil, "base key": newInt(4)}

	merged, err := mergeStringPtrIntMap(m1, m2)
	s.NoError(err)

	s.Equal(expected, merged)
}

func (s *TestMergeStruct) TestMergeMapStringIntPtrAreNotCopiedChangeInBaseDoNotAffectMerged() {
	// patch is nil, so keep base
	m1 := map[string]*int{"this": newInt(1), "is": newInt(3), "a key": newInt(4)}
	m2 := map[string]*int{"this": newInt(1), "a key": newInt(5)}
	expected := map[string]*int{"this": newInt(1), "a key": newInt(5)}

	merged, err := mergeStringPtrIntMap(m1, m2)
	s.NoError(err)

	s.Equal(expected, merged)
	*m1["this"] = 6
	s.Equal(1, *merged["this"])
}

func mergeStringIntMap(base, patch map[string]int) (map[string]int, error) {
	ret, err := utils.Merge(base, patch, nil)
	if err != nil {
		return nil, err
	}
	retTS := ret.(map[string]int)
	return retTS, nil
}

func mergeStringPtrIntMap(base, patch map[string]*int) (map[string]*int, error) {
	ret, err := utils.Merge(base, patch, nil)
	if err != nil {
		return nil, err
	}
	retTS := ret.(map[string]*int)
	return retTS, nil
}

func newBool(b bool) *bool          { return &b }
func newInt(n int) *int             { return &n }
func newInt64(n int64) *int64       { return &n }
func newString(s string) *string    { return &s }
func newInt8(n int8) *int8          { return &n }
func newInt16(n int16) *int16       { return &n }
func newInt32(n int32) *int32       { return &n }
func newFloat64(f float64) *float64 { return &f }
func newFloat32(f float32) *float32 { return &f }
func newUint(n uint) *uint          { return &n }
func newUint8(n uint8) *uint8       { return &n }
func newUint16(n uint16) *uint16    { return &n }
func newUint32(n uint32) *uint32    { return &n }
func newUint64(n uint64) *uint64    { return &n }
