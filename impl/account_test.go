package impl

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	commonsAccount "gitlab.kenda.com.tw/kenda/commons/v2/util/account"
	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
)

func Test_GetTokenInfo(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, _ := initializeDB(t)
	// missing token.
	{
		_, err := dm.GetTokenInfo(ctx, mcom.GetTokenInfoRequest{Token: ""})
		assert.ErrorIs(err, mcomErr.Error{
			Code:    mcomErr.Code_USER_UNKNOWN_TOKEN,
			Details: "missing token",
		})
	}
	// failed to get token.
	{
		var e *session
		monkey.PatchInstanceMethod(reflect.TypeOf(e), "GetToken", func(session *session, id string) (*models.Token, error) {
			return &models.Token{}, fmt.Errorf("failed to get token")
		})

		_, err := dm.GetTokenInfo(ctx, mcom.GetTokenInfoRequest{Token: "123456"})
		monkey.UnpatchAll()

		assert.EqualError(err, "failed to get token")
	}
	// normal.
	{
		var e *session
		monkey.PatchInstanceMethod(reflect.TypeOf(e), "GetToken", func(session *session, id string) (*models.Token, error) {
			return &models.Token{
				ID:          models.EncryptedData("token"),
				BoundUser:   "tester",
				ExpiryTime:  123456789,
				CreatedTime: 123456789,
				Valid:       true,
				Info: models.UserInfo{
					Roles: []roles.Role{roles.Role_INSPECTOR, roles.Role_QUALITY_CONTROLLER},
				},
			}, nil
		})

		user, err := dm.GetTokenInfo(ctx, mcom.GetTokenInfoRequest{Token: "token"})
		monkey.UnpatchAll()

		assert.NoError(err)
		assert.Equal(mcom.GetTokenInfoReply{
			User:        "tester",
			Valid:       true,
			ExpiryTime:  time.Unix(0, 123456789),
			CreatedTime: time.Unix(0, 123456789),
			Roles:       []roles.Role{roles.Role_INSPECTOR, roles.Role_QUALITY_CONTROLLER},
		}, user)
	}
}

func Test_Encryption(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)

	assert.NoError(dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
		mcom.CreateAccountRequest{
			ID:    testUser,
			Roles: []roles.Role{roles.Role_INSPECTOR},
		}.WithSpecifiedPassword(testPassword)}))

	var account models.Account
	assert.NoError(db.Take(&account).Error)
	assert.True(bytes.Equal(account.Password, encryptPW(testPassword)))
	assert.NoError(dm.DeleteAccount(ctx, mcom.DeleteAccountRequest{ID: testUser}))
}

func encryptPW(pw string) models.EncryptedData {
	return models.Encrypt([]byte(pw))
}

func deleteAllSignData(db *gorm.DB) error {
	return newClearMaster(db, &models.Account{}, &models.User{}, &models.Token{}, &models.Department{}).Clear()
}

