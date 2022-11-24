package models

import (
	"database/sql"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
)

func TestTokenInfo_implementation(t *testing.T) {
	assert := assert.New(t)

	// good.
	{
		expected := UserInfo{
			Roles: []roles.Role{
				roles.Role_OPERATOR,
				roles.Role_BEARER,
			},
		}
		v, err := expected.Value()
		assert.NoError(err)

		var s UserInfo
		assert.NoError(s.Scan(v))
		assert.Equal(expected, s)
	}
}

func TestUser_Resigned(t *testing.T) {
	assert := assert.New(t)

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2021, time.September, 20, 8, 0, 0, 0, time.Local)
	})
	defer monkey.UnpatchAll()

	// the user is in-service staff.
	{
		// no leave date.
		{
			user := User{}
			assert.False(user.Resigned())
		}
		// specified leave date.
		{
			user := User{
				LeaveDate: sql.NullTime{
					Time:  time.Date(2021, time.September, 21, 0, 0, 0, 0, time.Local),
					Valid: true,
				},
			}
			assert.False(user.Resigned())
		}
	}
	// the user has resigned.
	{
		user := User{
			LeaveDate: sql.NullTime{
				Time:  time.Date(2021, time.September, 20, 0, 0, 0, 0, time.Local),
				Valid: true,
			},
		}
		assert.True(user.Resigned())
	}
}
