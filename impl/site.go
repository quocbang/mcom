package impl

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
)

func (dm *DataManager) BindRecordsCheck(ctx context.Context, req mcom.BindRecordsCheckRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	var rec models.BindRecords
	if err := session.db.
		Model(&models.BindRecords{}).
		Where(`station_id = ? AND site_name = ? AND site_index = ?`, req.Site.Station, req.Site.SiteID.Name, req.Site.SiteID.Index).
		Take(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcomErr.Error{Code: mcomErr.Code_STATION_SITE_BIND_RECORD_NOT_FOUND, Details: "no records in the specified site"}
		}
		return err
	}

	var notFoundList []string
	for _, res := range req.Resources {
		if !rec.Records.Contains(res) {
			notFoundList = append(notFoundList, "resource id:"+res.ResourceID+", product type:"+res.ProductType)
		}
	}

	if len(notFoundList) == 0 {
		return nil
	}
	return mcomErr.Error{
		Code:    mcomErr.Code_STATION_SITE_BIND_RECORD_NOT_FOUND,
		Details: "not found resources: " + strings.Join(notFoundList, ";"),
	}
}

// Deprecated: use V2 instead
func (tx *txDataManager) addBindRecord(req mcom.MaterialResourceBindRequest) error {
	for _, detail := range req.Details {
		newResources := make([]string, len(detail.Resources))
		for i, res := range detail.Resources {
			newResources[i] = fmt.Sprintf(`"%s %s":{}`, res.ResourceID, res.ProductType)
		}
		newResStr := "{" + strings.Join(newResources, ",") + "}"

		sqlStr := "UPDATE " + models.BindRecords{}.TableName() + ` AS br
		SET  records = jsonb_set(br.records, '{set}',br.records::jsonb->'set'||?)
		WHERE station_id = ? AND site_name = ? AND site_index = ?`

		result := tx.db.Exec(sqlStr, newResStr, req.Station, detail.Site.Name, detail.Site.Index)

		if err := result.Error; err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			if err := tx.createBindRecord(req.Station, detail.Site.Name, detail.Site.Index, detail.Resources); err != nil {
				return err
			}
		}
	}
	return nil
}

func (tx *txDataManager) addBindRecordV2(req mcom.MaterialResourceBindRequestV2) error {
	for _, detail := range req.Details {
		newResources := make([]string, len(detail.Resources))
		for i, res := range detail.Resources {
			newResources[i] = fmt.Sprintf(`"%s %s":{}`, res.ResourceID, res.ProductType)
		}
		newResStr := "{" + strings.Join(newResources, ",") + "}"

		sqlStr := "UPDATE " + models.BindRecords{}.TableName() + ` AS br
		SET  records = jsonb_set(br.records, '{set}',br.records::jsonb->'set'||?)
		WHERE station_id = ? AND site_name = ? AND site_index = ?`

		result := tx.db.Exec(sqlStr, newResStr, detail.Site.Station, detail.Site.SiteID.Name, detail.Site.SiteID.Index)

		if err := result.Error; err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			if err := tx.createBindRecord(detail.Site.Station, detail.Site.SiteID.Name, detail.Site.SiteID.Index, detail.Resources); err != nil {
				return err
			}
		}
	}
	return nil
}

func (tx *txDataManager) createBindRecord(station string, siteName string, siteIndex int16, resources []mcom.BindMaterialResource) error {
	rec := models.BindRecords{
		StationID: station,
		SiteName:  siteName,
		SiteIndex: siteIndex,
	}

	for _, res := range resources {
		rec.Records.Add(models.UniqueMaterialResource{
			ResourceID:  res.ResourceID,
			ProductType: res.ProductType,
		})
	}
	return tx.db.Model(&models.BindRecords{}).Create(&rec).Error
}