func Test_Sign(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	opt := mcom.WithTokenExpiredAfter(8 * time.Hour)

	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}
	defer dm.Close()

	assert.NoError(deleteAllSignData(db))

	assert.NoError(dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
		mcom.CreateAccountRequest{
			ID:    testUser,
			Roles: []roles.Role{roles.Role_INSPECTOR},
		}.WithSpecifiedPassword(testPassword)}))

	if !assert.NoError(err) {
		return
	}

	{ // SignOut: insufficient request: token.
		err := dm.SignOut(ctx, mcom.SignOutRequest{})
		assert.ErrorIs(err, mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		})
	}
	{ // SignOut: user unknown token.
		err := dm.SignOut(ctx, mcom.SignOutRequest{Token: "123456"})
		assert.ErrorIs(err, mcomErr.Error{
			Code: mcomErr.Code_USER_UNKNOWN_TOKEN,
		})
	}
	{ // SignIn: missing id or password.
		_, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  "",
			Password: "",
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
			Details: "missing id or password",
		}, err)
	}
	{ // SignIn: account not found.
		_, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  "NotFound",
			Password: testPassword,
			ADUser:   false,
		})
		assert.ErrorIs(err, mcomErr.Error{
			Code:    mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
			Details: "account not found",
		})
	}
	{ // SignIn: missing password.
		_, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  testBadUser,
			Password: "",
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
			Details: "missing id or password",
		}, err)
	}
	{ // SignIn: bad password.
		_, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  testUser,
			Password: testBadPassword,
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
			Details: "wrong password",
		}, err)
	}
	{ // SignIn: user not found.
		_, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  testUser,
			Password: testPassword,
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD})
	}
	{ // CreateDepartment: test data generation.
		err := dm.CreateDepartments(ctx, []string{testDepartmentID, testDepartmentA, testDepartmentB, testDepartmentC})
		assert.NoError(err)
	}
	{ // CreateUser: test data generation.
		err = dm.CreateUsers(ctx, mcom.CreateUsersRequest{
			Users: []mcom.User{
				{
					ID:           testUser,
					DepartmentID: testDepartmentID,
				},
			},
		})
		assert.NoError(err)
	}
	{ // SignIn/Out: good case.
		reply, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  testUser,
			Password: testPassword,
		}, opt)
		assert.NoError(err)

		token := models.Token{}
		err = db.Where(` bound_user = ? `, testUser).Take(&token).Error
		assert.NoError(err)

		deps := []models.Department{}
		err = db.Where("id LIKE ? ", strings.TrimRight(testDepartmentID, "0")+"%").
			Order("id").
			Find(&deps).Error
		assert.NoError(err)

		departments := make([]mcom.Department, len(deps))
		for i, v := range deps {
			departments[i] = mcom.Department{
				OID: v.ID,
				ID:  v.ID,
			}
		}
		assert.Equal([]byte(token.ID), models.Encrypt([]byte(reply.Token)))
		assert.Equal(departments, reply.Departments)

		err = dm.SignOut(ctx, mcom.SignOutRequest{Token: reply.Token})
		assert.True(reply.MustChangePassword)
		assert.NoError(err)
	}
	{ // AD user SignIn Auth error.
		var e *commonsAccount.ADAgent
		monkey.PatchInstanceMethod(reflect.TypeOf(e), "Auth", func(agent *commonsAccount.ADAgent, id, pwd string) error {
			return fmt.Errorf("user not found")
		})

		_, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  testADUser,
			Password: testPassword,
			ADUser:   true,
		}, opt)

		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD})
		monkey.UnpatchAll()
	}
	{ // AD user SignIn SearchUserRoles error.
		var e *commonsAccount.ADAgent
		monkey.PatchInstanceMethod(reflect.TypeOf(e), "Auth", func(agent *commonsAccount.ADAgent, id, pwd string) error {
			return nil
		})
		monkey.PatchInstanceMethod(reflect.TypeOf(e), "SearchUserRoles", func(agent *commonsAccount.ADAgent, id string) ([]string, error) {
			return nil, fmt.Errorf("roles not found")
		})

		_, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  testADUser,
			Password: testPassword,
			ADUser:   true,
		}, opt)

		assert.EqualError(err, "roles not found")
		monkey.UnpatchAll()
	}
	{ // AD user SignIn: failed to get the department of the user.
		var e *commonsAccount.ADAgent
		monkey.PatchInstanceMethod(reflect.TypeOf(e), "Auth", func(agent *commonsAccount.ADAgent, id, pwd string) error {
			return nil
		})
		monkey.PatchInstanceMethod(reflect.TypeOf(e), "SearchUserRoles", func(agent *commonsAccount.ADAgent, id string) ([]string, error) {
			return []string{"Planner", "Scheduler"}, nil
		})

		_, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  testADUser,
			Password: testPassword,
			ADUser:   true,
		}, opt)
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD})
		monkey.UnpatchAll()
	}
	{ // AD user SignIn/Out: good case.
		err = dm.CreateUsers(ctx, mcom.CreateUsersRequest{
			Users: []mcom.User{{
				ID:           testADUser,
				Account:      testADUser,
				DepartmentID: testDepartmentID,
			}},
		})
		assert.NoError(err)

		var e *commonsAccount.ADAgent
		monkey.PatchInstanceMethod(reflect.TypeOf(e), "Auth", func(agent *commonsAccount.ADAgent, id, pwd string) error {
			return nil
		})
		monkey.PatchInstanceMethod(reflect.TypeOf(e), "SearchUserRoles", func(agent *commonsAccount.ADAgent, id string) ([]string, error) {
			return []string{
				roles.Role_PLANNER.String(),
				roles.Role_SCHEDULER.String(),
			}, nil
		})

		reply, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  testADUser,
			Password: testPassword,
			ADUser:   true,
		}, opt)
		assert.NoError(err)
		monkey.UnpatchAll()

		token := models.Token{}
		err = db.Where(` bound_user = ? `, testADUser).Take(&token).Error
		assert.NoError(err)

		deps := []models.Department{}
		err = db.Where("id LIKE ? ", strings.TrimRight(testDepartmentID, "0")+"%").
			Order("id").
			Find(&deps).Error
		assert.NoError(err)

		departments := make([]mcom.Department, len(deps))
		for i, v := range deps {
			departments[i] = mcom.Department{
				OID: v.ID,
				ID:  v.ID,
			}
		}
		assert.Equal([]byte(token.ID), models.Encrypt([]byte(reply.Token)))
		assert.Equal(departments, reply.Departments)
		assert.Equal([]roles.Role{roles.Role_PLANNER, roles.Role_SCHEDULER}, reply.Roles)

		err = dm.SignOut(ctx, mcom.SignOutRequest{Token: reply.Token})
		assert.NoError(err)
	}
	{ // UpdateUser: good case, update leave date.
		leaveDate, err := time.Parse(dateLayout, "1911-01-01")
		assert.NoError(err)
		leaveDate = leaveDate.Local()

		err = dm.UpdateUser(ctx, mcom.UpdateUserRequest{
			ID:        testUser,
			LeaveDate: leaveDate,
		})
		assert.NoError(err)

		// check result.
		dep := &models.Department{}
		err = db.Where(` id = ? `, testDepartmentID).Take(dep).Error
		assert.NoError(err)

		userInfo := &models.User{}
		err = db.Where(` id = ? and department_id = ? `, testUser, dep.ID).Take(userInfo).Error
		assert.NoError(err)

		// set to the same timezone for the comparison.
		userInfo.LeaveDate.Time = userInfo.LeaveDate.Time.Local()
		assert.Equal(&models.User{
			ID:           testUser,
			DepartmentID: dep.ID,
			LeaveDate: sql.NullTime{
				Time:  leaveDate,
				Valid: true,
			},
		}, userInfo)
	}
	{ // UpdaterUser : department not found.
		assert.ErrorIs(dm.UpdateUser(ctx, mcom.UpdateUserRequest{
			ID:           testUser,
			DepartmentID: "Not Found",
		}), mcomErr.Error{Code: mcomErr.Code_DEPARTMENT_NOT_FOUND})
	}
	{ // SignIn: user resign.
		opt := mcom.WithTokenExpiredAfter(8 * time.Hour)
		_, err := dm.SignIn(ctx, mcom.SignInRequest{
			Account:  testUser,
			Password: testPassword,
		}, opt)
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD})
		assert.Equal(mcomErr.Error{
			Code: mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
		}, err)
	}

	assert.NoError(deleteAllSignData(db))
}

