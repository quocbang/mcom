package impl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commons_context "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// GetTokenInfo implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetTokenInfo(ctx context.Context, req mcom.GetTokenInfoRequest) (mcom.GetTokenInfoReply, error) {
	if req.Token == "" {
		return mcom.GetTokenInfoReply{}, mcomErr.Error{
			Code:    mcomErr.Code_USER_UNKNOWN_TOKEN,
			Details: "missing token",
		}
	}

	session := dm.newSession(ctx)
	// get token info.
	tokenInfo, err := session.GetToken(string(req.Token))
	if err != nil {
		return mcom.GetTokenInfoReply{}, err
	}

	// token expired.
	if tokenInfo.ExpiryTime.Time().Before(time.Now()) {
		zap.L().Warn("token expired",
			zap.String("base16-encoded encrypted token", fmt.Sprintf("%x", tokenInfo.ID)),
		)
	}

	return mcom.GetTokenInfoReply{
		User:        tokenInfo.BoundUser,
		Valid:       tokenInfo.Valid,
		ExpiryTime:  tokenInfo.ExpiryTime.Time(),
		CreatedTime: tokenInfo.CreatedTime.Time(),
		Roles:       tokenInfo.Info.Roles,
	}, nil
}

func (session *session) getEmployeesDepartments(userID string, ad bool) ([]mcom.Department, error) {
	var user models.User
	var condition map[string]interface{}
	if ad {
		condition = map[string]interface{}{"account": userID}
	} else {
		condition = map[string]interface{}{"id": userID}
	}
	if err := session.db.Where(condition).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, mcomErr.Error{
				Code: mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
			}
		}
		return nil, fmt.Errorf("failed to get the department of the user: %v", err)
	}

	if user.Resigned() {
		return nil, mcomErr.Error{
			Code: mcomErr.Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD,
		}
	}

	var dep models.Department
	if err := session.db.Where("id = ?", user.DepartmentID).
		Take(&dep).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, mcomErr.Error{Code: mcomErr.Code_DEPARTMENT_NOT_FOUND}
		}
		return nil, err
	}

	var deps []models.Department
	if err := session.db.Where("id LIKE ?", strings.TrimRight(dep.ID, "0")+"%").
		Order("id").
		Find(&deps).Error; err != nil {
		return nil, err
	}

	departments := make([]mcom.Department, len(deps))
	for i, dep := range deps {
		departments[i] = mcom.Department{
			OID: dep.ID,
			ID:  dep.ID,
		}
	}
	return departments, nil
}

func (session *session) createToken(user string, expiryTime time.Time, info models.UserInfo) (string, error) {
	token := uuid.NewV4().String()
	// insert token.
	if err := session.db.Create(&models.Token{
		ID:          models.EncryptedData(token),
		BoundUser:   user,
		ExpiryTime:  types.ToTimeNano(expiryTime),
		CreatedTime: types.ToTimeNano(time.Now()),
		Valid:       true,
		Info:        info,
	}).Error; err != nil {
		return "", err
	}

	return token, nil
}

// GetToken gets token info.
func (session *session) GetToken(id string) (*models.Token, error) {
	var token models.Token
	if err := session.db.Where(&models.Token{
		ID: models.EncryptedData(id),
	}).Take(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, mcomErr.Error{
				Code: mcomErr.Code_USER_UNKNOWN_TOKEN,
			}
		}
		return nil, err
	}
	return &token, nil
}

