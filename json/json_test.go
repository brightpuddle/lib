package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	a := assert.New(t)
	s := Set("", "a.b", "c")
	a.Equal(s, `{"a":{"b":"c"}}`)
}

func TestSetRaw(t *testing.T) {
	a := assert.New(t)
	s := SetRaw("", "a", `{"b":"c"}`)
	a.Equal(s, `{"a":{"b":"c"}}`)
}

type childStruct struct {
	OK   bool
	List []bool
	Dict map[string]bool
}

type testStructInterface interface {
	Check()
}

type testStruct struct {
	List          []string
	PtrList       *[]string
	PtrDict       *map[string]string
	Dict          map[string]string
	Child         childStruct
	ListChild     []childStruct
	DictChild     map[string]childStruct
	DictListChild map[string][]childStruct
	ListDictChild []map[string]childStruct
	Str           string
}

func (testStruct) Check() {}

func TestAllocNull(t *testing.T) {
	a := assert.New(t)
	s := testStruct{}
	allocNull(&s)
	a.NotNil(s.List)
	a.NotNil(s.Dict)
	a.NotNil(s.PtrList)
	a.NotNil(s.PtrDict)
	a.NotNil(s.Child.List)
	a.NotNil(s.Child.Dict)

	// Check for data loss
	s = testStruct{
		List:    []string{"a", "b", "c"},
		Dict:    map[string]string{"a": "a"},
		PtrList: &[]string{"a", "b", "c"},
		PtrDict: &map[string]string{"a": "a"},
	}
	allocNull(&s)
	a.Len(s.List, 3)
	a.Contains(s.Dict, "a")
	a.Len(*s.PtrList, 3)
	a.Contains(*s.PtrDict, "a")
}

func TestMarshal(t *testing.T) {
	a := assert.New(t)

	// Check struct directly
	s := testStruct{}
	res := Marshal(&s)
	a.NotContains(res, "error")
	a.NotContains(res, "null")

	// Verify interface
	f := func(s testStructInterface) {
		res = Marshal(s)
		a.NotContains(res, "error")
		a.NotContains(res, "null")
	}
	s = testStruct{}
	f(&s)
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)
	s := testStruct{}
	err := Unmarshal(`{"List":["a","b","c"]}`, &s)
	a.NoError(err)
	a.Contains(s.List, "a", "b", "c")
}

// Bench external JSON Marshal function
func BenchmarkMarshalExternal(b *testing.B) {
	b.ReportAllocs()
	s := testStruct{}
	for n := 0; n < b.N; n++ {
		json.Marshal(s)
	}
}

// Bench internal JSON Marshal function
func BenchmarkMarshal(b *testing.B) {
	b.ReportAllocs()
	s := testStruct{}
	for n := 0; n < b.N; n++ {
		Marshal(&s)
	}
}