func deleteAllDepartmentsData(db *gorm.DB) error {
	return newClearMaster(db, &models.Department{}).Clear()
}

func deleteAllAccount(db *gorm.DB) error {
	return newClearMaster(db, &models.Account{}).Clear()
}

func Test_ChangeUserPassword(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}

	defer dm.Close()
	assert.NoError(deleteAllAccount(db))

	{ // insufficient request: id.
		err := dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
			ChangePassword: &struct {
				NewPassword string
				OldPassword string
			}{
				NewPassword: "123123",
				OldPassword: "456456",
			}})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // insufficient request: old password.
		ctx = commonsCtx.WithUserID(context.Background(), testUser)
		err := dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
			UserID: "user",
			ChangePassword: &struct {
				NewPassword string
				OldPassword string
			}{NewPassword: "123123", OldPassword: ""},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // insufficient request: new password.
		err := dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
			UserID: "user",
			ChangePassword: &struct {
				NewPassword string
				OldPassword string
			}{NewPassword: "", OldPassword: "456456"},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // user not found or bad password.
		err := dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
			UserID: "user",
			ChangePassword: &struct {
				NewPassword string
				OldPassword string
			}{NewPassword: "123123", OldPassword: "456456"},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_ACCOUNT_BAD_OLD_PASSWORD,
		}, err)
	}
	{ // update with the same password.
		assert.NoError(dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{mcom.CreateAccountRequest{
			ID:    testUser,
			Roles: []roles.Role{roles.Role_BEARER},
		}.WithSpecifiedPassword("123123")}))
		assert.NoError(err)
		err = dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
			UserID: testUser,
			ChangePassword: &struct {
				NewPassword string
				OldPassword string
			}{NewPassword: "123123", OldPassword: "123123"},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_ACCOUNT_SAME_AS_OLD_PASSWORD,
		}, err)
		assert.NoError(deleteAllAccount(db))
	}
	{ // update with the wrong password.
		err := dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
			UserID: testUser,
			ChangePassword: &struct {
				NewPassword string
				OldPassword string
			}{NewPassword: "987987", OldPassword: "654654"},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_ACCOUNT_BAD_OLD_PASSWORD,
		}, err)
	}
	{ // good case.
		// add mock user data.
		assert.NoError(dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				ID:    testUser,
				Roles: []roles.Role{roles.Role_BEARER},
			}.WithSpecifiedPassword("123123")}))
		err = dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
			UserID: testUser,
			ChangePassword: &struct {
				NewPassword string
				OldPassword string
			}{NewPassword: "456456", OldPassword: "123123"},
		})
		assert.NoError(err)

		// check result.
		result := &models.Account{}
		err = db.Where(`id = ?`, testUser).Find(result).Error
		assert.NoError(err)
		assert.Equal(&models.Account{
			ID:       testUser,
			Password: encryptPW("456456"),
			Roles:    []int64{int64(roles.Role_BEARER)},
		}, result)
	}

	assert.NoError(deleteAllAccount(db))
}

