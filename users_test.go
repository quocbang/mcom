package mcom

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSignInOptions(t *testing.T) {
	assert := assert.New(t)

	const (
		expiredAfter = 10 * time.Hour
	)

	// with options.
	{
		o := ParseSignInOptions([]SignInOption{
			WithTokenExpiredAfter(expiredAfter),
		})
		assert.Equal(SignInOptions{
			ExpiredAfter: expiredAfter,
		}, o)
	}
	// no option.
	{
		expectedOptions := SignInOptions{
			ExpiredAfter: 8 * time.Hour,
		}
		assert.Equal(expectedOptions, ParseSignInOptions([]SignInOption{}))
		assert.Equal(expectedOptions, ParseSignInOptions(nil))
	}
}

func Test_ParseListUnauthorizedUsersOptions(t *testing.T) {
	assert := assert.New(t)
	{ // case no option.
		actual := ParseListUnauthorizedUsersOptions([]ListUnauthorizedUsersOption{})
		expect := ListUnauthorizedUsersOptions{}
		assert.Equal(expect, actual)
	}
	{ // case with option.
		actual := ParseListUnauthorizedUsersOptions([]ListUnauthorizedUsersOption{ExcludeUsers([]string{"Rockefeller"})})
		expect := ListUnauthorizedUsersOptions{
			ExcludeUsers: []string{"Rockefeller"},
		}
		assert.Equal(expect, actual)
	}
}

func Test_ParseSignInStationOptions(t *testing.T) {
	assert := assert.New(t)
	{ // default.
		opts := ParseSignInStationOptions([]SignInStationOption{})
		assert.True(opts.VerifyWorkDate(time.Now())) // always true.
	}
	{ // WithVerifyWorkDateHandler.
		opts := ParseSignInStationOptions([]SignInStationOption{WithVerifyWorkDateHandler(func(t time.Time) bool {
			return t == time.Date(2022, 2, 22, 22, 22, 22, 22, time.Local)
		})})
		assert.False(opts.VerifyWorkDate(time.Now()))
		assert.True(opts.VerifyWorkDate(time.Date(2022, 2, 22, 22, 22, 22, 22, time.Local)))
	}
}
