package impl

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
)

// CreateBatch implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateBatch(ctx context.Context, req mcom.CreateBatchRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	newBatch := models.Batch{
		WorkOrder: req.WorkOrder,
		Number:    req.Number,
		Status:    req.Status, // default zero is preparing.
		RecordsID: []string{},
		Note:      req.Note,
		UpdatedBy: commonsCtx.UserID(ctx),
	}

	session := dm.newSession(ctx)
	if err := session.db.Create(&newBatch).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return mcomErr.Error{Code: mcomErr.Code_BATCH_ALREADY_EXISTS}
		}
		return err
	}
	return nil
}

func (dm *DataManager) GetBatch(ctx context.Context, req mcom.GetBatchRequest) (mcom.GetBatchReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetBatchReply{}, err
	}

	var batch models.Batch
	session := dm.newSession(ctx)
	if err := session.db.
		Where(&models.Batch{WorkOrder: req.WorkOrder, Number: req.Number}).
		Take(&batch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetBatchReply{}, mcomErr.Error{Code: mcomErr.Code_BATCH_NOT_FOUND}
		}
		return mcom.GetBatchReply{}, err
	}

	records, err := session.listFeedRecordsByIDs(batch.RecordsID)
	if err != nil {
		return mcom.GetBatchReply{}, err
	}

	return mcom.GetBatchReply{Info: parseBatchInfo(batch, records)}, nil
}

// ListBatches implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListBatches(ctx context.Context, req mcom.ListBatchesRequest) (mcom.ListBatchesReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListBatchesReply{}, err
	}

	session := dm.newSession(ctx)

	joinResult := []struct {
		models.Batch
		models.FeedRecord
	}{}

	if err := session.db.Model(&models.Batch{}).Joins("LEFT JOIN feed_record ON feed_record.id = ANY(batch.records_id)").Where(models.Batch{WorkOrder: req.WorkOrder}).Select("*").Scan(&joinResult).Error; err != nil {
		return nil, err
	}

	type batchPK struct {
		workOrder string
		number    int16
	}

	type batchRecordPair struct {
		batch   models.Batch
		records []models.FeedRecord
	}
	batchRecordPairMap := make(map[batchPK]batchRecordPair)

	for _, res := range joinResult {
		if val, ok := batchRecordPairMap[batchPK{workOrder: res.WorkOrder, number: res.Number}]; ok {
			batchRecordPairMap[batchPK{workOrder: res.WorkOrder, number: res.Number}] = batchRecordPair{
				batch:   val.batch,
				records: append(val.records, res.FeedRecord),
			}
		} else {
			if res.FeedRecord.ID != "" {
				batchRecordPairMap[batchPK{workOrder: res.WorkOrder, number: res.Number}] = batchRecordPair{
					batch:   res.Batch,
					records: []models.FeedRecord{res.FeedRecord},
				}
			} else {
				batchRecordPairMap[batchPK{workOrder: res.WorkOrder, number: res.Number}] = batchRecordPair{
					batch:   res.Batch,
					records: []models.FeedRecord{},
				}
			}
		}
	}

	res := make([]mcom.BatchInfo, len(batchRecordPairMap))
	index := 0
	for _, val := range batchRecordPairMap {
		res[index] = parseBatchInfo(val.batch, val.records)
		index++
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Number < res[j].Number
	})
	return res, nil
}

// UpdateBatch implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) UpdateBatch(ctx context.Context, req mcom.UpdateBatchRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	if result := session.db.Where(models.Batch{WorkOrder: req.WorkOrder, Number: req.Number}).Updates(models.Batch{
		Note:      req.Note,
		Status:    req.Status,
		UpdatedBy: commonsCtx.UserID(ctx),
	}); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return mcomErr.Error{Code: mcomErr.Code_BATCH_NOT_FOUND, Details: "Update target not found"}
	}
	return nil
}

func (tx *txDataManager) appendFeedRecords(req appendFeedRecordRequest) error {
	if err := tx.db.Model(&models.FeedRecord{}).Create(req.Records).Error; err != nil {
		return err
	}

	recordIDs := make([]string, len(req.Records))
	for i, rec := range req.Records {
		recordIDs[i] = rec.ID
	}

	if req.WorkOrder == "" {
		return nil
	}

	return tx.db.Exec("UPDATE batch SET records_id = records_id || ARRAY[?] where work_order = ? AND number = ?", recordIDs, req.WorkOrder, req.Number).Error
}

