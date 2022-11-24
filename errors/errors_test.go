package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	assert := assert.New(t)

	code, details := Code_DEPARTMENT_NOT_FOUND, "test error"
	e := Error{
		Code:    code,
		Details: details,
	}
	assert.EqualError(e, fmt.Sprintf("%s: %s", Code_name[int32(code)], details))

	// without details.
	e = Error{
		Code: code,
	}
	assert.EqualError(e, Code_name[int32(code)])
}

func TestAs(t *testing.T) {
	assert := assert.New(t)

	// a Error type.
	{
		expectedErr := Error{Code: Code_NONE}
		var err error = expectedErr
		e, ok := As(err)
		assert.True(ok)
		assert.Equal(expectedErr, e)

		werr := fmt.Errorf("wrap it: %w", err)
		e, ok = As(werr)
		assert.True(ok)
		assert.Equal(expectedErr, e)

		var perr error = &expectedErr
		e, ok = As(perr)
		assert.True(ok)
		assert.Equal(expectedErr, e)

		wperr := fmt.Errorf("wrap it: %w", perr)
		e, ok = As(wperr)
		assert.True(ok)
		assert.Equal(expectedErr, e)
	}
	// not a Error type.
	{
		err := fmt.Errorf("test error")
		e, ok := As(err)
		assert.False(ok)
		assert.Empty(e)

		werr := fmt.Errorf("wrap it: %w", err)
		e, ok = As(werr)
		assert.False(ok)
		assert.Empty(e)
	}
}
