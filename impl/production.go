package impl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/workorder"
)

func (session *session) getReservedSequence(station string, date time.Time, index int32) (int32, error) {
	if station == "" {
		return 0, nil
	}

	resvDate, err := parseDate(date)
	if err != nil {
		return 0, err
	}

	wo := &models.WorkOrder{}
	if err := session.db.Where(` station = ? AND reserved_date = ? `, station, resvDate).
		Order(` reserved_sequence desc `).
		Find(wo).Error; err != nil {
		return 0, err
	}

	seqNo := wo.ReservedSequence + index + 1
	return seqNo, nil
}

func parseWorkOrderInsertData(ctx context.Context, w mcom.CreateWorkOrder, seqNo int32) (models.WorkOrder, error) {
	resvDate, err := parseDate(w.Date)
	if err != nil {
		return models.WorkOrder{}, err
	}

	return models.WorkOrder{
		RecipeID:         w.RecipeID,
		ProcessOID:       w.ProcessOID,
		ProcessName:      w.ProcessName,
		ProcessType:      w.ProcessType,
		DepartmentID:     w.DepartmentOID,
		Status:           w.Status,
		Station:          w.Station,
		ReservedDate:     resvDate,
		ReservedSequence: seqNo,
		Information: models.WorkOrderInformation{
			BatchQuantityDetails: w.BatchesQuantity.Detail(),
			Unit:                 w.Unit,
			Parent:               w.Parent,
		},
		UpdatedBy: commonsCtx.UserID(ctx),
		CreatedBy: commonsCtx.UserID(ctx),
	}, nil
}

// CreateWorkOrders implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateWorkOrders(ctx context.Context, req mcom.CreateWorkOrdersRequest) (mcom.CreateWorkOrdersReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.CreateWorkOrdersReply{}, err
	}

	session := dm.newSession(ctx)

	type productInfoKey struct {
		RecipeID    sql.NullString
		ProcessName string
		ProcessType string
	}

	// key: process oid.
	productInfo := make(map[productInfoKey]models.OutputProduct)

	var products []struct {
		RecipeID    sql.NullString
		ProcessName string `gorm:"column:name"`
		ProcessType string `gorm:"column:type"`
		models.OutputProduct
	}

	conditions := make([][3]string, len(req.WorkOrders))
	for i, r := range req.WorkOrders {
		conditions[i] = [3]string{r.RecipeID, r.ProcessName, r.ProcessType}
	}

	if err := session.db.Model(&models.RecipeProcessDefinition{}).Where(`(recipe_id,name,type) IN ?`, conditions).Find(&products).Error; err != nil {
		return mcom.CreateWorkOrdersReply{}, err
	}

	for _, p := range products {
		productInfo[productInfoKey{
			RecipeID:    p.RecipeID,
			ProcessName: p.ProcessName,
			ProcessType: p.ProcessType,
		}] = models.OutputProduct{
			ID:   p.OutputProduct.ID,
			Type: p.OutputProduct.Type,
		}
	}
	// #endregion get process info

	workOrders := make([]models.WorkOrder, len(req.WorkOrders))
	for i, v := range req.WorkOrders {
		product, ok := productInfo[productInfoKey{
			RecipeID: sql.NullString{
				String: v.RecipeID,
				Valid:  true,
			},
			ProcessName: v.ProcessName,
			ProcessType: v.ProcessType,
		}]
		if !ok {
			return mcom.CreateWorkOrdersReply{}, mcomErr.Error{
				Code:    mcomErr.Code_PROCESS_NOT_FOUND,
				Details: fmt.Sprintf("recipe id: %s, process name: %s, process type: %s", v.RecipeID, v.ProcessName, v.ProcessType),
			}
		}

		seqNo, err := session.getReservedSequence(v.Station, v.Date, int32(i))
		if err != nil {
			return mcom.CreateWorkOrdersReply{}, err
		}

		workOrderInfo, err := parseWorkOrderInsertData(ctx, v, seqNo)
		if err != nil {
			return mcom.CreateWorkOrdersReply{}, err
		}

		workOrderInfo.Information.ProductID = product.ID
		workOrderInfo.Information.ProductType = product.Type
		workOrderInfo.Information.BatchQuantityDetails = v.BatchesQuantity.Detail()

		workOrders[i] = workOrderInfo
	}

	if err := session.db.Create(&workOrders).Error; err != nil {
		return mcom.CreateWorkOrdersReply{}, err
	}

	workOrderIDs := make([]string, len(workOrders))
	for i, v := range workOrders {
		workOrderIDs[i] = v.ID
	}
	return mcom.CreateWorkOrdersReply{IDs: workOrderIDs}, nil
}

