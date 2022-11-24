package impl

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
)

// ListStationState implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListStationState(context.Context) (mcom.ListStationStateReply, error) {
	types := make([]mcom.StationState, len(stations.State_value)-1)
	for i := 1; ; i++ {
		name, ok := stations.State_name[int32(i)]
		if !ok {
			break
		}
		types[i-1] = mcom.StationState{
			Name:  name,
			Value: stations.State(i),
		}
	}
	return types, nil
}

func parseSlot(subType sites.SubType, resource *models.Slot) *models.Slot {
	if resource == nil {
		return nil
	}
	switch subType {
	case sites.SubType_OPERATOR:
		return &models.Slot{
			Operator: resource.Operator,
		}
	case sites.SubType_MATERIAL:
		return &models.Slot{
			Material: resource.Material,
		}
	case sites.SubType_TOOL:
		return &models.Slot{
			Tool: resource.Tool,
		}
	}
	return nil
}

func parseContainer(subType sites.SubType, resources *models.Container) *models.Container {
	results := make(models.Container, len(*resources))
	for i, v := range *resources {
		switch subType {
		case sites.SubType_MATERIAL:
			results[i] = models.BoundResource{
				Material: v.Material,
			}
		}
	}
	return &results
}

func parseCollection(subType sites.SubType, resources *models.Collection) *models.Collection {
	results := make(models.Collection, len(*resources))
	for i, v := range *resources {
		switch subType {
		case sites.SubType_MATERIAL:
			results[i] = models.BoundResource{
				Material: v.Material,
			}
		}
	}
	return &results
}

func parseQueue(subType sites.SubType, resources *models.Queue) *models.Queue {
	results := make(models.Queue, len(*resources))
	for i, v := range *resources {
		switch subType {
		case sites.SubType_MATERIAL:
			results[i] = models.Slot{
				Material: v.Material,
			}
		}
	}
	return &results
}

func parseColqueue(subType sites.SubType, resources *models.Colqueue) *models.Colqueue {
	results := make(models.Colqueue, len(*resources))
	for i, r := range *resources {
		res := make(models.Collection, len(r))
		for j, v := range r {
			switch subType {
			case sites.SubType_MATERIAL:
				res[j] = models.BoundResource{
					Material: v.Material,
				}
			}
		}
		results[i] = r
	}
	return &results
}

func parseContent(siteType sites.Type, subType sites.SubType, content models.SiteContents) models.SiteContent {
	switch siteType {
	case sites.Type_SLOT:
		return models.SiteContent{Slot: parseSlot(subType, content.Content.Slot)}
	case sites.Type_CONTAINER:
		if content.Content.Container != nil {
			return models.SiteContent{Container: parseContainer(subType, content.Content.Container)}
		}
	case sites.Type_COLLECTION:
		if content.Content.Collection != nil {
			return models.SiteContent{Collection: parseCollection(subType, content.Content.Collection)}
		}
	case sites.Type_QUEUE:
		if content.Content.Queue != nil {
			return models.SiteContent{Queue: parseQueue(subType, content.Content.Queue)}
		}
	case sites.Type_COLQUEUE:
		if content.Content.Colqueue != nil {
			return models.SiteContent{Colqueue: parseColqueue(subType, content.Content.Colqueue)}
		}
	}
	return models.SiteContent{}
}

