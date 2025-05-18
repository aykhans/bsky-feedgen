package types

import "fmt"

type GeneratorCursor string

var (
	GeneratorCursorLastGenerated GeneratorCursor = "last-generated"
	GeneratorCursorFirstPost     GeneratorCursor = "first-post"
)

func (c GeneratorCursor) String() string {
	return string(c)
}

func (c GeneratorCursor) IsValid() bool {
	return c == GeneratorCursorLastGenerated || c == GeneratorCursorFirstPost
}

func (c GeneratorCursor) Equal(other GeneratorCursor) bool {
	return c == other
}

func (c GeneratorCursor) IsLastGenerated() bool {
	return c == GeneratorCursorLastGenerated
}

func (c GeneratorCursor) IsFirstPost() bool {
	return c == GeneratorCursorFirstPost
}

func (c *GeneratorCursor) Set(value string) error {
	switch value {
	case GeneratorCursorLastGenerated.String(), "":
		*c = GeneratorCursorLastGenerated
	case GeneratorCursorFirstPost.String():
		*c = GeneratorCursorFirstPost
	default:
		return fmt.Errorf("invalid cursor value: %s", value)
	}

	return nil
}