func (tx *txDataManager) createUser(id, account, departmentID string) error {
	user := &models.User{}
	if err := tx.db.
		Where(` id = ? AND department_id = ? `, id, departmentID).
		Find(user).Error; err != nil {
		return err
	}

	if user.ID != "" && !user.Resigned() {
		return mcomErr.Error{
			Code: mcomErr.Code_USER_ALREADY_EXISTS,
		}
	}

	if err := tx.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"account", "leave_date", "department_id"}),
	}).Create(&models.User{
		ID:           id,
		Account:      account,
		DepartmentID: departmentID,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (session *session) updateUser(id, account, departmentID string, leaveDate time.Time) error {
	updates := &models.User{}
	if account != "" {
		updates.Account = account
	}
	if departmentID != "" {
		updates.DepartmentID = departmentID
	}
	if !leaveDate.IsZero() {
		updates.LeaveDate = sql.NullTime{
			Time:  leaveDate,
			Valid: true,
		}
	}

	result := session.db.Model(&models.User{
		ID: id,
	}).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return mcomErr.Error{
			Code: mcomErr.Code_USER_NOT_FOUND,
		}
	}
	return nil
}

func (session *session) deleteUser(id string) error {
	if err := session.db.Delete(&models.User{
		ID: id,
	}).Error; err != nil {
		return err
	}
	return nil
}

// CreateUsers implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateUsers(ctx context.Context, req mcom.CreateUsersRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	tx := session.beginTx()
	defer tx.Rollback() // nolint: errcheck

	var allDepartments []models.Department
	if err := session.db.Find(&allDepartments).Error; err != nil {
		return err
	}

	for _, user := range req.Users {
		departmentFound := false
		for _, dep := range allDepartments {
			if dep.ID == user.DepartmentID {
				departmentFound = true
			}
		}
		if !departmentFound {
			return mcomErr.Error{Code: mcomErr.Code_DEPARTMENT_NOT_FOUND}
		}

		if err := tx.createUser(user.ID, user.Account, user.DepartmentID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// UpdateUser implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) UpdateUser(ctx context.Context, req mcom.UpdateUserRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	if req.DepartmentID != "" {
		if session.db.Where(&models.Department{ID: req.DepartmentID}).Find(&models.Department{}).RowsAffected == 0 {
			return mcomErr.Error{Code: mcomErr.Code_DEPARTMENT_NOT_FOUND}
		}
	}

	return session.updateUser(req.ID, req.Account, req.DepartmentID, req.LeaveDate)
}

// DeleteUser implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) DeleteUser(ctx context.Context, req mcom.DeleteUserRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	return session.deleteUser(req.ID)
}

// getDepartmentInfo gets the department, it returns Code_DEPARTMENT_NOT_FOUND
// error code if the department is not found.
func (session *session) getDepartmentInfo(id string) (*models.Department, error) {
	var department models.Department
	if err := session.db.Where(` id = ? `, id).
		Take(&department).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, mcomErr.Error{
				Code: mcomErr.Code_DEPARTMENT_NOT_FOUND,
			}
		}
		return nil, err
	}
	return &department, nil
}

// CreateDepartments implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateDepartments(ctx context.Context, ids mcom.CreateDepartmentsRequest) error {
	if len(ids) == 0 {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "there is no department to create",
		}
	}

	session := dm.newSession(ctx)
	tx := session.beginTx()
	defer tx.Rollback() // nolint: errcheck

	departments := make([]models.Department, len(ids))
	for i, id := range ids {
		departments[i] = models.Department{ID: id}
	}
	if err := tx.db.Create(&departments).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return mcomErr.Error{Code: mcomErr.Code_DEPARTMENT_ALREADY_EXISTS}
		}
		return err
	}

	return tx.Commit()
}

func (session *session) deprecateDepartment(id string) error {
	result := session.db.Where(` id = ? `, id).Delete(models.Department{})
	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return mcomErr.Error{
			Code: mcomErr.Code_DEPARTMENT_NOT_FOUND,
		}
	}
	return nil
}

// DeleteDepartment implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) DeleteDepartment(ctx context.Context, req mcom.DeleteDepartmentRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	return session.deprecateDepartment(string(req.DepartmentID))
}

// UpdateDepartment implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) UpdateDepartment(ctx context.Context, req mcom.UpdateDepartmentRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	result := session.db.
		Model(&models.Department{
			ID: req.OldID,
		}).
		Updates(&models.Department{
			ID: req.NewID,
		})
	if err := result.Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return mcomErr.Error{Code: mcomErr.Code_DEPARTMENT_ALREADY_EXISTS}
		}
		return err
	}
	if result.RowsAffected == 0 {
		return mcomErr.Error{
			Code: mcomErr.Code_DEPARTMENT_NOT_FOUND,
		}
	}
	return nil
}

func (dm *DataManager) ListAllDepartment(ctx context.Context) (mcom.ListAllDepartmentReply, error) {
	session := dm.newSession(ctx)
	var departments []models.Department
	if err := session.db.Order("id").Find(&departments).Error; err != nil {
		return mcom.ListAllDepartmentReply{}, err
	}

	ids := make([]string, len(departments))
	for i := range departments {
		ids[i] = departments[i].ID
	}

	return mcom.ListAllDepartmentReply{
		IDs: ids,
	}, nil
}