func (tx *txDataManager) listStationSites(sites []models.UniqueSite, opts ...parseStationSitesOption) ([]mcom.ListStationSite, error) {
	opt := parseParseStationSitesOptions(opts)
	siteCounts := len(sites)
	result := make([]mcom.ListStationSite, siteCounts)
	conditionSite := make([][3]string, siteCounts)
	for i, site := range sites {
		conditionSite[i] = [3]string{site.Station, site.SiteID.Name, strconv.Itoa(int(site.SiteID.Index))}
	}

	var targetSites []models.Site
	if err := tx.db.Where(`(station, name, index) IN ?`, conditionSite).Find(&targetSites).Error; err != nil {
		return nil, err
	}
	if len(targetSites) != siteCounts {
		return nil, fmt.Errorf("site not found")
	}

	conditionSiteContents := make([][3]string, siteCounts)
	for i, cs := range conditionSite {
		conditionSiteContents[i] = [3]string{cs[0], cs[1], cs[2]}
	}
	var targetSitesContents []models.SiteContents
	contentsQuery := tx.db.Where(`(station, name, index) IN ?`, conditionSiteContents)
	if opt.selectForUpdate {
		if err := contentsQuery.
			Clauses(clause.Locking{
				Strength: "UPDATE",
				Options:  "NOWAIT",
			}).Find(&targetSitesContents).Error; err != nil {
			return nil, err
		}
	} else {
		if err := contentsQuery.Find(&targetSitesContents).Error; err != nil {
			return nil, err
		}
	}
	if len(targetSitesContents) != siteCounts {
		return nil, fmt.Errorf("siteContents not found")
	}

	// #region map siteContents
	mapContents := make(map[models.SiteID]models.SiteContents)
	for _, contents := range targetSitesContents {
		mapContents[models.SiteID{
			Name:  contents.Name,
			Index: contents.Index,
		}] = contents
	}
	// #endregion map siteContents

	for i := 0; i < siteCounts; i++ {
		result[i] = mcom.ListStationSite{
			Information: mcom.ListStationSitesInformation{
				UniqueSite: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  targetSites[i].Name,
						Index: targetSites[i].Index,
					},
					Station: targetSites[i].Station,
				},
				Type:    targetSites[i].Attributes.Type,
				SubType: targetSites[i].Attributes.SubType,
			},
			Content: parseContent(targetSites[i].Attributes.Type, targetSites[i].Attributes.SubType, mapContents[models.SiteID{Name: targetSites[i].Name, Index: targetSites[i].Index}]),
		}
	}

	// sort.
	sort.Slice(result, func(i, j int) bool {
		if result[i].Information.SiteID.Name == result[j].Information.SiteID.Name {
			return result[i].Information.SiteID.Index < result[j].Information.SiteID.Index
		}
		return result[i].Information.SiteID.Name < result[j].Information.SiteID.Name
	})

	return result, nil
}

type parseStationSitesOptions struct {
	selectForUpdate bool
}

type parseStationSitesOption func(*parseStationSitesOptions)

func selectForUpdate() parseStationSitesOption {
	return func(gso *parseStationSitesOptions) {
		gso.selectForUpdate = true
	}
}

// parse parse.
func parseParseStationSitesOptions(opts []parseStationSitesOption) parseStationSitesOptions {
	var opt parseStationSitesOptions
	for _, elem := range opts {
		elem(&opt)
	}
	return opt
}

func (session *session) listStations(stations []models.Station) ([]mcom.Station, error) {
	resultSlice := make([]mcom.Station, len(stations))
	tx := session.beginTx()
	defer tx.Rollback() // nolint: errcheck
	for i, v := range stations {
		sites, err := tx.listStationSites(v.Sites)
		if err != nil {
			return nil, err
		}

		resultSlice[i] = mcom.Station{
			ID:                 v.ID,
			AdminDepartmentOID: v.AdminDepartmentID,
			Sites:              sites,
			State:              v.State,
			Information: mcom.StationInformation{
				Code:        v.Information.Code,
				Description: v.Information.Description,
			},
			UpdatedBy:  v.UpdatedBy,
			UpdatedAt:  v.UpdatedAt.Time(),
			InsertedBy: v.CreatedBy,
			InsertedAt: v.CreatedAt.Time(),
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return resultSlice, nil
}

func (session *session) GetStation(station models.Station) (mcom.GetStationReply, error) {
	rep, err := session.listStations([]models.Station{station})
	if err != nil {
		return mcom.GetStationReply{}, err
	}
	return mcom.GetStationReply(rep[0]), nil
}

func (session *session) getStation(id string) (station models.Station, err error) {
	if err = session.db.Where(models.Station{ID: id}).Take(&station).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = mcomErr.Error{
				Code:    mcomErr.Code_STATION_NOT_FOUND,
				Details: fmt.Sprintf("station not found, id: %s", id),
			}
			return
		}
		return
	}
	return
}