// appendFeedRecordRequest definition.
type appendFeedRecordRequest struct {
	WorkOrder string
	Number    int16
	Records   []models.FeedRecord
}

func parseBatchInfo(batch models.Batch, records []models.FeedRecord) mcom.BatchInfo {
	return mcom.BatchInfo{
		WorkOrder: batch.WorkOrder,
		Number:    batch.Number,
		Status:    int32(batch.Status),
		Records:   records,
		Note:      batch.Note,
		UpdatedAt: batch.UpdatedAt,
		UpdatedBy: batch.UpdatedBy,
	}
}

// Feed implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) Feed(ctx context.Context, req mcom.FeedRequest) (mcom.FeedReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.FeedReply{}, err
	}

	session := dm.newSession(ctx)

	tx := session.beginTx()
	defer tx.Rollback() // nolint: errcheck

	feedRecord := make([]models.FeedDetail, len(req.FeedContent))

	toUpdateResources := []models.MaterialsWithoutQuantity{}
	for i, elem := range req.FeedContent {
		if materialsWithoutQuantity, err := tx.feed(elem, &feedRecord[i]); err != nil {
			return mcom.FeedReply{}, err
		} else if materialsWithoutQuantity.ResourceID != "" {
			toUpdateResources = append(toUpdateResources, materialsWithoutQuantity)
		}
	}

	feedRecordID := strings.ToUpper(xid.New().String())
	if err := tx.appendFeedRecords(appendFeedRecordRequest{
		WorkOrder: req.Batch.WorkOrder,
		Number:    req.Batch.Number,
		Records: []models.FeedRecord{
			{
				ID:         feedRecordID,
				OperatorID: commonsCtx.UserID(ctx),
				Materials:  feedRecord,
				Time:       time.Now(),
			}},
	}); err != nil {
		return mcom.FeedReply{}, err
	}

	if len(toUpdateResources) != 0 {
		if err := tx.updateFeedResources(toUpdateResources); err != nil {
			return mcom.FeedReply{}, err
		}
	}

	return mcom.FeedReply{FeedRecordID: feedRecordID}, tx.Commit()
}