// ListUserRoles implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListUserRoles(ctx context.Context, req mcom.ListUserRolesRequest) (mcom.ListUserRolesReply, error) {
	if req.DepartmentOID == "" {
		return mcom.ListUserRolesReply{}, mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "empty departmentID",
		}
	}

	session := dm.newSession(ctx)
	_, err := session.getDepartmentInfo(req.DepartmentOID)
	if err != nil {
		return mcom.ListUserRolesReply{}, err
	}

	accounts := []models.Account{}
	if err := session.db.
		Where(` id IN (?) `,
			session.db.Select(` id `).
				Where(` department_id = ? `, req.DepartmentOID).
				Find(&[]models.User{}),
		).Order("id").Find(&accounts).Error; err != nil {
		return mcom.ListUserRolesReply{}, err
	}

	userRoles := make([]mcom.UserRoles, len(accounts))
	for i, v := range accounts {
		userRoles[i] = mcom.UserRoles{
			ID:    v.ID,
			Roles: convertInt64ArrayToRoles(v.Roles),
		}
	}

	return mcom.ListUserRolesReply{
		Users: userRoles,
	}, nil
}

// ListRoles implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListRoles(context.Context) (mcom.ListRolesReply, error) {
	r := make([]mcom.Role, len(roles.Role_name))
	for i := 0; i < len(roles.Role_name); i++ {
		roleName, ok := roles.Role_name[int32(i)]
		if !ok {
			continue
		}

		r[i] = mcom.Role{
			Name:  roleName,
			Value: roles.Role(i),
		}
	}
	return mcom.ListRolesReply{
		Roles: r,
	}, nil
}

// ListUnauthorizedUsers implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListUnauthorizedUsers(ctx context.Context, req mcom.ListUnauthorizedUsersRequest, opts ...mcom.ListUnauthorizedUsersOption) (mcom.ListUnauthorizedUsersReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListUnauthorizedUsersReply{}, err
	}

	opt := mcom.ParseListUnauthorizedUsersOptions(opts)
	isExcludedID := func(id string) bool {
		for _, excludeID := range opt.ExcludeUsers {
			if id == excludeID {
				return true
			}
		}
		return false
	}
	session := dm.newSession(ctx)
	_, err := session.getDepartmentInfo(string(req.DepartmentOID))
	if err != nil {
		return nil, err
	}

	var users Users
	var accounts Accounts

	eg := errgroup.Group{}
	eg.Go(func() error {
		return session.db.Where(&models.User{DepartmentID: string(req.DepartmentOID)}).Find(&users).Error
	})
	eg.Go(func() error { return session.db.Find(&accounts).Error })
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// users in specified department.
	resultUsers := users.Where(func(u models.User) bool { return !isExcludedID(u.ID) })
	// remove users with MES account.
	resultUsers.RemoveAll(func(u models.User) bool {
		return accounts.Any(func(au models.Account) bool { return au.ID == u.ID })
	})

	res := make([]string, len(resultUsers))
	for i, user := range resultUsers {
		res[i] = user.ID
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})

	return res, nil
}

// SignInStation implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) SignInStation(ctx context.Context, req mcom.SignInStationRequest, opts ...mcom.SignInStationOption) error {
	userID := commons_context.UserID(ctx)
	if userID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "missing user id"}
	}
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	o := mcom.ParseSignInStationOptions(opts)

	if !o.VerifyWorkDate(req.WorkDate) {
		return mcomErr.Error{
			Code: mcomErr.Code_BAD_WORK_DATE,
		}
	}

	session := dm.newSession(ctx)

	// check station exist.
	station, err := session.getStation(req.Station)
	if err != nil {
		return err
	}
	if !station.Sites.Contains(req.Site) {
		if !o.CreateSiteIfNotExists {
			return mcomErr.Error{
				Code: mcomErr.Code_STATION_SITE_NOT_FOUND,
			}
		}
		if err := dm.UpdateStation(ctx, mcom.UpdateStationRequest{
			ID: req.Station,
			Sites: []mcom.UpdateStationSite{
				{
					ActionMode: sites.ActionType_ADD,
					Information: mcom.SiteInformation{
						Name:    req.Site.Name,
						Index:   int(req.Site.Index),
						Type:    sites.Type_SLOT,
						SubType: sites.SubType_OPERATOR,
					},
				},
			},
		}); err != nil {
			return err
		}
	}

	siteContents, err := session.getSiteContent(models.UniqueSite{
		SiteID:  req.Site,
		Station: req.Station,
	})
	if err != nil {
		return err
	}

	attr, err := session.getSiteAttributes(models.UniqueSite{
		SiteID:  req.Site,
		Station: req.Station,
	})
	if err != nil {
		return err
	}
	if attr.SubType != sites.SubType_OPERATOR {
		return mcomErr.Error{Code: mcomErr.Code_STATION_SITE_SUB_TYPE_MISMATCH}
	}

	if !o.Force &&
		siteContents.Slot != nil &&
		siteContents.Slot.Operator != nil &&
		!siteContents.Slot.Operator.AllowedLogin(userID) {
		return mcomErr.Error{
			Code: mcomErr.Code_PREVIOUS_USER_NOT_SIGNED_OUT,
		}
	}

	return session.db.
		Model(&models.SiteContents{}).
		Where(" name = ? AND index = ? AND station = ?", req.Site.Name, req.Site.Index, req.Station).
		Select("content", "updated_by").
		Updates(&models.SiteContents{
			Content: models.SiteContent{
				Slot: &models.Slot{
					Operator: &models.OperatorSite{
						EmployeeID: userID,
						Group:      int8(req.Group),
						WorkDate:   req.WorkDate,
					},
				},
			},
			UpdatedBy: userID,
		}).Error
}

