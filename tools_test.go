package mcom

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func Test_ToolResource(t *testing.T) {
	assert := assert.New(t)

	{ // for an empty resource.
		tr := ToolResource{}
		assert.NoError(tr.Empty())
		assert.ErrorIs(tr.Require(), mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "Key: 'ToolResource.ResourceID' Error:Field validation for 'ResourceID' failed on the 'required' tag",
		})
	}
	{ // for a not empty resource.
		tr := ToolResource{
			ResourceID:  "resource",
			ToolID:      "tool",
			BindingSite: models.UniqueSite{},
		}

		assert.NoError(tr.Require())
		assert.ErrorIs(tr.Empty(), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "there should be no resources while clearing a slot.",
		})
	}
}

func Test_ToModelResource(t *testing.T) {
	assert := assert.New(t)
	tr := ToolResource{}
	assert.Equal(tr.ToModelResource(), models.ToolResource{})
}
