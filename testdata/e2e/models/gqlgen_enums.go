package models

import (
	"fmt"
	"io"
	"strconv"
)

func (e MyEnum) MarshalGQL(w io.Writer) {
	switch e {
	case MyEnumValueA:
		fmt.Fprintf(w, strconv.Quote("ValueA"))
		break
	case MyEnumValueB:
		fmt.Fprintf(w, strconv.Quote("ValueB"))
		break
	}
}

func (e *MyEnum) UnmarshalGQL(v interface{}) error {
	switch v.(string) {
	case "ValueA":
		*e = MyEnumValueA
		break
	case "ValueB":
		*e = MyEnumValueB
		break
	}
	return nil
}

