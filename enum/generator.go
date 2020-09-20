package enum

import "io"

type Generator interface {
	Filename() string
	Generate(io.Writer, []EnumType) error
}