func Test_ListRoles(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	dm, _, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}

	defer dm.Close()

	{ // good case.
		result, err := dm.ListRoles(ctx)
		assert.NoError(err)
		expected := []mcom.Role{
			{Name: "ROLE_UNSPECIFIED", Value: roles.Role_ROLE_UNSPECIFIED},
			{Name: "ADMINISTRATOR", Value: roles.Role_ADMINISTRATOR},
			{Name: "LEADER", Value: roles.Role_LEADER},
			{Name: "PLANNER", Value: roles.Role_PLANNER},
			{Name: "SCHEDULER", Value: roles.Role_SCHEDULER},
			{Name: "INSPECTOR", Value: roles.Role_INSPECTOR},
			{Name: "QUALITY_CONTROLLER", Value: roles.Role_QUALITY_CONTROLLER},
			{Name: "OPERATOR", Value: roles.Role_OPERATOR},
			{Name: "BEARER", Value: roles.Role_BEARER},
		}
		assert.Equal(len(expected), len(result.Roles))
		for i := range expected {
			assert.Equal(expected[i], result.Roles[i])
		}
	}
}

func Test_CreateAccounts(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}

	defer dm.Close()
	assert.NoError(deleteAllAccount(db))
	{ // CreateAccounts: insufficient request.
		err := dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "CreateAccounts: empty request",
		}, err)
	}
	{ // CreateAccounts: empty user id.
		err := dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				Roles: mcom.Roles{roles.Role_BEARER},
			},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // CreateAccounts: nil roles.
		err := dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				ID: "123123",
			},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // CreateAccounts: good case, add new user role.
		assert.NoError(dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				ID:    testUserA,
				Roles: mcom.Roles{roles.Role_BEARER},
			}.WithDefaultPassword(),
			mcom.CreateAccountRequest{
				ID:    testUserB,
				Roles: mcom.Roles{roles.Role_OPERATOR},
			}.WithDefaultPassword(),
		}))

		account := []models.Account{}
		err = db.Where(`id IN ?`, []string{testUserA, testUserB}).Find(&account).Error
		assert.NoError(err)

		expected := []models.Account{
			{
				ID:                 testUserA,
				Password:           encryptPW(testUserA),
				Roles:              []int64{int64(roles.Role_BEARER)},
				MustChangePassword: true,
			}, {
				ID:                 testUserB,
				Password:           encryptPW(testUserB),
				Roles:              []int64{int64(roles.Role_OPERATOR)},
				MustChangePassword: true,
			},
		}
		assert.ElementsMatch(expected, account)
	}
	{ // CreateAccounts: bad case, user already exists.
		assert.ErrorIs(dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				ID:    testUserA,
				Roles: mcom.Roles{roles.Role_INSPECTOR},
			}.WithDefaultPassword(),
			mcom.CreateAccountRequest{
				ID:    testUserB,
				Roles: mcom.Roles{roles.Role_OPERATOR, roles.Role_BEARER},
			}.WithDefaultPassword(),
		}), mcomErr.Error{Code: mcomErr.Code_ACCOUNT_ALREADY_EXISTS})
	}
	{ // UpdateAccount: good case, update user role.
		assert.NoError(dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
			UserID: testUserA,
			Roles:  []roles.Role{roles.Role_INSPECTOR},
		}))
		assert.NoError(dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
			UserID: testUserB,
			Roles:  []roles.Role{roles.Role_OPERATOR, roles.Role_BEARER},
		}))

		Account := []models.Account{}
		err = db.Where(`id IN ?`, []string{testUserA, testUserB}).Find(&Account).Error
		assert.NoError(err)

		expected := []models.Account{
			{
				ID:                 testUserA,
				Password:           encryptPW(testUserA),
				Roles:              []int64{int64(roles.Role_INSPECTOR)},
				MustChangePassword: true,
			}, {
				ID:                 testUserB,
				Password:           encryptPW(testUserB),
				Roles:              []int64{int64(roles.Role_OPERATOR), int64(roles.Role_BEARER)},
				MustChangePassword: true,
			},
		}
		assert.Equal(expected, Account)
	}
	{ // UpdateAccount: update an empty roles slice (not allowed)
		assert.ErrorIs(dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{
			UserID: testUserB,
			Roles:  []roles.Role{},
		}), mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the request did not match any conditions of UpdateAccount",
		})
		var account models.Account
		assert.NoError(db.Where(&models.Account{ID: testUserB}).Take(&account).Error)
		assert.NotEqual(pq.Int64Array{}, account.Roles)
	}
	{ // CreateAccounts: bad case , add Administrator role.
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_ACCOUNT_ROLES_NOT_PERMIT,
			Details: "Administrator and Leader are not allowed to add",
		}, dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				ID:    testUserA,
				Roles: mcom.Roles{roles.Role_ADMINISTRATOR},
			}.WithDefaultPassword(),
		}))
	}
	{ // CreateAccounts: good case, add new accounts with specifiedPassword.
		assert.NoError(dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				ID:    "testUserC",
				Roles: mcom.Roles{roles.Role_BEARER},
			}.WithSpecifiedPassword("hahaha"),
			mcom.CreateAccountRequest{
				ID:    "testUserD",
				Roles: mcom.Roles{roles.Role_OPERATOR},
			}.WithDefaultPassword(),
		}))

		account := []models.Account{}
		err = db.Where(`id IN ?`, []string{"testUserC", "testUserD"}).Find(&account).Error
		assert.NoError(err)

		expected := []models.Account{
			{
				ID:                 "testUserC",
				Password:           encryptPW("hahaha"),
				Roles:              []int64{int64(roles.Role_BEARER)},
				MustChangePassword: true,
			}, {
				ID:                 "testUserD",
				Password:           encryptPW("testUserD"),
				Roles:              []int64{int64(roles.Role_OPERATOR)},
				MustChangePassword: true,
			},
		}
		assert.Equal(expected, account)
	}

	assert.NoError(deleteAllAccount(db))
}