// GetStation implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetStation(ctx context.Context, req mcom.GetStationRequest) (mcom.GetStationReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetStationReply{}, err
	}

	session := dm.newSession(ctx)
	resModel, err := session.getStation(req.ID)
	if err != nil {
		return mcom.GetStationReply{}, err
	}
	return session.GetStation(resModel)
}

// ListStations implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListStations(ctx context.Context, req mcom.ListStationsRequest) (mcom.ListStationsReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListStationsReply{}, err
	}

	session := dm.newSession(ctx)
	dataCounts, stations, err := listHandler[models.Station](&session, req, func(d *gorm.DB) *gorm.DB {
		return d.Model(&models.Station{}).Where(models.Station{AdminDepartmentID: req.DepartmentOID})
	})
	if err != nil {
		return mcom.ListStationsReply{}, err
	}

	s, err := session.listStations(stations)
	if err != nil {
		return mcom.ListStationsReply{}, err
	}
	return mcom.ListStationsReply{
		Stations: s,
		PaginationReply: mcom.PaginationReply{
			AmountOfData: dataCounts,
		},
	}, nil
}

func (tx *txDataManager) getSiteInfo(name string, index int, station string) (*models.Site, error) {
	site := &models.Site{}
	if err := tx.db.
		Where(`name = ? AND index = ? AND station = ?`, name, index, station).
		Take(site).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return site, nil
}

func (tx *txDataManager) CheckIfSitesExist(sites []willCreateSite) ([]models.UniqueSite, error) {
	condition := make([][3]string, len(sites))
	res := make([]models.UniqueSite, len(sites))
	for i, s := range sites {
		condition[i] = [3]string{s.Name, strconv.Itoa(s.Index), s.Station}
		res[i] = models.UniqueSite{
			SiteID: models.SiteID{
				Name:  s.Name,
				Index: int16(s.Index),
			},
			Station: s.Station,
		}
	}

	var count int64
	err := tx.db.Model(&models.Site{}).Where(`(name, index, station) IN ?`, condition).Count(&count).Error
	if err != nil {
		return []models.UniqueSite{}, err
	}
	if count != int64(len(condition)) {
		return []models.UniqueSite{}, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND}
	}
	return res, nil
}

func (tx *txDataManager) createSites(createdBy, departmentOID, station string, updatedSites []willCreateSite) ([]models.UniqueSite, error) {
	if len(updatedSites) == 0 {
		return []models.UniqueSite{}, nil
	}

	stationSites := make([]models.UniqueSite, len(updatedSites))
	createdSites := make([]models.Site, len(updatedSites))
	createdSiteContents := make([]models.SiteContents, len(updatedSites))
	for i, v := range updatedSites {
		site, err := tx.getSiteInfo(v.Name, v.Index, station)
		if err != nil {
			return nil, err
		}

		if site != nil {
			return nil, mcomErr.Error{
				Code:    mcomErr.Code_STATION_SITE_ALREADY_EXISTS,
				Details: fmt.Sprintf("name: %v, index: %d", v.Name, v.Index),
			}
		} else {
			if v.SubType == sites.SubType_SUB_TYPE_UNSPECIFIED || v.Type == sites.Type_TYPE_UNSPECIFIED {
				return nil, mcomErr.Error{
					Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
					Details: fmt.Sprintf("missing type or subtype while adding sites, name: %v, index: %d", v.Name, v.Index),
				}
			}
			if v.Limitation == nil {
				v.Limitation = []string{}
			}
			createdSites[i] = models.Site{
				Name:              v.Name,
				Index:             int16(v.Index),
				AdminDepartmentID: departmentOID,
				Attributes: models.SiteAttributes{
					Type:       v.Type,
					SubType:    v.SubType,
					Limitation: v.Limitation,
				},
				Station:   station,
				UpdatedBy: createdBy,
				CreatedBy: createdBy,
			}
			var content models.SiteContent
			switch v.Type {
			case sites.Type_COLLECTION:
				content = models.NewCollectionSiteContent()
			case sites.Type_SLOT:
				content = models.NewSlotSiteContent()
			case sites.Type_COLQUEUE:
				content = models.NewColqueueSiteContent()
			case sites.Type_CONTAINER:
				content = models.NewContainerSiteContent()
			case sites.Type_QUEUE:
				content = models.NewQueueSiteContent()
			}
			createdSiteContents[i] = models.SiteContents{
				Name:      v.Name,
				Index:     int16(v.Index),
				Station:   station,
				Content:   content,
				UpdatedBy: createdBy,
			}
		}
		stationSites[i] = models.UniqueSite{
			SiteID: models.SiteID{
				Name:  v.Name,
				Index: int16(v.Index),
			},
			Station: station,
		}
	}
	if err := tx.db.Create(createdSites).Error; err != nil {
		return nil, err
	}

	if err := tx.db.Create(createdSiteContents).Error; err != nil {
		return nil, err
	}
	return stationSites, nil
}

