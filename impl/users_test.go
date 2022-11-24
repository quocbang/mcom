package impl

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	testDataBuilder "gitlab.kenda.com.tw/kenda/mcom/impl/testbuilder"
	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
)

const (
	testUser         = "123456"
	testUserA        = "A12345"
	testUserB        = "B12345"
	testADUser       = "456789"
	testNewADAccount = "AD123"
	testPassword     = "123456"
	testBadUser      = "223768"
	testBadPassword  = "721228"

	testLeaveDate = "2099-01-02"

	testDepartmentID          = "T1000"
	testNewDepartmentID       = "N1234"
	testDepartmentA           = "T1100"
	testDepartmentB           = "T1200"
	testDepartmentC           = "T1230"
	testDepartmentIDNotExists = "T5678"
)

func Test_Users(t *testing.T) {
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
	assert.NoError(deleteAllDepartmentsData(db))

	{ // UpdateUser: insufficient request: ID, DepartmentID.
		err := dm.UpdateUser(ctx, mcom.UpdateUserRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the user id is empty or one of the department id or leave date or account is empty",
		}, err)
	}
	{ // UpdateUser: new department id not found.
		err := dm.UpdateUser(ctx, mcom.UpdateUserRequest{
			ID:           testUser,
			DepartmentID: testNewDepartmentID,
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_DEPARTMENT_NOT_FOUND})
	}
	{ // CreateDepartment: insufficient request: DepartmentID.
		err := dm.CreateDepartments(ctx, nil)
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "there is no department to create",
		}, err)
	}
	{ // CreateDepartment: good case, add new department.
		err := dm.CreateDepartments(ctx, []string{testNewDepartmentID})
		assert.NoError(err)
	}
	{ // CreateDepartment: department already exists.
		err := dm.CreateDepartments(ctx, []string{testNewDepartmentID})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_DEPARTMENT_ALREADY_EXISTS,
		}, err)
	}
	{ // UpdateUser: user not found.
		err := dm.UpdateUser(ctx, mcom.UpdateUserRequest{
			ID:           testUser,
			DepartmentID: testNewDepartmentID,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_USER_NOT_FOUND,
		}, err)
	}
	{ // CreateUser: insufficient users.
		err := dm.CreateUsers(ctx, mcom.CreateUsersRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "there is no user to create",
		}, err)
	}
	{ // CreateUser: insufficient request: ID, DepartmentID.
		err := dm.CreateUsers(ctx, mcom.CreateUsersRequest{
			Users: []mcom.User{{
				ID:           "",
				DepartmentID: "",
			}},
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the user id or the department id is empty",
		}, err)
	}
	{ // CreateUser: department not found.
		err := dm.CreateUsers(ctx, mcom.CreateUsersRequest{
			Users: []mcom.User{
				{
					ID:           testUser,
					DepartmentID: testDepartmentID,
				},
			},
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_DEPARTMENT_NOT_FOUND})
	}
	{ // CreateDepartment: good case, add department.
		err := dm.CreateDepartments(ctx, []string{testDepartmentID})
		assert.NoError(err)
	}
	{ // CreateUser: good case.
		err = dm.CreateUsers(ctx, mcom.CreateUsersRequest{
			Users: []mcom.User{
				{
					ID:           testUser,
					DepartmentID: testDepartmentID,
				},
			},
		})
		assert.NoError(err)

		// check user info.
		dep := &models.Department{}
		err = db.Where(` id = ? `, testDepartmentID).Take(dep).Error
		assert.NoError(err)

		userInfo := &models.User{}
		err = db.Where(` id = ? and department_id = ? `, testUser, dep.ID).Take(userInfo).Error
		assert.NoError(err)
		assert.Equal(&models.User{
			ID:           testUser,
			DepartmentID: dep.ID,
			LeaveDate: sql.NullTime{
				Time:  time.Time{},
				Valid: false,
			},
		}, userInfo)
	}
	{ // CreateUser: create AD user, good case.
		err = dm.CreateUsers(ctx, mcom.CreateUsersRequest{
			Users: []mcom.User{
				{
					ID:           testADUser,
					Account:      testADUser,
					DepartmentID: testDepartmentID,
				},
			},
		})
		assert.NoError(err)

		// check user info.
		dep := &models.Department{}
		err = db.Where(` id = ? `, testDepartmentID).Take(dep).Error
		assert.NoError(err)

		userInfo := &models.User{}
		err = db.Where(` id = ? and department_id = ? `, testADUser, dep.ID).Take(userInfo).Error
		assert.NoError(err)
		assert.Equal(&models.User{
			ID:           testADUser,
			Account:      testADUser,
			DepartmentID: dep.ID,
			LeaveDate: sql.NullTime{
				Time:  time.Time{},
				Valid: false,
			},
		}, userInfo)
	}
	{ // CreateUser: user already exists.
		err = dm.CreateUsers(ctx, mcom.CreateUsersRequest{
			Users: []mcom.User{
				{
					ID:           testUser,
					DepartmentID: testDepartmentID,
				},
			},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_USER_ALREADY_EXISTS,
		}, err)
	}
	{ // UpdateUser: good case, update department.
		err = dm.UpdateUser(ctx, mcom.UpdateUserRequest{
			ID:           testUser,
			DepartmentID: testNewDepartmentID,
		})
		assert.NoError(err)

		// check result.
		dep := &models.Department{}
		err = db.Where(` id = ? `, testNewDepartmentID).Take(dep).Error
		assert.NoError(err)

		userInfo := &models.User{}
		err = db.Where(` id = ? and department_id = ? `, testUser, dep.ID).Take(userInfo).Error
		assert.NoError(err)
		assert.Equal(&models.User{
			ID:           testUser,
			DepartmentID: dep.ID,
		}, userInfo)
	}
	{ // UpdateUser: good case, update leave date.
		leaveDate, err := time.Parse(dateLayout, testLeaveDate)
		assert.NoError(err)
		leaveDate = leaveDate.Local()

		err = dm.UpdateUser(ctx, mcom.UpdateUserRequest{
			ID:        testUser,
			LeaveDate: leaveDate,
		})
		assert.NoError(err)

		// check result.
		dep := &models.Department{}
		err = db.Where(` id = ? `, testNewDepartmentID).Take(dep).Error
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
				Time:  leaveDate.Local(),
				Valid: true,
			},
		}, userInfo)
	}
	{ // UpdateUser: good case, update AD account.
		err = dm.UpdateUser(ctx, mcom.UpdateUserRequest{
			ID:      testADUser,
			Account: testNewADAccount,
		})
		assert.NoError(err)

		// check result.
		dep := &models.Department{}
		err = db.Where(` id = ? `, testDepartmentID).Take(dep).Error
		assert.NoError(err)

		userInfo := &models.User{}
		err = db.Where(` id = ? and department_id = ? `, testADUser, dep.ID).Take(userInfo).Error
		assert.NoError(err)

		// set to the same timezone for the comparison.
		userInfo.LeaveDate.Time = userInfo.LeaveDate.Time.Local()
		assert.Equal(&models.User{
			ID:           testADUser,
			Account:      testNewADAccount,
			DepartmentID: dep.ID,
			LeaveDate: sql.NullTime{
				Time:  time.Time{}.Local(),
				Valid: false,
			},
		}, userInfo)
	}
	{ // CreateUser: good case,user reinstatement.
		err = dm.CreateUsers(ctx, mcom.CreateUsersRequest{
			Users: []mcom.User{
				{
					ID:           testUser,
					DepartmentID: testDepartmentID,
				},
			},
		})
		assert.NoError(err)

		// check user info.
		dep := &models.Department{}
		err = db.Where(` id = ? `, testDepartmentID).Take(dep).Error
		assert.NoError(err)

		userInfo := &models.User{}
		err = db.Where(` id = ? `, testUser).Take(userInfo).Error
		assert.NoError(err)
		assert.Equal(&models.User{
			ID:           testUser,
			DepartmentID: dep.ID,
			LeaveDate: sql.NullTime{
				Time:  time.Time{},
				Valid: false,
			},
		}, userInfo)
	}
	{ // DeleteUser: insufficient request: ID.
		err := dm.DeleteUser(ctx, mcom.DeleteUserRequest{ID: ""})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // DeleteUser: good case.
		err := dm.DeleteUser(ctx, mcom.DeleteUserRequest{ID: testUser})
		assert.NoError(err)
		err = dm.DeleteUser(ctx, mcom.DeleteUserRequest{ID: testADUser})
		assert.NoError(err)
	}
	{ // UpdateDepartment: insufficient request: OldID, NewID.
		err := dm.UpdateDepartment(ctx, mcom.UpdateDepartmentRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // UpdateDepartment: new department already exists.
		err := dm.UpdateDepartment(ctx, mcom.UpdateDepartmentRequest{
			OldID: testDepartmentID,
			NewID: testNewDepartmentID,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_DEPARTMENT_ALREADY_EXISTS,
		}, err)
	}
	{ // DeleteDepartment: insufficient request: ID.
		err := dm.DeleteDepartment(ctx, mcom.DeleteDepartmentRequest{DepartmentID: ""})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // DeleteDepartment: department not found.
		err := dm.DeleteDepartment(ctx, mcom.DeleteDepartmentRequest{DepartmentID: testDepartmentIDNotExists})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_DEPARTMENT_NOT_FOUND,
		}, err)
	}
	{ // DeleteDepartment: good case, deprecate new department.
		err := dm.DeleteDepartment(ctx, mcom.DeleteDepartmentRequest{DepartmentID: testNewDepartmentID})
		assert.NoError(err)
	}
	{ // UpdateDepartment: department not found.
		err := dm.UpdateDepartment(ctx, mcom.UpdateDepartmentRequest{
			OldID: testDepartmentIDNotExists,
			NewID: testNewDepartmentID,
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_DEPARTMENT_NOT_FOUND,
		}, err)
	}
	{ // UpdateDepartment: good case.
		err := dm.UpdateDepartment(ctx, mcom.UpdateDepartmentRequest{
			OldID: testDepartmentID,
			NewID: testNewDepartmentID,
		})
		assert.NoError(err)
	}
}

