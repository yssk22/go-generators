package helper

import (
	"fmt"
	"reflect"
	"strings"
)

func ParseFieldTag(tagName string, tagValue string) ([]string, error) {
	st := reflect.StructTag(tagValue + ` _gofix:"_magic"`)
	if st.Get("_gofix") != "_magic" {
		return nil, fmt.Errorf("struct field tag %s in %s not compatible with reflect.StructTag.Get", tagValue, tagName)
	}
	return strings.Split(st.Get(tagName), ","), nil
}
