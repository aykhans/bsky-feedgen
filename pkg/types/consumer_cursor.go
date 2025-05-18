package types

import "fmt"

type ConsumerCursor string

var (
	ConsumerCursorLastConsumed  ConsumerCursor = "last-consumed"
	ConsumerCursorFirstStream   ConsumerCursor = "first-stream"
	ConsumerCursorCurrentStream ConsumerCursor = "current-stream"
)

func (c ConsumerCursor) String() string {
	return string(c)
}

func (c ConsumerCursor) IsValid() bool {
	return c == ConsumerCursorLastConsumed || c == ConsumerCursorFirstStream || c == ConsumerCursorCurrentStream
}

func (c ConsumerCursor) Equal(other ConsumerCursor) bool {
	return c == other
}

func (c ConsumerCursor) IsLastConsumed() bool {
	return c == ConsumerCursorLastConsumed
}

func (c ConsumerCursor) IsFirstStream() bool {
	return c == ConsumerCursorFirstStream
}

func (c ConsumerCursor) IsCurrentStream() bool {
	return c == ConsumerCursorCurrentStream
}

func (c *ConsumerCursor) Set(value string) error {
	switch value {
	case ConsumerCursorLastConsumed.String(), "":
		*c = ConsumerCursorLastConsumed
	case ConsumerCursorFirstStream.String():
		*c = ConsumerCursorFirstStream
	case ConsumerCursorCurrentStream.String():
		*c = ConsumerCursorCurrentStream
	default:
		return fmt.Errorf("invalid cursor value: %s", value)
	}

	return nil
}
