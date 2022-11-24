package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_scanBadTypeError_Error(t *testing.T) {
	assert := assert.New(t)

	assert.EqualError(&scanBadTypeError{
		structName: "test",
		src:        "???",
	}, "bad src type [string] for struct [test]")
}