// CreateStation implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateStation(ctx context.Context, req mcom.CreateStationRequest) error {
	req.Correct()
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	tx := dm.beginTx(ctx)
	defer tx.Rollback() // nolint: errcheck

	toCreateSites, toAssociateSites := splitOwnSitesAndForeignSites(toWillCreateSites(req.Sites), req.ID)

	createdSites, err := tx.createSites(commonsCtx.UserID(ctx), req.DepartmentOID, req.ID, toCreateSites)
	if err != nil {
		return err
	}

	associatedSites, err := tx.CheckIfSitesExist(toAssociateSites)
	if err != nil {
		return err
	}

	state := req.State
	if req.State == stations.State_UNSPECIFIED {
		state = stations.State_SHUTDOWN
	}

	if err := tx.db.
		Create(&models.Station{
			ID:                req.ID,
			AdminDepartmentID: req.DepartmentOID,
			Sites:             append(createdSites, associatedSites...),
			State:             state,
			Information: models.StationInformation{
				Code:        req.Information.Code,
				Description: req.Information.Description,
			},
			UpdatedBy: commonsCtx.UserID(ctx),
			CreatedBy: commonsCtx.UserID(ctx),
		}).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return mcomErr.Error{
				Code: mcomErr.Code_STATION_ALREADY_EXISTS,
			}
		}
	}
	return tx.Commit()
}

type splitSiteType interface {
	IsOwnedBy(station string) bool
}

type willCreateSite mcom.SiteInformation

func toWillCreateSites(sites []mcom.SiteInformation) []willCreateSite {
	res := make([]willCreateSite, len(sites))
	for i := range sites {
		res[i] = willCreateSite(sites[i])
	}
	return res
}

func (wcs willCreateSite) IsOwnedBy(station string) bool {
	return wcs.Station == station
}

type willDeleteSite models.UniqueSite

func (wds willDeleteSite) IsOwnedBy(station string) bool {
	return wds.Station == station
}

func splitOwnSitesAndForeignSites[T splitSiteType](sites []T, station string) (ownSites, foreignSites []T) {
	for _, site := range sites {
		if site.IsOwnedBy(station) {
			ownSites = append(ownSites, site)
		} else {
			foreignSites = append(foreignSites, site)
		}
	}
	return
}

type stationComposer struct {
	station models.Station

	currentSites    map[models.UniqueSite]struct{}
	willCreateSites []willCreateSite
	willDeleteSites []willDeleteSite
}

func newStationComposer(station models.Station) *stationComposer {
	m := make(map[models.UniqueSite]struct{}, len(station.Sites))
	for _, site := range station.Sites {
		m[site] = struct{}{}
	}
	return &stationComposer{
		station: station,

		currentSites:    m,
		willCreateSites: make([]willCreateSite, 0),
		willDeleteSites: make([]willDeleteSite, 0),
	}
}

