package testbuilder

import (
	"context"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
)

type testUserBuilder struct {
	userID          string
	account         string
	departmentID    string
	roles           mcom.Roles
	buildAccount    bool
	buildDepartment bool
}

func (tub testUserBuilder) GetInfo() (userName string, account string, departmentID string) {
	return tub.userID, tub.account, tub.departmentID
}

func NewTestUserBuilder() *testUserBuilder {
	res := testUserBuilder{}
	res.Reset()
	return &res
}

func (tub *testUserBuilder) Reset() {
	tub.userID = "testUser"
	tub.account = "testAccount"
	tub.departmentID = "testDepartment"
	tub.buildAccount = true
	tub.buildDepartment = true
	tub.roles = mcom.Roles{roles.Role_OPERATOR}
}

func (tub *testUserBuilder) WithUserName(userName string) *testUserBuilder {
	tub.userID = userName
	return tub
}

func (tub *testUserBuilder) WithAccount(account string) *testUserBuilder {
	tub.account = account
	return tub
}

func (tub *testUserBuilder) WithDepartmentID(departmentID string) *testUserBuilder {
	tub.departmentID = departmentID
	return tub
}

func (tub *testUserBuilder) WithRoles(roles mcom.Roles) *testUserBuilder {
	tub.roles = roles
	return tub
}

func (tub *testUserBuilder) DoNotBuildAccount() *testUserBuilder {
	tub.buildAccount = false
	return tub
}

func (tub *testUserBuilder) DoNotBuildDepartment() *testUserBuilder {
	tub.buildDepartment = false
	return tub
}

// build department, user, account.
func (tub testUserBuilder) Build(dm mcom.DataManager) error {
	ctx := context.Background()
	if tub.buildDepartment {
		if err := dm.CreateDepartments(ctx, mcom.CreateDepartmentsRequest{tub.departmentID}); err != nil {
			return err
		}
	}

	if err := dm.CreateUsers(ctx, mcom.CreateUsersRequest{
		Users: []mcom.User{
			{
				ID:           tub.userID,
				Account:      tub.account,
				DepartmentID: tub.departmentID,
			},
		},
	}); err != nil {
		return err
	}

	if tub.buildAccount {
		if err := dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				ID:    tub.account,
				Roles: tub.roles,
			}.WithDefaultPassword(),
		}); err != nil {
			return err
		}
	}

	return nil
}
