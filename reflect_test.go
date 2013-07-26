package toml

import (
	"fmt"
	"reflect"
	"testing"
)

type MyStruct struct {
	Ranking []string
	Bands   map[string]Primitive
}

func TestReflect(t *testing.T) {
	var thestruct MyStruct

	s := reflect.ValueOf(thestruct)
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%d: %s %s = %v\n", i,
			typeOfT.Field(i).Name,
			f.Type(),
			f.Interface())
	}
}
