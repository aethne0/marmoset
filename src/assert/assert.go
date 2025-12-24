package assert

import (
	"cmp"
	"fmt"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

type pair struct {
	label string
	value any // maybe theres a Display/Debug trait
}

var titleStyle = color.New(color.Bold, color.FgBlack, color.BgHiRed)
var typeStyle = color.New(color.Italic, color.FgHiRed)
var keyStyle = color.New(color.FgYellow)
var valueStyle = color.New(color.FgHiBlue)

func makeMsg(typeString string, msg string, pairs []pair) string {
	_, file, line, _ := runtime.Caller(2)

	defaultPairs := []pair{
		{label: "Where", value: fmt.Sprintf("%s:%d", file, line)},
		{label: "Message", value: fmt.Sprintf("%s", msg)},
	}

	defaultPairs = append(defaultPairs, pairs...)

	maxLen := 0
	for _, p := range defaultPairs {
		maxLen = max(maxLen, len(p.label))
	}

	var s strings.Builder
	s.WriteString(titleStyle.Sprintf("Assert error:"))
	s.WriteString(" ")
	s.WriteString(typeStyle.Sprintf("%s\n", typeString))

	for _, p := range defaultPairs {
		s.WriteString(keyStyle.Sprintf("%s:", p.label))
		s.WriteString(strings.Repeat(" ", 5+maxLen-len(p.label)))
		s.WriteString(valueStyle.Sprintf("%#v\n", p.value))
	}

	return s.String()

}

func Unreachable(msg string) {
	panic(makeMsg("Unreachable", msg, []pair{}))
}

func True(condition bool, msg string) {
	if !condition {
		panic(makeMsg("True", msg, []pair{}))
	}
}

func False(condition bool, msg string) {
	if condition {
		panic(makeMsg("False", msg, []pair{}))
	}
}

func Eq[T comparable](a T, b T, msg string) {
	if a != b {
		panic(makeMsg("left == right", msg, []pair{
			{label: "Left", value: a},
			{label: "Right", value: b},
		}))
	}
}

func Ne[T comparable](a T, b T, msg string) {
	if a == b {
		panic(makeMsg("left != right", msg, []pair{
			{label: "Left", value: a},
			{label: "Right", value: b},
		}))
	}
}

func Gt[T cmp.Ordered](a T, b T, msg string) {
	if a > b {
		panic(makeMsg("left > right", msg, []pair{
			{label: "Left", value: a},
			{label: "Right", value: b},
		}))
	}
}

func Lt[T cmp.Ordered](a T, b T, msg string) {
	if a < b {
		panic(makeMsg("left < right", msg, []pair{
			{label: "Left", value: a},
			{label: "Right", value: b},
		}))
	}
}

func Ge[T cmp.Ordered](a T, b T, msg string) {
	if a >= b {
		panic(makeMsg("left >= right", msg, []pair{
			{label: "Left", value: a},
			{label: "Right", value: b},
		}))
	}
}

func Le[T cmp.Ordered](a T, b T, msg string) {
	if a <= b {
		panic(makeMsg("left <= right", msg, []pair{
			{label: "Left", value: a},
			{label: "Right", value: b},
		}))
	}
}

func Todo(msg ...string) {
	if len(msg) > 0 {
		panic(makeMsg("Todo!", msg[0], []pair{}))
	} else {
		panic(makeMsg("Todo!", "(unimplemented todo)", []pair{}))
	}
}
