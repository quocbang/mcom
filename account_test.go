package mcom

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
)

func Test_CheckPermission(t *testing.T) {
	assert := assert.New(t)
	{ // not permit.
		rs := Roles{roles.Role_ADMINISTRATOR}
		assert.ErrorIs(rs.CheckPermission(), mcomErr.Error{
			Code:    mcomErr.Code_ACCOUNT_ROLES_NOT_PERMIT,
			Details: "Administrator and Leader are not allowed to add",
		})

		rs = Roles{roles.Role_LEADER}
		assert.ErrorIs(rs.CheckPermission(), mcomErr.Error{
			Code:    mcomErr.Code_ACCOUNT_ROLES_NOT_PERMIT,
			Details: "Administrator and Leader are not allowed to add",
		})
	}

	{ // good case.
		rs := Roles{roles.Role_BEARER}
		assert.NoError(rs.CheckPermission())
	}
}