func (composer *stationComposer) maybeSetDepartment(oid string) bool {
	if oid == "" {
		return false
	}
	composer.station.AdminDepartmentID = oid
	return true
}

func (composer *stationComposer) maybeSetState(state stations.State) bool {
	if state == stations.State_UNSPECIFIED {
		return false
	}
	composer.station.State = state
	return true
}

func (composer *stationComposer) maybeSetInformation(info mcom.StationInformation) bool {
	var emptyInfo mcom.StationInformation
	if info == emptyInfo {
		return false
	}
	composer.station.Information = models.StationInformation{
		Code:        info.Code,
		Description: info.Description,
	}
	return true
}

func (composer *stationComposer) setUpdater(id string) {
	composer.station.UpdatedBy = id
}

func (composer *stationComposer) addSite(site mcom.SiteInformation) error {
	stationSite := models.UniqueSite{
		SiteID: models.SiteID{
			Name:  site.Name,
			Index: int16(site.Index),
		},
		Station: site.Station,
	}

	if _, ok := composer.currentSites[stationSite]; ok {
		return mcomErr.Error{
			Code:    mcomErr.Code_STATION_SITE_ALREADY_EXISTS,
			Details: fmt.Sprintf("site name=%s, site index=%d", site.Name, site.Index),
		}
	}
	composer.currentSites[stationSite] = struct{}{}
	composer.willCreateSites = append(composer.willCreateSites, willCreateSite(site))
	return nil
}

func (composer *stationComposer) deleteSite(site models.UniqueSite) error {
	if _, ok := composer.currentSites[site]; !ok {
		return mcomErr.Error{
			Code:    mcomErr.Code_STATION_SITE_NOT_FOUND,
			Details: fmt.Sprintf("site name=%s, site index=%d", site.SiteID.Name, site.SiteID.Index),
		}
	}
	delete(composer.currentSites, site)
	composer.willDeleteSites = append(composer.willDeleteSites, willDeleteSite(site))
	return nil
}

func (composer *stationComposer) composeStation() models.Station {
	composer.station.Sites = composer.listCurrentSites()
	return composer.station
}

// listCurrentSites return current sites ordered by name, index.
func (composer *stationComposer) listCurrentSites() []models.UniqueSite {
	sites, i := make([]models.UniqueSite, len(composer.currentSites)), 0
	for site := range composer.currentSites {
		sites[i] = site
		i++
	}

	sort.Slice(sites, func(i, j int) bool {
		if sites[i].SiteID.Name == sites[j].SiteID.Name {
			return sites[i].SiteID.Index < sites[j].SiteID.Index
		}
		return sites[i].SiteID.Name < sites[j].SiteID.Name
	})
	return sites
}

func (composer *stationComposer) listWillCreateSites() []willCreateSite {
	return composer.willCreateSites
}

func (composer *stationComposer) listWillDeleteSites() []willDeleteSite {
	return composer.willDeleteSites
}

func (tx *txDataManager) updateStation(id string, data models.Station) error {
	result := tx.db.
		Model(&models.Station{ID: id}).
		Updates(models.Station{
			AdminDepartmentID: data.AdminDepartmentID,
			Sites:             data.Sites,
			State:             data.State,
			Information:       data.Information,
			UpdatedAt:         data.UpdatedAt,
			UpdatedBy:         data.UpdatedBy,
			CreatedAt:         data.CreatedAt,
			CreatedBy:         data.CreatedBy,
		})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return mcomErr.Error{
			Code: mcomErr.Code_STATION_NOT_FOUND,
		}
	}
	return nil
}