func parseUpdatesCondition(ctx context.Context, w mcom.UpdateWorkOrder, currentWorkOrder models.WorkOrder) (*models.WorkOrder, error) {
	resvDate, err := parseDate(w.Date)
	if err != nil {
		return nil, err
	}

	updates := &models.WorkOrder{}
	updates.Status = w.Status
	updates.UpdatedBy = commonsCtx.UserID(ctx)

	if !resvDate.IsZero() {
		updates.ReservedDate = resvDate
	}

	updates.Station = w.Station
	updates.ReservedSequence = w.Sequence
	updates.ProcessOID = w.ProcessOID
	updates.ProcessName = w.ProcessName
	updates.ProcessType = w.ProcessType
	updates.RecipeID = w.RecipeID

	shouldUpdateQuantity := w.BatchesQuantity != nil
	shouldUpdateAbnormality := w.Abnormality != workorder.Abnormality_ABNORMALITY_UNSPECIFIED
	shouldUpdateInformation := shouldUpdateAbnormality || shouldUpdateQuantity

	if shouldUpdateInformation {
		updates.Information = currentWorkOrder.Information
		if shouldUpdateQuantity {
			updates.Information.BatchQuantityDetails = w.BatchesQuantity.Detail()
		}
		if shouldUpdateAbnormality {
			updates.Information.Abnormality = w.Abnormality
		}
	}
	return updates, nil
}

// UpdateWorkOrders implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) UpdateWorkOrders(ctx context.Context, req mcom.UpdateWorkOrdersRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, dm.lockTimeout)
	defer cancel()

	tx := dm.beginTx(ctx)
	defer tx.Rollback() // nolint: errcheck

	// #region get work orders
	ids := make([]string, len(req.Orders))
	for i := range req.Orders {
		ids[i] = req.Orders[i].ID
	}

	var wos []models.WorkOrder

	if err := tx.db.Clauses(clause.Locking{
		Strength: "UPDATE",
	}).Where(`id IN ?`, ids).Find(&wos).Error; err != nil {
		return err
	}

	idWorkOrderPair := make(map[string]models.WorkOrder)
	for _, wo := range wos {
		idWorkOrderPair[wo.ID] = wo
	}

	// #endregion get work orders

	for _, v := range req.Orders {
		updates, err := parseUpdatesCondition(ctx, v, idWorkOrderPair[v.ID])
		if err != nil {
			return err
		}

		if err := tx.db.
			Model(&models.WorkOrder{ID: v.ID}).
			Updates(models.WorkOrder{
				ProcessOID:       updates.ProcessOID,
				ProcessName:      updates.ProcessName,
				ProcessType:      updates.ProcessType,
				RecipeID:         updates.RecipeID,
				Status:           updates.Status,
				Station:          updates.Station,
				ReservedDate:     updates.ReservedDate,
				ReservedSequence: updates.ReservedSequence,
				Information:      updates.Information,
				UpdatedBy:        updates.UpdatedBy,
			}).Error; err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (session *session) getFeedBatchInfo(workOrder string) (int, error) {
	batches := []models.Batch{}
	if err := session.db.Where(`work_order = ?`, workOrder).Find(&batches).Error; err != nil {
		return 0, err
	}
	return len(batches), nil
}

func (session *session) getCollectBatchInfo(workOrder string) (int, decimal.Decimal, error) {
	records := []models.CollectRecord{}
	if err := session.db.Where(`work_order = ?`, workOrder).Find(&records).Error; err != nil {
		return 0, decimal.Zero, err
	}

	collectedSequence := 0
	sum := decimal.Zero

	for _, r := range records {
		if int(r.Sequence) > collectedSequence {
			collectedSequence = int(r.Sequence)
		}
		sum = sum.Add(r.Detail.Quantity)
	}
	return collectedSequence, sum, nil
}

func (session *session) parseListWorkOrder(WorkOrders ...models.WorkOrder) ([]mcom.GetWorkOrderReply, error) {
	workOrders := make([]mcom.GetWorkOrderReply, len(WorkOrders))
	for i, w := range WorkOrders {
		currentBatch, err := session.getFeedBatchInfo(w.ID)
		if err != nil {
			return nil, err
		}

		collectedSequences, sum, err := session.getCollectBatchInfo(w.ID)
		if err != nil {
			return nil, err
		}

		resvDate, err := parseDate(w.ReservedDate)
		if err != nil {
			return nil, err
		}

		workOrders[i] = mcom.GetWorkOrderReply{
			ID: w.ID,
			Product: mcom.Product{
				ID:   w.Information.ProductID,
				Type: w.Information.ProductType,
			},
			Process: mcom.WorkOrderProcess{
				OID:  w.ProcessOID,
				Name: w.ProcessName,
				Type: w.ProcessType,
			},
			RecipeID:             w.RecipeID,
			Status:               w.Status,
			DepartmentOID:        w.DepartmentID,
			Station:              w.Station,
			Sequence:             w.ReservedSequence,
			Date:                 resvDate,
			Unit:                 w.Information.Unit,
			BatchQuantityDetails: w.Information.BatchQuantityDetails,
			CurrentBatch:         currentBatch,
			CollectedSequence:    collectedSequences,
			CollectedQuantity:    sum,
			UpdatedBy:            w.UpdatedBy,
			UpdatedAt:            w.UpdatedAt.Time(),
			InsertedBy:           w.CreatedBy,
			InsertedAt:           w.CreatedAt.Time(),
			Parent:               w.Information.Parent,
			Abnormality:          w.Information.Abnormality,
		}
	}

	return workOrders, nil
}

// GetWorkOrder implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetWorkOrder(ctx context.Context, req mcom.GetWorkOrderRequest) (mcom.GetWorkOrderReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetWorkOrderReply{}, err
	}

	session := dm.newSession(ctx)
	workOrder := models.WorkOrder{}
	if err := session.db.
		Where(&models.WorkOrder{
			ID: req.ID,
		}).Take(&workOrder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetWorkOrderReply{}, mcomErr.Error{
				Code: mcomErr.Code_WORKORDER_NOT_FOUND,
			}
		}
		return mcom.GetWorkOrderReply{}, err
	}

	results, err := session.parseListWorkOrder(workOrder)
	if err != nil {
		return mcom.GetWorkOrderReply{}, err
	}
	return results[0], err
}