// MaterialResourceBindV2 implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) MaterialResourceBindV2(ctx context.Context, req mcom.MaterialResourceBindRequestV2) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}
	updatedBy := commonsCtx.UserID(ctx)
	session := dm.newSession(ctx)
	tx, cancel := session.beginTx().withTimeout()
	defer cancel()
	defer tx.Rollback() // nolint: errcheck

	if err := tx.addBindRecordV2(req); err != nil {
		return err
	}

	stationSites, err := tx.getToBindSites(req)
	if err != nil {
		return err
	}

	// #region check sub type
	validSites := make(map[models.UniqueSite]struct{})
	for _, site := range stationSites {
		if site.Information.SubType == sites.SubType_MATERIAL {
			validSites[models.UniqueSite{
				SiteID:  site.Information.SiteID,
				Station: site.Information.Station,
			}] = struct{}{}
		}
	}
	for _, target := range req.Details {
		if _, ok := validSites[target.Site]; !ok {
			return mcomErr.Error{
				Code:    mcomErr.Code_BAD_REQUEST,
				Details: fmt.Sprintf("invalid site sub type, station: %s name: %s index: %d", target.Site.Station, target.Site.SiteID.Name, target.Site.SiteID.Index),
			}
		}
	}
	// #endregion check sub type

	siteTypeMapping := make(map[models.UniqueSite]sites.Type)
	for _, site := range stationSites {
		siteTypeMapping[models.UniqueSite{
			SiteID: models.SiteID{
				Name:  site.Information.SiteID.Name,
				Index: int16(site.Information.SiteID.Index),
			},
			Station: site.Information.Station,
		}] = site.Information.Type
	}

	if err := checkBadRequest(req, siteTypeMapping); err != nil {
		return err
	}

	// parse resources
	resources := []mcom.BindMaterialResource{}
	for _, detail := range req.Details {
		resources = append(resources, detail.Resources...)
	}
	resourcesManager := newResourcesManager(resources)
	contentsPool, err := newSiteContentsPoolV2(stationSites)
	if err != nil {
		return err
	}

	for _, detail := range req.Details {
		site, err := contentsPool.get(detail.Site)
		if err != nil {
			return err
		}
		resources := parseBoundResources(detail.Resources)
		var takenResources []models.BoundResource

		switch detail.Type {
		// container operations
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND:
			takenResources = site.Container.Bind(resources)
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_ADD:
			site.Container.Add(resources)
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAR:
			takenResources = site.Container.Clear()
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAN_DEVIATION:
			site.Container.CleanDeviation()

		// slot operations
		case bindtype.BindType_RESOURCE_BINDING_SLOT_BIND:
			takenResource := site.Slot.Bind(resources[0])
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}
		case bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR:
			takenResource := site.Slot.Clear()
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}

		// collection operations
		case bindtype.BindType_RESOURCE_BINDING_COLLECTION_BIND:
			takenResources = site.Collection.Bind(resources)
		case bindtype.BindType_RESOURCE_BINDING_COLLECTION_ADD:
			site.Collection.Add(resources)
		case bindtype.BindType_RESOURCE_BINDING_COLLECTION_CLEAR:
			takenResources = site.Collection.Clear()

		// collection queue operations
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_BIND:
			takenResources = site.Colqueue.Bind(site.Colqueue.ParseIndex(detail.Option), resources)
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD:
			site.Colqueue.Add(site.Colqueue.ParseIndex(detail.Option), resources)
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_CLEAR:
			takenResources = site.Colqueue.Clear()
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSH:
			site.Colqueue.Push(resources)
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSHPOP:
			takenResources = site.Colqueue.PushPop(resources)
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_POP:
			takenResources = site.Colqueue.Pop()
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_REMOVE:
			takenResources = site.Colqueue.Remove(site.Colqueue.ParseIndex(detail.Option))

		// queue operations
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND:
			takenResource := site.Queue.Bind(site.Queue.ParseIndex(detail.Option), resources[0])
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_POP:
			takenResource := site.Queue.Pop()
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSH:
			site.Queue.Push(resources[0])
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSHPOP:
			takenResource := site.Queue.PushPop(resources[0])
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_CLEAR:
			takenResources = site.Queue.Clear()
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_REMOVE:
			takenResource := site.Queue.Remove(site.Queue.ParseIndex(detail.Option))
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}
		default:
			return fmt.Errorf("bind type unspecified")
		}
		resourcesManager.add(takenResources)
	}

	stocksVariation, err := session.listStocksVariation(commonsCtx.Logger(ctx), resourcesManager)
	if err != nil {
		return err
	}

	resourcesVariation := resourcesManager.listResourcesVariation()

	contentsToUpdate := contentsPool.listContentsToUpdate()

	if err := tx.updateResource(resourcesVariation, updatedBy); err != nil {
		return err
	}
	if err := tx.updateStock(stocksVariation, updatedBy); err != nil {
		return err
	}
	if err := tx.updateSiteContents(contentsToUpdate, updatedBy); err != nil {
		return err
	}

	return tx.Commit()
}