func (dm *DataManager) SignOutStations(ctx context.Context, req mcom.SignOutStationsRequest) error {
	userID := commons_context.UserID(ctx)
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}
	if userID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "missing user id"}
	}

	tx, cancel := dm.beginTx(ctx).withTimeout()
	defer cancel()
	defer tx.Rollback() // nolint: errcheck
	if err := tx.signOutStations(userID, req); err != nil {
		return err
	}
	return tx.Commit()
}

func (tx *txDataManager) signOutStations(userID string, req mcom.SignOutStationsRequest) error {
	condition := make([][3]string, len(req.Sites))
	for i, site := range req.Sites {
		condition[i] = [3]string{
			site.Station,
			site.SiteID.Name,
			strconv.Itoa(int(site.SiteID.Index)),
		}
	}

	var siteContents []models.SiteContents
	if err := tx.db.Model(&models.SiteContents{}).
		Clauses(clause.Locking{
			Strength: "UPDATE",
		}).Where(`(station,name,index) IN ?`, condition).Find(&siteContents).Error; err != nil {
		return err
	}

	condition = nil

	for _, sc := range siteContents {
		if sc.Content.Slot.Operator.Current().EmployeeID == userID {
			condition = append(condition, [3]string{sc.Station, sc.Name, strconv.Itoa(int(sc.Index))})
		}
	}

	if len(condition) > 0 {
		return tx.db.Model(&models.SiteContents{}).
			Where(`(station,name,index) IN ?`, condition).
			Updates(map[string]interface{}{
				"content": models.SiteContent{
					Slot: &models.Slot{
						Operator: new(models.OperatorSite),
					},
				},
				"updated_by": userID,
			}).Error
	}
	return nil
}

// SignOutStation implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) SignOutStation(ctx context.Context, req mcom.SignOutStationRequest) error {
	userID := commons_context.UserID(ctx)
	if userID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "missing user id"}
	}
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)

	// check station exist.
	station, err := session.getStation(req.Station)
	if err != nil {
		return err
	}
	if !station.Sites.Contains(req.Site) {
		return mcomErr.Error{
			Code: mcomErr.Code_STATION_SITE_NOT_FOUND,
		}
	}

	siteContents, err := session.getSiteContent(models.UniqueSite{
		SiteID:  req.Site,
		Station: req.Station,
	})
	if err != nil {
		return err
	}

	attr, err := session.getSiteAttributes(models.UniqueSite{
		SiteID:  req.Site,
		Station: req.Station,
	})
	if err != nil {
		return err
	}
	if attr.SubType != sites.SubType_OPERATOR {
		return mcomErr.Error{Code: mcomErr.Code_STATION_SITE_SUB_TYPE_MISMATCH}
	}

	current := siteContents.Slot.Operator.Current()
	if current.EmployeeID == "" {
		logger := commons_context.Logger(ctx)
		logger.Info("user has not signed in", zap.String("user ID", userID))
		return nil
	}

	if current.EmployeeID != userID {
		return mcomErr.Error{
			Code:    mcomErr.Code_STATION_OPERATOR_NOT_MATCH,
			Details: "user logged in: " + siteContents.Slot.Operator.EmployeeID + ", user to log out: " + userID,
		}
	}

	siteContents.Slot.Clear()
	return session.db.
		Model(&models.SiteContents{}).
		Where(" name = ? AND index = ? AND station = ?", req.Site.Name, req.Site.Index, req.Station).
		Select("content", "updated_by").
		Updates(&models.SiteContents{
			Content:   siteContents,
			UpdatedBy: userID,
		}).Error
}