func TestDataManager_ListAllDepartments(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, models.Department{})
	assert.NoError(cm.Clear())
	{ // good case.
		assert.NoError(dm.CreateDepartments(ctx, mcom.CreateDepartmentsRequest{"12345", "aiueo", "abcde"}))
		rep, err := dm.ListAllDepartment(ctx)
		assert.NoError(err)
		expected := mcom.ListAllDepartmentReply{IDs: []string{"12345", "abcde", "aiueo"}}
		assert.Equal(expected, rep)
	}
}

func deleteAllUser(db *gorm.DB) error {
	return newClearMaster(db, &models.User{}).Clear()
}

func Test_ListUnauthorizedUsers(t *testing.T) {
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
	assert.NoError(deleteAllDepartmentsData(db))
	assert.NoError(deleteAllUser(db))

	{ // insufficient request, departmentOID.
		_, err = dm.ListUnauthorizedUsers(ctx, mcom.ListUnauthorizedUsersRequest{DepartmentOID: ""})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // department oid not found.
		_, err = dm.ListUnauthorizedUsers(ctx, mcom.ListUnauthorizedUsersRequest{DepartmentOID: testDepartmentID})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_DEPARTMENT_NOT_FOUND,
		}, err)
	}
	{ // good case.
		assert.NoError(dm.CreateDepartments(ctx, mcom.CreateDepartmentsRequest{testDepartmentID, testDepartmentA}))
		assert.NoError(dm.CreateUsers(ctx, mcom.CreateUsersRequest{
			Users: []mcom.User{
				{
					ID:           testUser,
					DepartmentID: testDepartmentID,
				},
				{
					ID:           testUserA,
					DepartmentID: testDepartmentID,
				},
				{
					ID:           testUserB,
					DepartmentID: testDepartmentID,
				},
				{
					ID:           testBadUser,
					DepartmentID: testDepartmentA,
				},
			},
		}))

		assert.NoError(dm.CreateAccounts(ctx, mcom.CreateAccountsRequest{
			mcom.CreateAccountRequest{
				ID:    testUserA,
				Roles: []roles.Role{roles.Role_BEARER},
			}.WithSpecifiedPassword("123456")}))

		users, err := dm.ListUnauthorizedUsers(ctx, mcom.ListUnauthorizedUsersRequest{DepartmentOID: testDepartmentID})
		assert.NoError(err)
		assert.Equal(mcom.ListUnauthorizedUsersReply{testUser, testUserB}, users)

		users, err = dm.ListUnauthorizedUsers(ctx, mcom.ListUnauthorizedUsersRequest{DepartmentOID: testDepartmentID}, mcom.ExcludeUsers([]string{testUserB}))
		assert.NoError(err)
		assert.Equal(mcom.ListUnauthorizedUsersReply{testUser}, users)
	}

	assert.NoError(deleteAllAccount(db))
	assert.NoError(deleteAllDepartmentsData(db))
	assert.NoError(deleteAllUser(db))
}