func (tx *txDataManager) getToBindSites(req mcom.ResourceBindRequest) ([]mcom.ListStationSite, error) {
	condition := make([][3]string, len(req.GetDetails()))
	for i, d := range req.GetDetails() {
		condition[i] = [3]string{d.GetUniqueSite().Station, d.GetUniqueSite().SiteID.Name, strconv.Itoa(int(d.GetUniqueSite().SiteID.Index))}
	}

	var sites []models.Site
	var contents []models.SiteContents

	if err := tx.db.Model(&models.Site{}).
		Clauses(clause.Locking{
			Strength: "UPDATE",
		}).Where("(station, name, index) IN ?", condition).
		Find(&sites).Error; err != nil {
		return nil, err
	}

	if err := tx.db.Model(&models.SiteContents{}).
		Clauses(clause.Locking{
			Strength: "UPDATE",
		}).Where("(station, name, index) IN ?", condition).
		Find(&contents).Error; err != nil {
		return nil, err
	}

	siteMap := make(map[models.UniqueSite]models.Site)
	for _, site := range sites {
		siteMap[models.UniqueSite{
			SiteID: models.SiteID{
				Name:  site.Name,
				Index: site.Index,
			},
			Station: site.Station,
		}] = site
	}

	contentMap := make(map[models.UniqueSite]models.SiteContent)
	for _, cont := range contents {
		contentMap[models.UniqueSite{
			SiteID: models.SiteID{
				Name:  cont.Name,
				Index: cont.Index,
			},
			Station: cont.Station,
		}] = cont.Content
	}

	res := make([]mcom.ListStationSite, len(req.GetDetails()))
	for i, d := range req.GetDetails() {
		if _, ok := siteMap[d.GetUniqueSite()]; !ok {
			return nil, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND}
		}
		if _, ok := contentMap[d.GetUniqueSite()]; !ok {
			return nil, fmt.Errorf("site content not found")
		}

		res[i] = mcom.ListStationSite{
			Information: mcom.ListStationSitesInformation{
				UniqueSite: d.GetUniqueSite(),
				Type:       siteMap[d.GetUniqueSite()].Attributes.Type,
				SubType:    siteMap[d.GetUniqueSite()].Attributes.SubType,
				Limitation: siteMap[d.GetUniqueSite()].Attributes.Limitation,
			},
			Content: contentMap[d.GetUniqueSite()],
		}
	}
	return res, nil
}