// ListWorkOrders implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListWorkOrders(ctx context.Context, req mcom.ListWorkOrdersRequest) (mcom.ListWorkOrdersReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListWorkOrdersReply{}, err
	}

	resvDate, err := parseDate(req.Date)
	if err != nil {
		return mcom.ListWorkOrdersReply{}, err
	}

	condition := &models.WorkOrder{
		Station:      req.Station,
		ReservedDate: resvDate,
	}
	if req.ID != "" {
		condition.ID = req.ID
	}

	session := dm.newSession(ctx)
	workOrders := []models.WorkOrder{}
	if err := session.db.
		Where(condition).
		Order(` reserved_sequence `).
		Find(&workOrders).Error; err != nil {
		return mcom.ListWorkOrdersReply{}, err
	}

	contents, err := session.parseListWorkOrder(workOrders...)
	if err != nil {
		return mcom.ListWorkOrdersReply{}, err
	}

	return mcom.ListWorkOrdersReply{
		WorkOrders: contents,
	}, nil
}

// ListWorkOrdersByDuration implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListWorkOrdersByDuration(ctx context.Context, req mcom.ListWorkOrdersByDurationRequest) (mcom.ListWorkOrdersByDurationReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListWorkOrdersByDurationReply{}, err
	}

	// #region parse date
	since, err := parseDate(req.Since)
	if err != nil {
		return mcom.ListWorkOrdersByDurationReply{}, err
	}
	until, err := parseDate(req.Until)
	if err != nil {
		return mcom.ListWorkOrdersByDurationReply{}, err
	}
	// #endregion parse date

	condition := &models.WorkOrder{
		ID:           req.ID,
		Station:      req.Station,
		DepartmentID: req.DepartmentID,
	}

	session := dm.newSession(ctx)
	workOrders := []models.WorkOrder{}

	sql := session.db.
		Where(condition)

	if len(req.Status) > 0 {
		sql = sql.Where(`status IN ?`, req.Status)
	}

	if since.Equal(until) {
		sql.Where(`reserved_date = ?`, since)
	} else {
		sql.Where(`reserved_date BETWEEN ? AND ?`, since, until)
	}

	sql.Order(`reserved_date`).Order(`reserved_sequence`)

	if req.Limit != 0 {
		sql = sql.Limit(req.Limit)
	}

	if err := sql.
		Find(&workOrders).Error; err != nil {
		return mcom.ListWorkOrdersByDurationReply{}, err
	}

	res, err := session.parseListWorkOrder(workOrders...)
	return mcom.ListWorkOrdersByDurationReply{
		Contents: res,
	}, err
}

func (dm *DataManager) ListWorkOrdersByIDs(ctx context.Context, req mcom.ListWorkOrdersByIDsRequest) (mcom.ListWorkOrdersByIDsReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListWorkOrdersByIDsReply{}, err
	}

	session := dm.newSession(ctx)

	workOrders := []models.WorkOrder{}

	sql := session.db
	sql = sql.Model(&models.WorkOrder{}).Where(`id IN ?`, req.IDs)
	if req.Station != "" {
		sql = sql.Where(`station = ?`, req.Station)
	}
	if err := sql.Find(&workOrders).Error; err != nil {
		return mcom.ListWorkOrdersByIDsReply{}, err
	}

	res := make(map[string]mcom.GetWorkOrderReply)
	reps, err := session.parseListWorkOrder(workOrders...)
	if err != nil {
		return mcom.ListWorkOrdersByIDsReply{}, err
	}
	for _, workOrder := range reps {
		res[workOrder.ID] = workOrder
	}
	return mcom.ListWorkOrdersByIDsReply{Contents: res}, nil
}
