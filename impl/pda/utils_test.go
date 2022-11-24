package pda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootStringConvert(t *testing.T) {
	assert := assert.New(t)
	{
		result := RootStringConvert("<root>&lt;test&gt;&lt;/test&gt;</root>")
		assert.Equal("<root><test></test></root>", result)
	}
}