// MaterialResourceBind implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
// Deprecated: use V2 instead
func (dm *DataManager) MaterialResourceBind(ctx context.Context, req mcom.MaterialResourceBindRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	isSharedSite := req.Station == ""

	updatedBy := commonsCtx.UserID(ctx)
	var stationSites []mcom.ListStationSite
	session := dm.newSession(ctx)
	tx := session.beginTx()
	defer tx.Rollback() // nolint: errcheck

	if err := tx.addBindRecord(req); err != nil {
		return err
	}

	if !isSharedSite {
		station, err := session.getStation(req.Station)
		if err != nil {
			return err
		}
		stationSites, err = tx.listStationSites(station.Sites, selectForUpdate())
		if err != nil {
			return err
		}
	} else {
		temp, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "",
			SiteName:  req.Details[0].Site.Name,
			SiteIndex: req.Details[0].Site.Index,
		})

		if err != nil {
			return err
		}

		stationSites = append(stationSites, mcom.ListStationSite{
			Information: mcom.ListStationSitesInformation{
				UniqueSite: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  temp.Name,
						Index: temp.Index,
					},
					Station: req.Station,
				},
				Type:    temp.Attributes.Type,
				SubType: temp.Attributes.SubType,
			},
			Content: temp.Content,
		})
	}

	// #region check sub type
	validSites := make(map[models.SiteID]struct{})
	for _, site := range stationSites {
		if site.Information.SubType == sites.SubType_MATERIAL {
			validSites[site.Information.SiteID] = struct{}{}
		}
	}
	for _, target := range req.Details {
		if _, ok := validSites[target.Site]; !ok {
			return mcomErr.Error{
				Code:    mcomErr.Code_BAD_REQUEST,
				Details: fmt.Sprintf("invalid site sub type, station: %s name: %s index: %d", req.Station, target.Site.Name, target.Site.Index),
			}
		}
	}
	// #endregion check sub type

	siteTypeMapping := make(map[models.UniqueSite]sites.Type)
	for _, site := range stationSites {
		siteTypeMapping[models.UniqueSite{
			SiteID: models.SiteID{
				Name:  site.Information.SiteID.Name,
				Index: int16(site.Information.SiteID.Index),
			},
			Station: req.Station,
		}] = site.Information.Type
	}

	checkSiteType := func(t sites.Type, site models.UniqueSite) error {
		if siteTypeMapping[site] != t {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "site type mismatch"}
		}
		return nil
	}

	checkReqSiteTypes := func(detail mcom.MaterialBindRequestDetail) error {
		site := models.UniqueSite{
			SiteID: models.SiteID{
				Name:  detail.Site.Name,
				Index: detail.Site.Index,
			},
			Station: req.Station,
		}
		switch detail.Type {
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND:
			if err := checkSiteType(sites.Type_CONTAINER, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_ADD:
			if err := checkSiteType(sites.Type_CONTAINER, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAR:
			if err := checkSiteType(sites.Type_CONTAINER, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAN_DEVIATION:
			if err := checkSiteType(sites.Type_CONTAINER, site); err != nil {
				return err
			}
			// slot operations.
		case bindtype.BindType_RESOURCE_BINDING_SLOT_BIND:
			if err := checkSiteType(sites.Type_SLOT, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR:
			if err := checkSiteType(sites.Type_SLOT, site); err != nil {
				return err
			}

			// collection operations.
		case bindtype.BindType_RESOURCE_BINDING_COLLECTION_BIND:
			if err := checkSiteType(sites.Type_COLLECTION, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_COLLECTION_ADD:
			if err := checkSiteType(sites.Type_COLLECTION, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_COLLECTION_CLEAR:
			if err := checkSiteType(sites.Type_COLLECTION, site); err != nil {
				return err
			}
			// collection queue operations.
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_BIND:
			if err := checkSiteType(sites.Type_COLQUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD:
			if err := checkSiteType(sites.Type_COLQUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_CLEAR:
			if err := checkSiteType(sites.Type_COLQUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSH:
			if err := checkSiteType(sites.Type_COLQUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSHPOP:
			if err := checkSiteType(sites.Type_COLQUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_POP:
			if err := checkSiteType(sites.Type_COLQUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_REMOVE:
			if err := checkSiteType(sites.Type_COLQUEUE, site); err != nil {
				return err
			}
			// queue operations.
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND:
			if err := checkSiteType(sites.Type_QUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_POP:
			if err := checkSiteType(sites.Type_QUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSH:
			if err := checkSiteType(sites.Type_QUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSHPOP:
			if err := checkSiteType(sites.Type_QUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_CLEAR:
			if err := checkSiteType(sites.Type_QUEUE, site); err != nil {
				return err
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_REMOVE:
			if err := checkSiteType(sites.Type_QUEUE, site); err != nil {
				return err
			}
		}
		return nil
	}

	// check bad request.
	for _, detail := range req.Details {
		if !isSharedSite {
			if err := session.checkStationSiteRelation(req.Station, detail.Site.Name, detail.Site.Index); err != nil {
				return err
			}
		}
		if err := checkReqSiteTypes(detail); err != nil {
			return err
		}
	}

	// parse resources.
	resources := []mcom.BindMaterialResource{}
	for _, detail := range req.Details {
		resources = append(resources, detail.Resources...)
	}
	resourcesManager := newResourcesManager(resources)
	contentsPool, err := newSiteContentsPool(stationSites, req.Station)
	if err != nil {
		return err
	}

	for _, detail := range req.Details {
		site, err := contentsPool.get(models.UniqueSite{SiteID: models.SiteID{
			Name:  detail.Site.Name,
			Index: detail.Site.Index,
		}, Station: req.Station})
		if err != nil {
			return err
		}
		resources := parseBoundResources(detail.Resources)
		var takenResources []models.BoundResource

		switch detail.Type {
		// container operations.
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND:
			takenResources = site.Container.Bind(resources)
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_ADD:
			site.Container.Add(resources)
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAR:
			takenResources = site.Container.Clear()
		case bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAN_DEVIATION:
			site.Container.CleanDeviation()

			// slot operations.
		case bindtype.BindType_RESOURCE_BINDING_SLOT_BIND:
			takenResource := site.Slot.Bind(resources[0])
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}
		case bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR:
			takenResource := site.Slot.Clear()
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}

			// collection operations.
		case bindtype.BindType_RESOURCE_BINDING_COLLECTION_BIND:
			takenResources = site.Collection.Bind(resources)
		case bindtype.BindType_RESOURCE_BINDING_COLLECTION_ADD:
			site.Collection.Add(resources)
		case bindtype.BindType_RESOURCE_BINDING_COLLECTION_CLEAR:
			takenResources = site.Collection.Clear()

			// collection queue operations.
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_BIND:
			takenResources = site.Colqueue.Bind(site.Colqueue.ParseIndex(detail.Option), resources)
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD:
			site.Colqueue.Add(site.Colqueue.ParseIndex(detail.Option), resources)
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_CLEAR:
			takenResources = site.Colqueue.Clear()
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSH:
			site.Colqueue.Push(resources)
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSHPOP:
			takenResources = site.Colqueue.PushPop(resources)
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_POP:
			takenResources = site.Colqueue.Pop()
		case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_REMOVE:
			takenResources = site.Colqueue.Remove(site.Colqueue.ParseIndex(detail.Option))

			// queue operations.
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND:
			takenResource := site.Queue.Bind(site.Queue.ParseIndex(detail.Option), resources[0])
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_POP:
			takenResource := site.Queue.Pop()
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSH:
			site.Queue.Push(resources[0])
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSHPOP:
			takenResource := site.Queue.PushPop(resources[0])
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_CLEAR:
			takenResources = site.Queue.Clear()
		case bindtype.BindType_RESOURCE_BINDING_QUEUE_REMOVE:
			takenResource := site.Queue.Remove(site.Queue.ParseIndex(detail.Option))
			if takenResource.Material != nil {
				takenResources = []models.BoundResource{takenResource}
			}
		default:
			return fmt.Errorf("bind type unspecified")
		}
		resourcesManager.add(takenResources)
	}

	stocksVariation, err := session.listStocksVariation(commonsCtx.Logger(ctx), resourcesManager)
	if err != nil {
		return err
	}

	resourcesVariation := resourcesManager.listResourcesVariation()

	contentsToUpdate := contentsPool.listContentsToUpdate()

	if err := tx.updateResource(resourcesVariation, updatedBy); err != nil {
		return err
	}
	if err := tx.updateStock(stocksVariation, updatedBy); err != nil {
		return err
	}
	if err := tx.updateSiteContents(contentsToUpdate, updatedBy); err != nil {
		return err
	}

	return tx.Commit()
}

func checkSiteType(t sites.Type, site models.UniqueSite, siteTypeMapping map[models.UniqueSite]sites.Type) error {
	if siteTypeMapping[site] != t {
		return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "site type mismatch"}
	}
	return nil
}

func checkReqSiteTypes(detail mcom.MaterialBindRequestDetailV2, siteTypeMapping map[models.UniqueSite]sites.Type) error {
	site := detail.Site
	switch detail.Type {
	case bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND,
		bindtype.BindType_RESOURCE_BINDING_CONTAINER_ADD,
		bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAR,
		bindtype.BindType_RESOURCE_BINDING_CONTAINER_CLEAN_DEVIATION:
		if err := checkSiteType(sites.Type_CONTAINER, site, siteTypeMapping); err != nil {
			return err
		}
	// slot operations
	case bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
		bindtype.BindType_RESOURCE_BINDING_SLOT_CLEAR:
		if err := checkSiteType(sites.Type_SLOT, site, siteTypeMapping); err != nil {
			return err
		}

	// collection operations
	case bindtype.BindType_RESOURCE_BINDING_COLLECTION_BIND,
		bindtype.BindType_RESOURCE_BINDING_COLLECTION_ADD,
		bindtype.BindType_RESOURCE_BINDING_COLLECTION_CLEAR:
		if err := checkSiteType(sites.Type_COLLECTION, site, siteTypeMapping); err != nil {
			return err
		}
	// collection queue operations
	case bindtype.BindType_RESOURCE_BINDING_COLQUEUE_BIND,
		bindtype.BindType_RESOURCE_BINDING_COLQUEUE_ADD,
		bindtype.BindType_RESOURCE_BINDING_COLQUEUE_CLEAR,
		bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSH,
		bindtype.BindType_RESOURCE_BINDING_COLQUEUE_PUSHPOP,
		bindtype.BindType_RESOURCE_BINDING_COLQUEUE_POP,
		bindtype.BindType_RESOURCE_BINDING_COLQUEUE_REMOVE:
		if err := checkSiteType(sites.Type_COLQUEUE, site, siteTypeMapping); err != nil {
			return err
		}
	// queue operations
	case bindtype.BindType_RESOURCE_BINDING_QUEUE_BIND,
		bindtype.BindType_RESOURCE_BINDING_QUEUE_POP,
		bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSH,
		bindtype.BindType_RESOURCE_BINDING_QUEUE_PUSHPOP,
		bindtype.BindType_RESOURCE_BINDING_QUEUE_CLEAR,
		bindtype.BindType_RESOURCE_BINDING_QUEUE_REMOVE:
		if err := checkSiteType(sites.Type_QUEUE, site, siteTypeMapping); err != nil {
			return err
		}
	default:
		return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "undefined bind type"}
	}
	return nil
}

func checkBadRequest(req mcom.MaterialResourceBindRequestV2, siteTypeMapping map[models.UniqueSite]sites.Type) error {
	for _, detail := range req.Details {
		if err := checkReqSiteTypes(detail, siteTypeMapping); err != nil {
			return err
		}
	}
	return nil
}

func parseBoundResources(resources []mcom.BindMaterialResource) []models.BoundResource {
	res := make([]models.BoundResource, len(resources))
	for i := range resources {
		res[i] = resources[i].ToBoundResource()
	}
	return res
}

func (tx *txDataManager) updateStock(stocksVariation map[updatedWarehouseStock]decimal.Decimal, updatedBy string) error {
	updatedStocks := make([]models.WarehouseStock, len(stocksVariation))
	var count int
	for key, val := range stocksVariation {
		updatedStocks[count] = models.WarehouseStock{
			ID:        key.ID,
			Location:  key.Location,
			ProductID: key.ProductID,
			Quantity:  val,
		}
		count++
	}

	if len(updatedStocks) > 0 {
		if err := tx.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "id"},
				{Name: "location"},
				{Name: "product_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"quantity": gorm.Expr("warehouse_stock.quantity+EXCLUDED.quantity"),
			}),
		}).Create(updatedStocks).Error; err != nil {
			return err
		}
	}
	return nil
}

func (tx *txDataManager) updateResource(resourcesVariation map[updatedResource]decimal.Decimal, updatedBy string) error {
	condition := make([][]string, len(resourcesVariation))
	var count int
	for key := range resourcesVariation {
		condition[count] = []string{key.id, key.productType}
		count++
	}

	var updatedResources []models.MaterialResource
	if err := tx.db.Model(&models.MaterialResource{}).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "NOWAIT",
		}).Where(`(id, product_type) IN ?`, condition).
		Find(&updatedResources).Error; err != nil {
		return err
	}

	for i := range updatedResources {
		updatedResources[i].Quantity = updatedResources[i].Quantity.
			Add(resourcesVariation[updatedResource{
				id:          updatedResources[i].ID,
				productType: updatedResources[i].ProductType,
			}])
		updatedResources[i].UpdatedBy = updatedBy
	}

	for _, res := range updatedResources {
		result := tx.db.Where(&models.MaterialResource{OID: res.OID}).Updates(models.MaterialResource{
			Quantity:          res.Quantity,
			Status:            res.Status,
			ExpiryTime:        res.ExpiryTime,
			Info:              res.Info,
			WarehouseID:       res.WarehouseID,
			WarehouseLocation: res.WarehouseLocation,
			FeedRecordsID:     res.FeedRecordsID,
			Station:           res.Station,
			UpdatedBy:         res.UpdatedBy,
		})
		if err := result.Error; err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			logger, err := zap.NewDevelopment()
			if err != nil {
				log.Fatal(err)
			}
			logger.Info("update resource table failed because of resource not found: " + res.ID)
		}
	}
	return nil
}

func (tx *txDataManager) updateSiteContents(contentsToUpdate map[models.UniqueSite]models.SiteContent, updatedBy string) error {
	condition := make([][]string, len(contentsToUpdate))
	var count int
	for key := range contentsToUpdate {
		condition[count] = []string{key.SiteID.Name, strconv.Itoa(int(key.SiteID.Index)), key.Station}
		count++
	}
	var updatedContents []models.SiteContents
	if err := tx.db.Model(&models.SiteContents{}).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "NOWAIT",
		}).Where(`(name, index, station) IN ?`, condition).
		Find(&updatedContents).Error; err != nil {
		return err
	}
	for i := range updatedContents {
		updatedContents[i].Content = contentsToUpdate[models.UniqueSite{
			SiteID: models.SiteID{
				Name:  updatedContents[i].Name,
				Index: updatedContents[i].Index,
			},
			Station: updatedContents[i].Station,
		}]
		updatedContents[i].UpdatedBy = updatedBy
	}
	for _, res := range updatedContents {
		if err := tx.db.Where(`name = ? AND index = ? AND station = ?`, res.Name, res.Index, res.Station).Updates(models.SiteContents{
			Content:   res.Content,
			UpdatedBy: res.UpdatedBy,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

type updatedResource struct {
	id          string
	productType string
}

type updatedWarehouseStock struct {
	ID        string
	Location  string
	ProductID string
}

type resourcesManager struct {
	resources                         []mcom.BindMaterialResource
	resourcesWithUnknownStockLocation []models.BoundResource
}

func newResourcesManager(resources []mcom.BindMaterialResource) resourcesManager {
	return resourcesManager{resources: resources, resourcesWithUnknownStockLocation: []models.BoundResource{}}
}

func (rm *resourcesManager) add(resources []models.BoundResource) {
	notNilResources := []models.BoundResource{}
	for _, res := range resources {
		if res.Material.Quantity != nil {
			notNilResources = append(notNilResources, res)
		}
	}
	rm.resourcesWithUnknownStockLocation = append(rm.resourcesWithUnknownStockLocation, notNilResources...)
}

func (rm resourcesManager) listResourcesVariation() map[updatedResource]decimal.Decimal {
	variation := make(map[updatedResource]decimal.Decimal)
	for _, res := range rm.resources {
		resourceToUpdate := updatedResource{
			id:          res.ResourceID,
			productType: res.ProductType,
		}
		if res.Quantity != nil {
			if val, ok := variation[resourceToUpdate]; ok {
				variation[resourceToUpdate] = val.Sub(*res.Quantity)
			} else {
				variation[resourceToUpdate] = res.Quantity.Neg()
			}
		}
	}

	for _, res := range rm.resourcesWithUnknownStockLocation {
		resourceToUpdate := updatedResource{
			id:          res.Material.ResourceID,
			productType: res.Material.ProductType,
		}
		if val, ok := variation[resourceToUpdate]; ok {
			variation[resourceToUpdate] = val.Add(*res.Material.Quantity)
		} else {
			variation[resourceToUpdate] = *res.Material.Quantity
		}
	}
	return variation
}

func (session *session) listStocksVariation(logger *zap.Logger, rm resourcesManager) (map[updatedWarehouseStock]decimal.Decimal, error) {
	locationMapping := make(map[struct{ resourceID, productType string }]mcom.Warehouse)

	variation := make(map[updatedWarehouseStock]decimal.Decimal)
	for _, res := range rm.resources {
		if res.Warehouse.ID == "" || res.Warehouse.Location == "" {
			continue
		}
		stockToUpdate := updatedWarehouseStock{
			ID:        res.Warehouse.ID,
			Location:  res.Warehouse.Location,
			ProductID: res.Material.ID,
		}

		if res.Quantity != nil {
			if val, ok := variation[stockToUpdate]; ok {
				variation[stockToUpdate] = val.Sub(*res.Quantity)
			} else {
				variation[stockToUpdate] = res.Quantity.Neg()
			}
		}

		locationMapping[struct {
			resourceID  string
			productType string
		}{resourceID: res.ResourceID, productType: res.ProductType}] = mcom.Warehouse{
			ID:       res.Warehouse.ID,
			Location: res.Warehouse.Location,
		}
	}

	condition := [][]string{}
	for _, res := range rm.resourcesWithUnknownStockLocation {
		if res.Material.ProductType == "" {
			logger.Info("update warehouse stock failed because of missing product type: resource id = " + res.Material.ResourceID)
			continue
		}
		if res.Material.Material.ID == "" {
			logger.Info("update warehouse stock failed because of missing product id: resource id = " + res.Material.ResourceID)
			continue
		}
		condition = append(condition, []string{res.Material.Material.ID, res.Material.ProductType})
	}

	var resources []models.MaterialResource
	if err := session.db.Model(&models.MaterialResource{}).Where(`(id, product_type) IN ?`, condition).Find(&resources).Error; err != nil {
		return nil, err
	}

	for _, res := range resources {
		locationMapping[struct {
			resourceID  string
			productType string
		}{
			resourceID:  res.ID,
			productType: res.ProductID,
		}] = mcom.Warehouse{
			ID:       res.WarehouseID,
			Location: res.WarehouseLocation,
		}
	}

	for _, unknown := range rm.resourcesWithUnknownStockLocation {
		warehouse, ok := locationMapping[struct {
			resourceID  string
			productType string
		}{
			resourceID:  unknown.Material.ResourceID,
			productType: unknown.Material.ProductType,
		}]

		if !ok || warehouse.ID == "" || warehouse.Location == "" {
			continue
		}

		stockToUpdate := updatedWarehouseStock{
			ID:        warehouse.ID,
			Location:  warehouse.Location,
			ProductID: unknown.Material.Material.ID,
		}

		if val, ok := variation[stockToUpdate]; ok {
			variation[stockToUpdate] = val.Add(*unknown.Material.Quantity)
		} else {
			variation[stockToUpdate] = *unknown.Material.Quantity
		}
	}
	return variation, nil
}

type siteContentsPool struct {
	sites map[models.UniqueSite]mcom.ListStationSite
}

func (cp siteContentsPool) listContentsToUpdate() map[models.UniqueSite]models.SiteContent {
	res := make(map[models.UniqueSite]models.SiteContent)
	for key, site := range cp.sites {
		res[key] = site.Content
	}
	return res
}

func newSiteContentsPoolV2(sites []mcom.ListStationSite) (siteContentsPool, error) {
	indexMap := make(map[models.UniqueSite]mcom.ListStationSite)
	for _, site := range sites {
		id := models.UniqueSite{
			SiteID:  site.Information.SiteID,
			Station: site.Information.Station,
		}
		if _, ok := indexMap[id]; ok {
			return siteContentsPool{}, fmt.Errorf("duplicated target sites")
		}
		indexMap[id] = site
	}
	return siteContentsPool{
		sites: indexMap,
	}, nil
}

// Deprecated: use V2 instead
func newSiteContentsPool(sites []mcom.ListStationSite, station string) (siteContentsPool, error) {
	indexMap := make(map[models.UniqueSite]mcom.ListStationSite)
	for _, site := range sites {
		id := models.UniqueSite{
			SiteID:  site.Information.SiteID,
			Station: station,
		}
		if _, ok := indexMap[id]; ok {
			return siteContentsPool{}, fmt.Errorf("duplicated target sites")
		}
		indexMap[id] = site
	}
	return siteContentsPool{
		sites: indexMap,
	}, nil
}

func (cp *siteContentsPool) get(site models.UniqueSite) (*models.SiteContent, error) {
	if val, ok := cp.sites[site]; ok {
		return &val.Content, nil
	}
	return nil, fmt.Errorf("site not found")
}

// ListSiteMaterials implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListSiteMaterials(ctx context.Context, req mcom.ListSiteMaterialsRequest) (mcom.ListSiteMaterialsReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListSiteMaterialsReply{}, err
	}

	session := dm.newSession(ctx)
	if req.Station != "" {
		if err := session.checkStationSiteRelation(req.Station, req.Site.Name, int16(req.Site.Index)); err != nil {
			return nil, err
		}
	}

	siteAttributes, siteContent, err := session.getSiteInfo(models.UniqueSite{SiteID: models.SiteID{
		Name:  req.Site.Name,
		Index: req.Site.Index,
	}, Station: req.Station})
	if err != nil {
		return nil, err
	}

	materials := []mcom.SiteMaterial{}
	switch siteAttributes.Type {
	case sites.Type_CONTAINER:
		for _, res := range *siteContent.Container {
			materials = append(materials, mcom.SiteMaterial{
				ID:         res.Material.Material.ID,
				Grade:      res.Material.Material.Grade,
				ResourceID: res.Material.ResourceID,
				Quantity:   res.Material.Quantity,
				ExpiryTime: res.Material.ExpiryTime.Time(),
			})
		}
	case sites.Type_SLOT:
		if siteContent.Slot.Material != nil {
			materials = append(materials, mcom.SiteMaterial{
				ID:         siteContent.Slot.Material.Material.ID,
				Grade:      siteContent.Slot.Material.Material.Grade,
				ResourceID: siteContent.Slot.Material.ResourceID,
				Quantity:   siteContent.Slot.Material.Quantity,
				ExpiryTime: siteContent.Slot.Material.ExpiryTime.Time(),
			})
		}
	case sites.Type_COLQUEUE:
		for _, set := range *siteContent.Colqueue {
			for _, res := range set {
				materials = append(materials, mcom.SiteMaterial{
					ID:         res.Material.Material.ID,
					Grade:      res.Material.Material.Grade,
					ResourceID: res.Material.ResourceID,
					Quantity:   res.Material.Quantity,
					ExpiryTime: res.Material.ExpiryTime.Time(),
				})
			}
		}
	case sites.Type_COLLECTION:
		for _, res := range *siteContent.Collection {
			materials = append(materials, mcom.SiteMaterial{
				ID:         res.Material.Material.ID,
				Grade:      res.Material.Material.Grade,
				ResourceID: res.Material.ResourceID,
				Quantity:   res.Material.Quantity,
				ExpiryTime: res.Material.ExpiryTime.Time(),
			})
		}
	case sites.Type_QUEUE:
		for _, slot := range *siteContent.Queue {
			if slot.Material != nil {
				materials = append(materials, mcom.SiteMaterial{
					ID:         slot.Material.Material.ID,
					Grade:      slot.Material.Material.Grade,
					ResourceID: slot.Material.ResourceID,
					Quantity:   slot.Material.Quantity,
					ExpiryTime: slot.Material.ExpiryTime.Time(),
				})
			}
		}
	default:
		return nil, fmt.Errorf("unsupported site type: %v", siteAttributes.Type)
	}

	// sort by id.
	sort.Slice(materials, func(i, j int) bool {
		return materials[i].ID < materials[j].ID
	})

	return materials, nil
}

// getSiteInfo return site attributes and site content.
func (session *session) getSiteInfo(site models.UniqueSite) (models.SiteAttributes, models.SiteContent, error) {
	attributes, err := session.getSiteAttributes(site)
	if err != nil {
		return models.SiteAttributes{}, models.SiteContent{}, err
	}

	content, err := session.getSiteContent(site)
	if err != nil {
		return models.SiteAttributes{}, models.SiteContent{}, err
	}

	return attributes, content, nil
}

// getSiteAttributes returns site attributes.
func (session *session) getSiteAttributes(site models.UniqueSite) (models.SiteAttributes, error) {
	// get site info.
	var siteInfo models.Site
	if err := session.db.Where(` name = ? AND index = ? AND station = ?`, site.SiteID.Name, site.SiteID.Index, site.Station).
		Take(&siteInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.SiteAttributes{}, mcomErr.Error{
				Code:    mcomErr.Code_STATION_SITE_NOT_FOUND,
				Details: fmt.Sprintf("site not found, station: %s, name: %s, index: %d", site.Station, site.SiteID.Name, site.SiteID.Index),
			}
		}
		return models.SiteAttributes{}, err
	}

	return siteInfo.Attributes, nil
}

// getSiteContent returns site content.
func (session *session) getSiteContent(site models.UniqueSite) (models.SiteContent, error) {
	// get site info.
	var siteInfo models.SiteContents
	if err := session.db.Where(` name = ? AND index = ? AND station = ?`, site.SiteID.Name, site.SiteID.Index, site.Station).
		Take(&siteInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.SiteContent{}, fmt.Errorf("site content not found, station: %s, name: %s, index: %d", site.Station, site.SiteID.Name, site.SiteID.Index)
		}
		return models.SiteContent{}, err
	}

	return siteInfo.Content, nil
}