// UpdateStation implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) UpdateStation(ctx context.Context, req mcom.UpdateStationRequest) error {
	req.Correct()
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	station, err := session.getStation(req.ID)
	if err != nil {
		return err
	}

	composer := newStationComposer(station)

	// add or delete sites.
	for i, s := range req.Sites {
		var err error
		switch s.ActionMode {
		case sites.ActionType_ADD:
			err = composer.addSite(s.Information)
		case sites.ActionType_REMOVE:
			err = composer.deleteSite(models.UniqueSite{
				SiteID: models.SiteID{
					Name:  s.Information.Name,
					Index: int16(s.Information.Index),
				},
				Station: s.Information.Station,
			})
		default:
			err = fmt.Errorf("unsupported site action mode [%d] at %d-index of the Sites field", s.ActionMode, i)
		}

		if err != nil {
			return err
		}
	}

	composer.maybeSetDepartment(req.DepartmentOID)
	composer.maybeSetState(req.State)
	composer.maybeSetInformation(req.Information)
	composer.setUpdater(commonsCtx.UserID(ctx))

	tx := dm.beginTx(ctx)
	defer tx.Rollback() // nolint: errcheck

	toCreateSites, toAssociateSites := splitOwnSitesAndForeignSites(composer.listWillCreateSites(), req.ID)

	if _, err := tx.createSites(commonsCtx.UserID(ctx), station.AdminDepartmentID, req.ID, toCreateSites); err != nil {
		return err
	}

	if _, err := tx.CheckIfSitesExist(toAssociateSites); err != nil {
		return err
	}

	toDeleteSites, _ := splitOwnSitesAndForeignSites(composer.listWillDeleteSites(), req.ID)

	updatedBy := commonsCtx.UserID(ctx)

	if err := session.isStationSitesEmpty(toDeleteSites); err != nil {
		return err
	}
	if err := tx.maybeDeleteSites(toDeleteSites, updatedBy); err != nil {
		return err
	}

	station = composer.composeStation()
	if err := tx.updateStation(req.ID, station); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteStation implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) DeleteStation(ctx context.Context, req mcom.DeleteStationRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	targetStation, err := dm.GetStation(ctx, mcom.GetStationRequest{
		ID: req.StationID,
	})
	if err != nil {
		return err
	}
	session := dm.newSession(ctx)
	uniqueSites := []willDeleteSite{}
	for _, site := range targetStation.Sites {
		uniqueSites = append(uniqueSites, willDeleteSite{
			SiteID:  site.Information.SiteID,
			Station: site.Information.Station,
		})
	}

	toDelete, _ := splitOwnSitesAndForeignSites(uniqueSites, req.StationID)

	if err := session.isStationSitesEmpty(toDelete); err != nil {
		return err
	}

	tx := dm.beginTx(ctx)
	defer tx.Rollback() // nolint: errcheck
	updatedBy := commonsCtx.UserID(ctx)

	if err := tx.maybeDeleteSites(toDelete, updatedBy); err != nil {
		return err
	}

	result := tx.db.Where(models.Station{ID: req.StationID}).Delete(&models.Station{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return mcomErr.Error{
			Code: mcomErr.Code_STATION_NOT_FOUND,
		}
	}

	return tx.Commit()
}

// maybeDeleteSites deletes sites.
func (tx *txDataManager) maybeDeleteSites(sites []willDeleteSite, updatedBy string) error {
	if len(sites) == 0 {
		return nil
	}

	toDelete := make([][3]string, len(sites))
	for i, site := range sites {
		toDelete[i] = [3]string{site.Station, site.SiteID.Name, strconv.Itoa(int(site.SiteID.Index))}
		associatedStations, err := tx.listAssociatedStations(mcom.ListAssociatedStationsRequest{
			Site: models.UniqueSite(site),
		})
		if err != nil {
			return err
		}
		if len(associatedStations.StationIDs) > 1 {
			return fmt.Errorf("please dissociate the stations first")
		}
	}

	if err := tx.db.Where("(station, name, index) IN ?", toDelete).Delete(&models.SiteContents{}).Error; err != nil {
		return err
	}
	if err := tx.db.Where(`(station_id, site_name, site_index) IN ?`, toDelete).Delete(&models.BindRecords{}).Error; err != nil {
		return err
	}
	return tx.db.Where("(station, name, index) IN ?", toDelete).Delete(&models.Site{}).Error
}

func newRemainingObjectsError(site willDeleteSite) error {
	return mcomErr.Error{
		Code: mcomErr.Code_STATION_SITE_REMAINING_OBJECTS,
		Details: fmt.Sprintf("remaining objects in [station: %s, site name: %s, site index: %d]",
			site.Station,
			site.SiteID.Name,
			site.SiteID.Index,
		),
	}
}

// return nil if all the sites are empty.
func (session *session) isStationSitesEmpty(sites []willDeleteSite) error {
	conditions := make([][3]string, len(sites))
	for i, site := range sites {
		conditions[i] = [3]string{site.Station, site.SiteID.Name, strconv.Itoa(int(site.SiteID.Index))}
	}
	var siteContents []models.SiteContents
	if err := session.db.Where(`(station, name, index) IN ?`, conditions).Find(&siteContents).Error; err != nil {
		return err
	}
	if len(siteContents) != len(sites) {
		return fmt.Errorf("site contents not found")
	}

	for i, sc := range siteContents {
		if sc.Content.Slot != nil &&
			(sc.Content.Slot.Material != nil ||
				sc.Content.Slot.Operator != nil ||
				sc.Content.Slot.Tool != nil) {
			return newRemainingObjectsError(sites[i])
		}
		if sc.Content.Container != nil && len(*sc.Content.Container) > 0 {
			return newRemainingObjectsError(sites[i])
		}
		if sc.Content.Collection != nil && len(*sc.Content.Collection) > 0 {
			return newRemainingObjectsError(sites[i])
		}
		if sc.Content.Queue != nil && len(*sc.Content.Queue) > 0 {
			return newRemainingObjectsError(sites[i])
		}
		if sc.Content.Colqueue != nil && len(*sc.Content.Colqueue) > 0 {
			return newRemainingObjectsError(sites[i])
		}
	}
	return nil
}

func (session *session) createStationGroup(id string, stations []string) error {
	if err := session.db.Create(&models.StationGroup{
		ID:       id,
		Stations: pq.StringArray(stations),
	}).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return mcomErr.Error{
				Code: mcomErr.Code_STATION_GROUP_ALREADY_EXISTS,
			}
		}
		return err
	}
	return nil
}