func deleteSignInOutData(db *gorm.DB) error {
	return newClearMaster(db, &models.User{}, &models.Station{}, &models.SiteContents{}, &models.Site{}, &models.Department{}, &models.Account{}).Clear()
}

func Test_SignInStation(t *testing.T) {
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
	assert.NoError(deleteSignInOutData(db))

	{ // insufficient request.
		err := dm.SignInStation(ctx, mcom.SignInStationRequest{})
		assert.Error(err)
		assert.Equal(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing station",
		}, err)
	}
	{ // bad work date.
		err := dm.SignInStation(ctx, mcom.SignInStationRequest{
			Station:  testStationA,
			Site:     models.SiteID{Name: testStationA},
			WorkDate: time.Date(2000, time.April, 1, 1, 1, 1, 1, time.Local),
		}, mcom.WithVerifyWorkDateHandler(func(t time.Time) bool { return false }))
		assert.ErrorIs(err, mcomErr.Error{
			Code: mcomErr.Code_BAD_WORK_DATE,
		})
	}
	{ // station not found.
		// add test user.
		assert.NoError(
			testDataBuilder.NewTestUserBuilder().
				WithAccount(testUser).
				WithDepartmentID(testDepartmentID).
				WithUserName(testUser).
				Build(dm),
		)

		err = dm.SignInStation(ctx, mcom.SignInStationRequest{
			Station:  testStationA,
			Site:     models.SiteID{Name: testStationA},
			WorkDate: time.Now(),
		})
		assert.Error(err)
		assert.Equal(mcomErr.Error{
			Code:    mcomErr.Code_STATION_NOT_FOUND,
			Details: fmt.Sprintf("station not found, id: %v", testStationA),
		}, err)
	}
	{ // station site not found.
		// add test station.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            testStationA,
			DepartmentOID: testDepartmentID,
			Sites: []mcom.SiteInformation{{
				Name:    testStationA,
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_OPERATOR,
			}},
			State: stations.State_IDLE,
		}))

		err = dm.SignInStation(ctx, mcom.SignInStationRequest{
			Station:  testStationA,
			Site:     models.SiteID{Name: "not found"},
			WorkDate: time.Now(),
		})
		assert.Error(err)
		assert.Equal(mcomErr.Error{
			Code: mcomErr.Code_STATION_SITE_NOT_FOUND,
		}, err)
	}
	{ // normal.
		workdate := time.Now()
		assert.NoError(dm.SignInStation(ctx, mcom.SignInStationRequest{
			Station:  testStationA,
			Site:     models.SiteID{Name: testStationA},
			WorkDate: workdate,
		}))

		actual, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: testStationA,
			SiteName:  testStationA,
			SiteIndex: 0,
		})
		assert.NoError(err)

		// 由於time套件的內部實作關係導致無法直接以assert比較。
		actualWorkDate := actual.Content.Slot.Operator.WorkDate

		actual.Content.Slot.Operator.WorkDate = workdate

		expected := mcom.GetSiteReply{
			Name:               testStationA,
			Index:              0,
			AdminDepartmentOID: testDepartmentID,
			Attributes: mcom.SiteAttributes{
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_OPERATOR,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Operator: &models.OperatorSite{
						EmployeeID: testUser,
						Group:      0,
						WorkDate:   workdate,
					},
				},
			},
		}

		assert.True(actualWorkDate.Equal(workdate))
		actual.Attributes.LimitHandler = nil // function assertion will always fail.
		assert.Equal(expected, actual)
	}
	{ // user not sign out (same user) (should pass)
		err = dm.SignInStation(ctx, mcom.SignInStationRequest{
			Station:  testStationA,
			Site:     models.SiteID{Name: testStationA},
			Group:    123,
			WorkDate: time.Now(),
		})
		assert.NoError(err)

		actual, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: testStationA,
			SiteName:  testStationA,
			SiteIndex: 0,
		})
		assert.NoError(err)

		// WorkDate has been asserted in normal case.
		expected := mcom.GetSiteReply{
			Name:               testStationA,
			Index:              0,
			AdminDepartmentOID: testDepartmentID,
			Attributes: mcom.SiteAttributes{
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_OPERATOR,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Operator: &models.OperatorSite{
						EmployeeID: testUser,
						Group:      123,
						WorkDate:   actual.Content.Slot.Operator.WorkDate,
					},
				},
			},
		}

		actual.Attributes.LimitHandler = nil // function assertion will always fail.
		assert.Equal(expected, actual)
	}
	{ // user not sign out (should not pass)
		anotherUserCtx := commonsCtx.WithUserID(context.Background(), "another_user")
		err = dm.SignInStation(anotherUserCtx, mcom.SignInStationRequest{
			Station:  testStationA,
			Site:     models.SiteID{Name: testStationA},
			WorkDate: time.Now(),
		})
		assert.ErrorIs(mcomErr.Error{Code: mcomErr.Code_PREVIOUS_USER_NOT_SIGNED_OUT}, err)
	}
	{ // wrong sub type (should not pass)
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            "WrongSubType",
			DepartmentOID: "d",
			Sites: []mcom.SiteInformation{{
				Name:    "WrongSubType",
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_MATERIAL,
			}},
			State: stations.State_IDLE,
		}))

		err = dm.SignInStation(ctx, mcom.SignInStationRequest{
			Station:  "WrongSubType",
			Site:     models.SiteID{Name: "WrongSubType"},
			WorkDate: time.Now(),
		})
		assert.ErrorIs(mcomErr.Error{Code: mcomErr.Code_STATION_SITE_SUB_TYPE_MISMATCH}, err)
	}
	{ // build a new site before sign in.
		workDate := time.Now().UTC()
		assert.NoError(dm.SignInStation(ctx, mcom.SignInStationRequest{
			Station:  testStationA,
			Site:     models.SiteID{Name: "not exist"},
			Group:    2,
			WorkDate: workDate,
		}, mcom.CreateSiteIfNotExists()))

		actual, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: testStationA,
			SiteName:  "not exist",
			SiteIndex: 0,
		})
		assert.NoError(err)

		// WorkDate has been asserted in normal case.
		expected := mcom.GetSiteReply{
			Name:               "not exist",
			Index:              0,
			AdminDepartmentOID: testDepartmentID,
			Attributes: mcom.SiteAttributes{
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_OPERATOR,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Operator: &models.OperatorSite{
						EmployeeID: testUser,
						Group:      2,
						WorkDate:   workDate,
					},
				},
			},
		}

		actual.Attributes.LimitHandler = nil // function assertion will always fail.
		assert.Equal(expected, actual)
	}
	{ // user not sign out (force)
		anotherUserCtx := commonsCtx.WithUserID(context.Background(), "another_user")
		assert.NoError(dm.SignInStation(anotherUserCtx, mcom.SignInStationRequest{
			Station:  testStationA,
			Site:     models.SiteID{Name: testStationA},
			Group:    2,
			WorkDate: time.Now(),
		}, mcom.ForceSignIn()))

		actual, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: testStationA,
			SiteName:  testStationA,
			SiteIndex: 0,
		})
		assert.NoError(err)

		// WorkDate has been asserted in normal case.
		expected := mcom.GetSiteReply{
			Name:               testStationA,
			Index:              0,
			AdminDepartmentOID: testDepartmentID,
			Attributes: mcom.SiteAttributes{
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_OPERATOR,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Operator: &models.OperatorSite{
						EmployeeID: "another_user",
						Group:      2,
						WorkDate:   actual.Content.Slot.Operator.WorkDate,
					},
				},
			},
		}

		actual.Attributes.LimitHandler = nil // function assertion will always fail.
		assert.Equal(expected, actual)
	}
	{ // build a not exist empty-name site
		now := time.Now().UTC()
		assert.NoError(dm.SignInStation(ctx, mcom.SignInStationRequest{
			Station:        "S01",
			Site:           models.SiteID{},
			Group:          2,
			WorkDate:       now,
			ExpiryDuration: time.Hour,
		}, mcom.CreateSiteIfNotExists(), mcom.ForceSignIn()))

		rep, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "S01",
			SiteName:  "",
			SiteIndex: 0,
		})
		assert.NoError(err)

		rep.Attributes.LimitHandler = nil
		assert.Equal(mcom.GetSiteReply{
			Name:               "",
			Index:              0,
			AdminDepartmentOID: "T1000",
			Attributes: mcom.SiteAttributes{
				Type:         sites.Type_SLOT,
				SubType:      sites.SubType_OPERATOR,
				LimitHandler: nil,
			},
			Content: models.SiteContent{
				Slot: &models.Slot{
					Operator: &models.OperatorSite{
						EmployeeID: "123456",
						Group:      2,
						WorkDate:   now,
					},
				},
			},
		}, rep)
	}
	assert.NoError(deleteSignInOutData(db))
}