func Test_ListUserRoles(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}

	defer dm.Close()
	assert.NoError(deleteAllAccount(db))
	assert.NoError(deleteAllDepartmentsData(db))
	assert.NoError(deleteAllUser(db))
	{ // ListUserRoles: insufficient request.
		_, err := dm.ListUserRoles(ctx, mcom.ListUserRolesRequest{DepartmentOID: ""})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "empty departmentID",
		}, err)
	}
	{ // ListUserRoles: department not found.
		_, err := dm.ListUserRoles(ctx, mcom.ListUserRolesRequest{DepartmentOID: testDepartmentID})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_DEPARTMENT_NOT_FOUND,
		}, err)
	}
	{ // ListUserRoles: good case.
		err := dm.CreateDepartments(ctx, []string{testDepartmentID})
		assert.NoError(err)

		err = dm.CreateUsers(ctx, mcom.CreateUsersRequest{
			Users: []mcom.User{{
				ID:           testUserA,
				DepartmentID: testDepartmentID,
			}, {
				ID:           testUserB,
				DepartmentID: testDepartmentID,
			}},
		})
		assert.NoError(err)

		err = db.Create([]*models.Account{
			{
				ID:       testUserA,
				Password: encryptPW(testPassword),
				Roles:    []int64{int64(roles.Role_INSPECTOR)},
			}, {
				ID:       testUserB,
				Password: encryptPW(testPassword),
				Roles:    []int64{int64(roles.Role_OPERATOR), int64(roles.Role_BEARER)},
			},
		}).Error
		assert.NoError(err)

		results, err := dm.ListUserRoles(ctx, mcom.ListUserRolesRequest{DepartmentOID: testDepartmentID})
		assert.NoError(err)
		assert.Equal(mcom.ListUserRolesReply{
			Users: []mcom.UserRoles{{
				ID:    testUserA,
				Roles: []roles.Role{roles.Role_INSPECTOR},
			}, {
				ID:    testUserB,
				Roles: []roles.Role{roles.Role_OPERATOR, roles.Role_BEARER},
			}},
		}, results)
	}

	assert.NoError(deleteAllAccount(db))
	assert.NoError(deleteAllDepartmentsData(db))
	assert.NoError(deleteAllUser(db))
}

