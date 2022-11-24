package mock

import (
	"encoding/json"
	"fmt"
)

type badRequestError struct {
	expected, actual string
}

func newBadRequestError(
	expected,
	actual interface{},
) *badRequestError {
	serialExpected, _ := json.Marshal(expected)
	serialActual, _ := json.Marshal(actual)
	return &badRequestError{expected: string(serialExpected), actual: string(serialActual)}
}

func (b *badRequestError) Error() string {
	return fmt.Sprintf("request is different, expected: %s, actual: %s", b.expected, b.actual)
}

func (b *badRequestError) ErrorWithStep(step uint) error {
	return fmt.Errorf("request is different in step-%d, expected: %s, actual: %s", step, b.expected, b.actual)
}

type mismatchInputOptionLengthError struct {
	expected, actual int
}

func newMismatchInputOptionLengthError(
	expected,
	actual int,
) *mismatchInputOptionLengthError {
	return &mismatchInputOptionLengthError{expected: expected, actual: actual}
}

func (m *mismatchInputOptionLengthError) Error() string {
	return fmt.Sprintf("input options length mismatch, expected: %d, actual: %d", m.expected, m.actual)
}

type badInputOptionError struct {
	expected, actual interface{}
}

func newBadInputOptionError(
	expected,
	actual interface{},
) *badInputOptionError {
	return &badInputOptionError{expected: expected, actual: actual}
}

func (b *badInputOptionError) Error() string {
	return fmt.Sprintf("input options are different, expected: %#v, actual: %#v", b.expected, b.actual)
}

func (b *badInputOptionError) ErrorWithStep(step uint) error {
	return fmt.Errorf("input options are different in step-%d, expected: %#v, actual: %#v", step, b.expected, b.actual)
}

type badResponseError struct {
	step uint
}

func newBadResponseError(step uint) *badResponseError {
	return &badResponseError{step: step}
}

func (b *badResponseError) Error() string {
	return fmt.Sprintf("response type error in step-%d", b.step)
}

func badOptionType(expectedType string) error {
	return fmt.Errorf("bad script, expected option type: %s", expectedType)
}