func Test_SignOutStation(t *testing.T) {
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
	assert.NoError(deleteSignInOutData(db))

	{ // insufficient request.
		err := dm.SignOutStation(ctx, mcom.SignOutStationRequest{})
		assert.Error(err)
		assert.Equal(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing station",
		}, err)
	}
	{ // station not found.
		// add test user.
		assert.NoError(
			testDataBuilder.NewTestUserBuilder().
				WithAccount(testUser).
				WithDepartmentID(testDepartmentID).
				WithUserName(testUser).
				DoNotBuildAccount().
				Build(dm),
		)

		err = dm.SignOutStation(ctx, mcom.SignOutStationRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: testStationA},
		})
		assert.Error(err)
		assert.Equal(mcomErr.Error{
			Code:    mcomErr.Code_STATION_NOT_FOUND,
			Details: fmt.Sprintf("station not found, id: %v", testStationA),
		}, err)
	}
	{ // station site not found.
		// add test station.
		assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
			ID:            testStationA,
			DepartmentOID: testDepartmentID,
			Sites: []mcom.SiteInformation{{
				Name:    testStationA,
				Index:   0,
				Type:    sites.Type_SLOT,
				SubType: sites.SubType_OPERATOR,
			}},
			State: stations.State_IDLE,
		}))

		err = dm.SignOutStation(ctx, mcom.SignOutStationRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: "not found"},
		})
		assert.Error(err)
		assert.Equal(mcomErr.Error{
			Code: mcomErr.Code_STATION_SITE_NOT_FOUND,
		}, err)
	}
	{ // user has not signed in (should pass)
		assert.NoError(dm.SignOutStation(ctx, mcom.SignOutStationRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: testStationA},
		}))
	}
	{ // user not match.
		ctx = commonsCtx.WithUserID(context.Background(), testBadUser)
		assert.NoError(dm.SignInStation(ctx, mcom.SignInStationRequest{
			Station:  testStationA,
			Site:     models.SiteID{Name: testStationA},
			WorkDate: time.Now(),
		}))

		ctx = commonsCtx.WithUserID(context.Background(), testUser)
		err = dm.SignOutStation(ctx, mcom.SignOutStationRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: testStationA},
		})
		assert.Error(err)
		assert.Equal(mcomErr.Error{
			Code:    mcomErr.Code_STATION_OPERATOR_NOT_MATCH,
			Details: "user logged in: 223768, user to log out: 123456",
		}, err)

		ctx = commonsCtx.WithUserID(context.Background(), testBadUser)
		assert.NoError(dm.SignOutStation(ctx, mcom.SignOutStationRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: testStationA},
		}))
		ctx = commonsCtx.WithUserID(context.Background(), testUser)
	}
	{ // normal.
		assert.NoError(dm.SignInStation(ctx, mcom.SignInStationRequest{
			Station: testStationA,
			Site: models.SiteID{
				Name:  testStationA,
				Index: 0,
			},
			WorkDate:       time.Now(),
			ExpiryDuration: 0,
		}))

		err = dm.SignOutStation(ctx, mcom.SignOutStationRequest{
			Station: testStationA,
			Site:    models.SiteID{Name: testStationA},
		})
		assert.NoError(err)
	}

	assert.NoError(deleteSignInOutData(db))
}