func (session *session) updateStationGroup(id string, stations []string) error {
	result := session.db.Model(&models.StationGroup{
		ID: id,
	}).Updates(&models.StationGroup{
		Stations: pq.StringArray(stations),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return mcomErr.Error{
			Code: mcomErr.Code_STATION_GROUP_ID_NOT_FOUND,
		}
	}
	return nil
}

// CreateStationGroup implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateStationGroup(ctx context.Context, req mcom.StationGroupRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	return session.createStationGroup(req.ID, req.Stations)
}

// UpdateStationGroup implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) UpdateStationGroup(ctx context.Context, req mcom.StationGroupRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	return session.updateStationGroup(req.ID, req.Stations)
}

// DeleteStationGroup implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) DeleteStationGroup(ctx context.Context, req mcom.DeleteStationGroupRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	if err := session.db.Delete(&models.StationGroup{
		ID: req.GroupID,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (dm *DataManager) ListStationIDs(ctx context.Context, req mcom.ListStationIDsRequest) (mcom.ListStationIDsReply, error) {
	session := dm.newSession(ctx)
	return session.listStationIDs(req.DepartmentOID)
}

func (session *session) listStationIDs(departmentOID string) (mcom.ListStationIDsReply, error) {
	type stationID struct {
		ID string
	}
	var stations []stationID
	if err := session.db.Model(&models.Station{}).
		Where(models.Station{
			AdminDepartmentID: departmentOID,
		}).Find(&stations).Error; err != nil {
		return mcom.ListStationIDsReply{}, err
	}

	res := make([]string, len(stations))
	for i, station := range stations {
		res[i] = station.ID
	}

	sort.Strings(res)
	return mcom.ListStationIDsReply{
		Stations: res,
	}, nil
}
