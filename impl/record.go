package impl

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"gorm.io/gorm"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func (session *session) listBatches(conditions models.Batch) ([]models.Batch, error) {
	var results []models.Batch
	err := session.db.
		Where(conditions).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (session *session) parseFeedRecords(workOrders []models.WorkOrder) ([]mcom.FeedRecord, error) {
	var reply []mcom.FeedRecord
	for _, w := range workOrders {
		batches, err := session.listBatches(models.Batch{
			WorkOrder: w.ID,
		})
		if err != nil {
			return nil, err
		}

		for _, batch := range batches {
			var feedMaterials []mcom.FeedMaterial
			records, err := session.listFeedRecordsByIDs(batch.RecordsID)
			if err != nil {
				return nil, err
			}
			for _, rec := range records {
				for _, detail := range rec.Materials {
					for _, material := range detail.Resources {
						feedMaterials = append(feedMaterials, mcom.FeedMaterial{
							ID:          material.ProductID,
							Grade:       material.Grade,
							ResourceID:  material.ResourceID,
							ProductType: material.ProductType,
							Status:      material.Status,
							ExpiryTime:  types.ToTimeNano(material.ExpiryTime),
							Quantity:    material.Quantity,
							Site:        detail.SiteID,
						})
					}
				}
			}

			reply = append(reply, mcom.FeedRecord{
				WorkOrder:   w.ID,
				Batch:       int32(batch.Number),
				RecipeID:    w.RecipeID,
				ProcessName: w.ProcessName,
				ProcessType: w.ProcessType,
				StationID:   w.Station,
				Materials:   feedMaterials,
			})
		}
	}

	return reply, nil
}

// ListFeedRecords implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListFeedRecords(ctx context.Context, req mcom.ListRecordsRequest) (mcom.ListFeedRecordReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListFeedRecordReply{}, err
	}

	session := dm.newSession(ctx)
	// 2. department ID + 日期取得當日排產工單資訊
	workOrders, err := session.listWorkOrders(req.DepartmentID, req.Date)
	if err != nil {
		return nil, err
	}

	return session.parseFeedRecords(workOrders)
}

func (session *session) listFeedRecordsByIDs(ids []string) ([]models.FeedRecord, error) {
	var res []models.FeedRecord
	err := session.db.Model(&models.FeedRecord{}).Where(`id IN ?`, ids).Find(&res).Error
	if err != nil {
		return nil, err
	}
	if len(res) != len(ids) {
		return nil, fmt.Errorf("records not found")
	}
	return res, nil
}

func (session *session) listCollectRecords(conditions models.CollectRecord) ([]models.CollectRecord, error) {
	var results []models.CollectRecord
	err := session.db.
		Where(conditions).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (session *session) parseCollectRecords(workOrders []models.WorkOrder) ([]mcom.CollectRecord, error) {
	// 工單資訊取回所有collect資訊
	var reply []mcom.CollectRecord
	for _, w := range workOrders {
		records, err := session.listCollectRecords(models.CollectRecord{
			WorkOrder: w.ID,
		})
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			resource, err := session.getMaterialResourceInfo(&models.MaterialResource{
				OID: record.ResourceOID,
			})
			if err != nil {
				return nil, err
			}

			reply = append(reply, mcom.CollectRecord{
				ResourceID: resource.ID,
				WorkOrder:  record.WorkOrder,
				RecipeID:   w.RecipeID,
				ProductID:  resource.ProductID,
				BatchCount: record.Detail.BatchCount,
				Quantity:   record.Detail.Quantity,
			})
		}
	}

	// sort by product id.
	sort.Slice(reply, func(i, j int) bool {
		return reply[i].ProductID < reply[j].ProductID
	})

	return reply, nil
}

// GetCollectRecord implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetCollectRecord(ctx context.Context, req mcom.GetCollectRecordRequest) (mcom.GetCollectRecordReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetCollectRecordReply{}, err
	}

	var record models.CollectRecord
	session := dm.newSession(ctx)
	if err := session.db.
		Where(models.CollectRecord{
			WorkOrder: req.WorkOrder,
			Sequence:  req.Sequence,
		}).Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetCollectRecordReply{}, mcomErr.Error{
				Code: mcomErr.Code_RECORD_NOT_FOUND,
			}
		}
		return mcom.GetCollectRecordReply{}, err
	}

	resource, err := session.getMaterialResourceInfo(&models.MaterialResource{
		OID: record.ResourceOID,
	})
	if err != nil {
		return mcom.GetCollectRecordReply{}, err
	}

	wo, err := dm.GetWorkOrder(ctx, mcom.GetWorkOrderRequest{ID: req.WorkOrder})
	if err != nil {
		return mcom.GetCollectRecordReply{}, err
	}

	return mcom.GetCollectRecordReply{
		ResourceID: resource.ID,
		LotNumber:  record.LotNumber,
		WorkOrder:  record.WorkOrder,
		RecipeID:   wo.RecipeID,
		ProductID:  resource.ProductID,
		BatchCount: record.Detail.BatchCount,
		Quantity:   record.Detail.Quantity,
		OperatorID: record.Detail.OperatorID,
	}, nil
}

// ListCollectRecords implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListCollectRecords(ctx context.Context, req mcom.ListRecordsRequest) (mcom.ListCollectRecordsReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListCollectRecordsReply{}, err
	}

	session := dm.newSession(ctx)
	// 2. department ID + 日期取得當日排產工單資訊
	workOrders, err := session.listWorkOrders(req.DepartmentID, req.Date)
	if err != nil {
		return nil, err
	}

	return session.parseCollectRecords(workOrders)
}

// CreateCollectRecord implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateCollectRecord(ctx context.Context, req mcom.CreateCollectRecordRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	return session.createCollectRecord(commonsCtx.UserID(ctx), req, req.ResourceOID)
}

func (session *session) createCollectRecord(operatorID string, req mcom.CreateCollectRecordRequest, resourceOID string) error {
	detail := models.CollectRecordDetail{
		OperatorID: operatorID,
		Quantity:   req.Quantity,
	}
	if req.BatchCount > 0 {
		detail.BatchCount = req.BatchCount
	}

	if err := session.db.Create(&models.CollectRecord{
		WorkOrder:   req.WorkOrder,
		Sequence:    req.Sequence,
		LotNumber:   req.LotNumber,
		Station:     req.Station,
		ResourceOID: resourceOID,
		Detail:      detail,
	}).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return mcomErr.Error{
				Code: mcomErr.Code_RECORD_ALREADY_EXISTS,
				Details: fmt.Sprintf("collect record already exists, work_order: %s, sequence: %d, lot_number: %s",
					req.WorkOrder, req.Sequence, req.LotNumber,
				),
			}
		}
		return err
	}
	return nil
}