func TestDataManager_SignOutStations(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, models.Station{}, models.Site{}, models.SiteContents{})
	assert.NoError(cm.Clear())
	// #region initialize
	const (
		userID1       = "user1"
		userID2       = "user2"
		stationID1    = "station1"
		stationID2    = "station2"
		siteName      = "siteName"
		departmentOID = "dep"
	)
	now := time.Now().UTC()
	userCtx1 := commonsCtx.WithUserID(ctx, userID1)
	userCtx2 := commonsCtx.WithUserID(ctx, userID2)
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            stationID1,
		DepartmentOID: departmentOID,
		Sites: []mcom.SiteInformation{
			{
				Name:       siteName,
				Index:      0,
				Type:       sites.Type_SLOT,
				SubType:    sites.SubType_OPERATOR,
				Limitation: []string{},
			},
		},
		State: stations.State_IDLE,
	}))
	assert.NoError(dm.CreateStation(ctx, mcom.CreateStationRequest{
		ID:            stationID2,
		DepartmentOID: departmentOID,
		Sites: []mcom.SiteInformation{
			{
				Name:       siteName,
				Index:      0,
				Type:       sites.Type_SLOT,
				SubType:    sites.SubType_OPERATOR,
				Limitation: []string{},
			},
		},
		State: stations.State_IDLE,
	}))
	// #endregion initialize
	{ // signing out from multiple stations.
		assert.NoError(dm.SignInStation(userCtx1, mcom.SignInStationRequest{
			Station: stationID1,
			Site: models.SiteID{
				Name:  siteName,
				Index: 0,
			},
			Group:          1,
			WorkDate:       now,
			ExpiryDuration: time.Hour,
		}))
		assert.NoError(dm.SignInStation(userCtx1, mcom.SignInStationRequest{
			Station: stationID2,
			Site: models.SiteID{
				Name:  siteName,
				Index: 0,
			},
			Group:          1,
			WorkDate:       now,
			ExpiryDuration: time.Hour,
		}))
		assert.NoError(dm.SignOutStations(userCtx1, mcom.SignOutStationsRequest{
			Sites: []models.UniqueSite{
				{
					SiteID: models.SiteID{
						Name:  siteName,
						Index: 0,
					},
					Station: stationID1,
				},
				{
					SiteID: models.SiteID{
						Name:  siteName,
						Index: 0,
					},
					Station: stationID2,
				},
			},
		}))

		site1, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: stationID1,
			SiteName:  siteName,
			SiteIndex: 0,
		})
		assert.NoError(err)

		site2, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: stationID2,
			SiteName:  siteName,
			SiteIndex: 0,
		})
		assert.NoError(err)

		expected := models.SiteContent{
			Slot: &models.Slot{
				Operator: new(models.OperatorSite),
			},
		}
		assert.Equal(expected, site1.Content)
		assert.Equal(expected, site2.Content)
	}
	{ // do NOT sign out from an incorrect station.
		assert.NoError(dm.SignInStation(userCtx1, mcom.SignInStationRequest{
			Station: stationID1,
			Site: models.SiteID{
				Name:  siteName,
				Index: 0,
			},
			Group:          1,
			WorkDate:       now,
			ExpiryDuration: time.Hour,
		}))
		assert.NoError(dm.SignInStation(userCtx2, mcom.SignInStationRequest{
			Station: stationID2,
			Site: models.SiteID{
				Name:  siteName,
				Index: 0,
			},
			Group:          1,
			WorkDate:       now,
			ExpiryDuration: time.Hour,
		}))

		assert.NoError(dm.SignOutStations(userCtx1, mcom.SignOutStationsRequest{
			Sites: []models.UniqueSite{
				{
					SiteID: models.SiteID{
						Name:  siteName,
						Index: 0,
					},
					Station: stationID1,
				},
				{
					SiteID: models.SiteID{
						Name:  siteName,
						Index: 0,
					},
					Station: stationID2,
				},
			},
		}))

		site1, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: stationID1,
			SiteName:  siteName,
			SiteIndex: 0,
		})
		assert.NoError(err)

		site2, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: stationID2,
			SiteName:  siteName,
			SiteIndex: 0,
		})
		assert.NoError(err)

		expectedEmpty := models.SiteContent{
			Slot: &models.Slot{
				Operator: new(models.OperatorSite),
			},
		}

		expectedNotEmpty := models.SiteContent{
			Slot: &models.Slot{
				Operator: &models.OperatorSite{
					EmployeeID: userID2,
					Group:      1,
					WorkDate:   now,
				},
			},
		}
		assert.Equal(expectedEmpty, site1.Content)
		assert.Equal(expectedNotEmpty, site2.Content)
	}
}