func Test_ResetUserPassword(t *testing.T) {
	assert := assert.New(t)
	ctx := commonsCtx.WithUserID(context.Background(), testUser)

	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}

	defer dm.Close()
	assert.NoError(deleteAllAccount(db))
	{ // insufficient request.
		err = dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{}, mcom.ResetPassword())
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // user not found.
		err = dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{UserID: testUser}, mcom.ResetPassword())
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_ACCOUNT_NOT_FOUND,
		}, err)
	}
	{ // good case.
		assert.NoError(dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				ID:    testUser,
				Roles: []roles.Role{roles.Role_BEARER},
			}.WithSpecifiedPassword("987654")}))

		err = dm.UpdateAccount(ctx, mcom.UpdateAccountRequest{UserID: testUser}, mcom.ResetPassword())
		assert.NoError(err)

		// check result.
		userInfo := &models.Account{}
		err = db.Where(`id=?`, testUser).Find(userInfo).Error
		assert.NoError(err)
		assert.Equal(userInfo, &models.Account{
			ID:                 testUser,
			Password:           encryptPW(testUser),
			Roles:              []int64{int64(roles.Role_BEARER)},
			MustChangePassword: true,
		})
	}

	assert.NoError(deleteAllAccount(db))
}

func Test_DeleteAccount(t *testing.T) {
	assert := assert.New(t)
	ctx := commonsCtx.WithUserID(context.Background(), testUser)

	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}

	defer dm.Close()
	{ // insufficient request.
		err = dm.DeleteAccount(ctx, mcom.DeleteAccountRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing user id",
		}, err)
	}
	{ // user not found.
		err = dm.DeleteAccount(ctx, mcom.DeleteAccountRequest{ID: testUser})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_ACCOUNT_NOT_FOUND,
		}, err)
	}
	{ // good case.
		assert.NoError(dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				ID:    testUser,
				Roles: []roles.Role{roles.Role_BEARER},
			}.WithSpecifiedPassword("123456")}))

		err = dm.DeleteAccount(ctx, mcom.DeleteAccountRequest{ID: testUser})
		assert.NoError(err)

		// check result.
		userInfo := &models.Account{}
		err = db.Where(`id=?`, testUser).Take(userInfo).Error
		assert.ErrorIs(gorm.ErrRecordNotFound, err)
	}
}
