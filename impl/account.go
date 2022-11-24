package impl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"

	commonsAccount "gitlab.kenda.com.tw/kenda/commons/v2/util/account"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
)

// SignIn implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) SignIn(ctx context.Context, req mcom.SignInRequest, opts ...mcom.SignInOption) (mcom.SignInReply, error) {
	opt := mcom.ParseSignInOptions(opts)

	session := dm.newSession(ctx)
	roles, mustChangePassword, err := session.checkAccount(req.Account, req.Password, req.ADUser)
	if err != nil {
		return mcom.SignInReply{}, err
	}

	departments, err := session.getEmployeesDepartments(req.Account, req.ADUser)
	if err != nil {
		return mcom.SignInReply{}, err
	}

	expiryTime := time.Now().Add(opt.ExpiredAfter).In(time.Local)
	token, err := session.createToken(req.Account, expiryTime, models.UserInfo{
		Roles: roles,
	})
	if err != nil {
		return mcom.SignInReply{}, err
	}

	return mcom.SignInReply{
		Token:              token,
		TokenExpiry:        expiryTime,
		Departments:        departments,
		Roles:              roles,
		MustChangePassword: mustChangePassword,
	}, nil
}

// SignOut implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) SignOut(ctx context.Context, req mcom.SignOutRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	// update token status.
	result := session.db.Model(&models.Token{
		ID: models.EncryptedData(req.Token),
	}).Select("Valid").Updates(&models.Token{
		Valid: false,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return mcomErr.Error{
			Code: mcomErr.Code_USER_UNKNOWN_TOKEN,
		}
	}
	return nil
}

// checkAccount returns account roles and whether the account should change the password or not.
func (session *session) checkAccount(id, pwd string, ad bool) ([]roles.Role, bool, error) {
	if id == "" || pwd == "" {
		return nil, false, mcomErr.Error{
			Code:    mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
			Details: "missing id or password",
		}
	}
	var mustChangePassword bool
	userRoles := pq.Int64Array{}
	if ad {
		if session.agent == nil {
			return nil, false, fmt.Errorf("unsupported AD login")
		}
		if err := session.agent.Auth(id, pwd); err != nil {
			return nil, false, mcomErr.Error{
				Code: mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
			}
		}

		results, err := session.agent.SearchUserRoles(id)
		if err != nil && !errors.Is(err, commonsAccount.ErrUserRolesNotExist) {
			return nil, false, err
		}

		for _, r := range results {
			v, ok := roles.Role_value[r]
			if !ok {
				continue
			}
			userRoles = append(userRoles, int64(v))
		}
	} else {
		var account models.Account
		if err := session.db.Where(models.Account{ID: id}).
			Take(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, mcomErr.Error{
					Code:    mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
					Details: "account not found",
				}
			}
			return nil, false, err
		}

		if !bytes.Equal(account.Password, models.Encrypt([]byte(pwd))) {
			return nil, false, mcomErr.Error{
				Code:    mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
				Details: "wrong password",
			}
		}
		userRoles = account.Roles
		mustChangePassword = account.MustChangePassword
	}

	rs := make([]roles.Role, len(userRoles))
	for i, role := range userRoles {
		rs[i] = roles.Role(role)
	}
	return rs, mustChangePassword, nil
}

type changeUserPasswordRequest struct {
	UserID      string
	OldPassword string
	NewPassword string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req changeUserPasswordRequest) CheckInsufficiency() error {
	if req.UserID == "" || req.NewPassword == "" || req.OldPassword == "" {
		return mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}
	}
	return nil
}

func (session *session) changeUserPassword(req changeUserPasswordRequest) error {
	userID := req.UserID
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	// Assuming that wrong id is impossible.
	updatedContent := make(map[string]interface{})
	updatedContent["password"] = models.EncryptedData(req.NewPassword)
	updatedContent["must_change_password"] = false
	result := session.db.
		Model(&models.Account{}).
		Where(`id = ? AND password = ?`, userID, models.EncryptedData(req.OldPassword)).
		Updates(updatedContent)
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return mcomErr.Error{
			Code: mcomErr.Code_ACCOUNT_BAD_OLD_PASSWORD,
		}
	}

	if req.NewPassword == req.OldPassword {
		return mcomErr.Error{Code: mcomErr.Code_ACCOUNT_SAME_AS_OLD_PASSWORD}
	}

	return nil
}

func convertInt64ArrayToRoles(rs pq.Int64Array) []roles.Role {
	results := make([]roles.Role, len(rs))
	for i, v := range rs {
		results[i] = roles.Role(v)
	}
	return results
}

func checkCreateAccountRequestAndReturnModel(req mcom.CreateAccountsRequest) ([]models.Account, error) {
	results := make([]models.Account, len(req))

	for index, elem := range req {
		if err := elem.Roles.CheckPermission(); err != nil {
			return []models.Account{}, err
		}
		results[index] = models.Account{
			ID:                 elem.ID,
			Password:           models.EncryptedData(elem.GetPassword()),
			Roles:              convertRolesToInt64Array(elem.Roles),
			MustChangePassword: true,
		}
	}

	return results, nil
}

// CreateAccounts implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateAccounts(ctx context.Context, req mcom.CreateAccountsRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	accounts, err := checkCreateAccountRequestAndReturnModel(req)
	if err != nil {
		return err
	}

	session := dm.newSession(ctx)

	if err := session.db.Create(&accounts).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return mcomErr.Error{Code: mcomErr.Code_ACCOUNT_ALREADY_EXISTS}
		}
		return err
	}
	return nil
}

// UpdateAccount implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) UpdateAccount(ctx context.Context, req mcom.UpdateAccountRequest, opts ...mcom.UpdateAccountOption) error {
	option := mcom.ParseUpdateAccountOptions(opts)
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	anyAction := false
	if option.ResetPassword {
		if err := session.resetUserPassword(req.UserID); err != nil {
			return err
		}
		anyAction = true
	} else if req.IsChangingPassword() {
		if err := session.changeUserPassword(changeUserPasswordRequest{
			UserID:      req.UserID,
			OldPassword: req.ChangePassword.OldPassword,
			NewPassword: req.ChangePassword.NewPassword,
		}); err != nil {
			return err
		}
		anyAction = true
	}
	if req.IsModifyingRoles() {
		if err := session.db.
			Where(&models.Account{ID: req.UserID}).
			Updates(&models.Account{Roles: convertRolesToInt64Array(req.Roles)}).Error; err != nil {
			return err
		}
		anyAction = true
	}

	if !anyAction {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "the request did not match any conditions of UpdateAccount",
		}
	}
	return nil
}

func convertRolesToInt64Array(roleSlice []roles.Role) pq.Int64Array {
	res := make([]int64, len(roleSlice))
	for i, role := range roleSlice {
		res[i] = int64(role)
	}
	return res
}

// resetUserPassword.
func (session *session) resetUserPassword(id string) error {
	updatedContent := make(map[string]interface{})
	updatedContent["password"] = models.EncryptedData(id)
	updatedContent["must_change_password"] = true
	result := session.db.
		Model(&models.Account{}).
		Where(`id=?`, id).
		Updates(updatedContent)
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return mcomErr.Error{
			Code: mcomErr.Code_ACCOUNT_NOT_FOUND,
		}
	}
	return nil
}

// DeleteAccount implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) DeleteAccount(ctx context.Context, req mcom.DeleteAccountRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	result := session.db.
		Where(`id=?`, req.ID).
		Delete(&models.Account{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return mcomErr.Error{
			Code: mcomErr.Code_ACCOUNT_NOT_FOUND,
		}
	}
	return nil
}