func (tx *txDataManager) updateFeedResources(toUpdateResources []models.MaterialsWithoutQuantity) error {
	type materialResource struct {
		resourceID  string
		productType string
	}

	resourceQtyPair := make(map[materialResource]decimal.Decimal, len(toUpdateResources))
	condition := make([][2]string, len(toUpdateResources))
	for i := range toUpdateResources {
		condition[i][0] = toUpdateResources[i].ResourceID
		condition[i][1] = toUpdateResources[i].ProductType
		resourceQtyPair[materialResource{
			resourceID:  toUpdateResources[i].ResourceID,
			productType: toUpdateResources[i].ProductType,
		}] = toUpdateResources[i].FedQuantity
	}

	var resources []struct {
		OID               string `gorm:"column:oid;type:uuid"`
		ID                string
		ProductID         string
		ProductType       string
		Quantity          decimal.Decimal
		WarehouseID       string
		WarehouseLocation string
	}

	txWithTimeout, cancel := tx.withTimeout()
	defer cancel()
	if err := txWithTimeout.db.Model(&models.MaterialResource{}).
		Clauses(clause.Locking{
			Strength: "UPDATE",
		}).Where(`(id, product_type) IN ?`, condition).Find(&resources).Error; err != nil {
		return err
	}

	// internal error : resources will be checked while binding.
	if len(resources) != len(toUpdateResources) {
		var ids string
		for _, res := range toUpdateResources {
			ids += " " + res.ResourceID
		}
		return fmt.Errorf("one of following resource not found: %s", ids)
	}

	for _, res := range resources {
		newQuantity := res.Quantity.Sub(resourceQtyPair[materialResource{
			resourceID:  res.ID,
			productType: res.ProductType,
		}])

		if err := tx.db.Model(models.MaterialResource{}).Where(models.MaterialResource{
			OID: res.OID,
		}).Update("quantity", newQuantity).Error; err != nil {
			return err
		}

		if res.WarehouseID != "" && res.WarehouseLocation != "" {
			toSub := resourceQtyPair[materialResource{
				resourceID:  res.ID,
				productType: res.ProductType,
			}]
			if err := tx.db.Model(&models.WarehouseStock{}).
				Where(models.WarehouseStock{
					ID:        res.WarehouseID,
					Location:  res.WarehouseLocation,
					ProductID: res.ProductID,
				}).UpdateColumn("quantity", gorm.Expr("quantity - ?", toSub)).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (tx *txDataManager) feed(fps mcom.FeedPerSite, feedDetail *models.FeedDetail) (models.MaterialsWithoutQuantity, error) {
	if fps1, ok := fps.(mcom.FeedPerSiteType1); ok {
		return tx.feedType1(fps1, feedDetail)
	}

	if fps2, ok := fps.(mcom.FeedPerSiteType2); ok {
		return tx.feedType2(fps2, feedDetail)
	}

	if fps3, ok := fps.(mcom.FeedPerSiteType3); ok {
		return tx.feedType3(fps3, feedDetail)
	}

	return models.MaterialsWithoutQuantity{}, fmt.Errorf("unknown feed type")
}

func (tx *txDataManager) feedType1(fps mcom.FeedPerSiteType1, feedDetail *models.FeedDetail) (materialsWithoutQuantity models.MaterialsWithoutQuantity, err error) {
	var records []models.FedMaterial
	site, contents, err := tx.getSiteAndContents(fps.Site.Station, fps.Site.SiteID.Name, fps.Site.SiteID.Index)
	if err != nil {
		return models.MaterialsWithoutQuantity{}, err
	}
	switch site.Attributes.Type {
	case sites.Type_CONTAINER:
		if fps.FeedAll {
			records = append(records, contents.Content.Container.FeedAll()...)
		} else {
			record, _ := contents.Content.Container.Feed(fps.Quantity)
			records = append(records, record...)
		}
	case sites.Type_SLOT:
		if fps.FeedAll {
			records = append(records, contents.Content.Slot.FeedAll()...)
		} else {
			record, _, toUpdateResource := contents.Content.Slot.Feed(fps.Quantity)
			if toUpdateResource {
				materialsWithoutQuantity = models.MaterialsWithoutQuantity{
					ResourceID:  record[0].ResourceID,
					ProductType: record[0].ProductType,
					FedQuantity: record[0].Quantity,
				}
			}
			records = append(records, record...)
		}
	case sites.Type_COLLECTION:
		if fps.FeedAll {
			records = append(records, contents.Content.Collection.FeedAll()...)
		} else {
			return models.MaterialsWithoutQuantity{}, mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "a collection site can only feed all materials"}
		}
	case sites.Type_QUEUE:
		if fps.FeedAll {
			recs, err := contents.Content.Queue.FeedAll()
			if err != nil {
				return models.MaterialsWithoutQuantity{}, err
			}
			records = append(records, recs...)
		} else {
			recs, _, toUpdateResource, err := contents.Content.Queue.Feed(fps.Quantity)
			if err != nil {
				return models.MaterialsWithoutQuantity{}, err
			}
			if toUpdateResource {
				materialsWithoutQuantity = models.MaterialsWithoutQuantity{
					ResourceID:  recs[0].ResourceID,
					ProductType: recs[0].ProductType,
					FedQuantity: recs[0].Quantity,
				}
			}
			records = append(records, recs...)
		}
	case sites.Type_COLQUEUE:
		if fps.FeedAll {
			recs, err := contents.Content.Colqueue.FeedAll()
			if err != nil {
				return models.MaterialsWithoutQuantity{}, err
			}
			records = append(records, recs...)
		} else {
			return models.MaterialsWithoutQuantity{}, mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "a colqueue  site can only feed all materials"}
		}
	default:
		return models.MaterialsWithoutQuantity{}, fmt.Errorf("undefined site type")
	}

	resources := make([]models.FeedResource, len(records))
	for i, record := range records {
		resources[i] = models.FeedResource{
			ResourceID:  record.ResourceID,
			ProductID:   record.Material.ID,
			ProductType: record.ProductType,
			Grade:       record.Material.Grade,
			Status:      record.Status,
			ExpiryTime:  record.ExpiryTime,
			Quantity:    record.Quantity,
		}
	}
	*feedDetail = models.FeedDetail{
		Resources:  resources,
		UniqueSite: fps.Site,
	}

	if err := tx.db.Where(`station = ? AND name = ? AND index = ?`, fps.Site.Station, fps.Site.SiteID.Name, fps.Site.SiteID.Index).
		Updates(models.SiteContents{Content: contents.Content, UpdatedBy: contents.UpdatedBy}).Error; err != nil {
		return models.MaterialsWithoutQuantity{}, err
	}
	return materialsWithoutQuantity, nil
}

func (tx *txDataManager) feedType2(fps mcom.FeedPerSiteType2, feedDetail *models.FeedDetail) (models.MaterialsWithoutQuantity, error) {
	record, err := tx.feedMaterialDirectly(fps)

	if err != nil {
		return models.MaterialsWithoutQuantity{}, err
	}

	resource := models.FeedResource{
		ResourceID:  record.ResourceID,
		ProductID:   record.Material.ID,
		ProductType: record.ProductType,
		Grade:       record.Material.Grade,
		Status:      record.Status,
		ExpiryTime:  record.ExpiryTime,
		Quantity:    record.Quantity,
	}

	*feedDetail = models.FeedDetail{
		Resources:  []models.FeedResource{resource},
		UniqueSite: fps.Site,
	}

	return models.MaterialsWithoutQuantity{}, nil
}

func (tx *txDataManager) feedType3(fps mcom.FeedPerSiteType3, feedDetail *models.FeedDetail) (models.MaterialsWithoutQuantity, error) {
	return tx.feedType2(mcom.FeedPerSiteType2{
		Quantity:    fps.Quantity,
		ResourceID:  fps.ResourceID,
		ProductType: fps.ProductType,
	}, feedDetail)
}

func (tx *txDataManager) feedMaterialDirectly(fps mcom.FeedPerSiteType2) (models.FedMaterial, error) {
	target := tx.db.Model(&models.MaterialResource{}).Where(models.MaterialResource{ID: fps.ResourceID, ProductType: fps.ProductType})
	if err := target.UpdateColumn(`quantity`, gorm.Expr(`quantity - ?`, fps.Quantity)).Error; err != nil {
		return models.FedMaterial{}, err
	}

	var resource models.MaterialResource
	if err := target.Take(&resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			commonsCtx.Logger(tx.ctx).Warn("resource not found", zap.String("id", fps.ResourceID), zap.String("product type", fps.ProductType))
			resource = models.MaterialResource{
				ID:          fps.ResourceID,
				ProductType: fps.ProductType,
			}
		} else {
			return models.FedMaterial{}, err
		}
	}

	if resource.WarehouseID != "" {
		if err := tx.db.Model(&models.WarehouseStock{}).
			Where(models.WarehouseStock{
				ID:        resource.WarehouseID,
				Location:  resource.WarehouseLocation,
				ProductID: resource.ProductID,
			}).UpdateColumn("quantity", gorm.Expr(`quantity - ?`, fps.Quantity)).
			Error; err != nil {
			return models.FedMaterial{}, err
		}
	}

	return models.FedMaterial{
		Material: models.Material{
			ID:    resource.ProductID,
			Grade: resource.Info.Grade,
		},
		ResourceID:  resource.ID,
		ProductType: resource.ProductType,
		Status:      resource.Status,
		ExpiryTime:  resource.ExpiryTime.Time(),
		Quantity:    fps.Quantity,
	}, nil
}

func (tx *txDataManager) getSiteAndContents(station string, name string, index int16) (models.Site, models.SiteContents, error) {
	var siteContents models.SiteContents
	if err := tx.db.Where(`station = ? AND name = ? AND index = ?`, station, name, index).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "NOWAIT",
		}).Take(&siteContents).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Site{}, models.SiteContents{}, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND}
		}
		return models.Site{}, models.SiteContents{}, err
	}
	var site models.Site
	if err := tx.db.Where(`station = ? AND name = ? AND index = ?`, station, name, index).
		Take(&site).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Site{}, models.SiteContents{}, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND}
		}
		return models.Site{}, models.SiteContents{}, err
	}
	return site, siteContents, nil
}
