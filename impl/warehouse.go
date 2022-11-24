package impl

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
)

// GetMaterialResource implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetMaterialResource(ctx context.Context, req mcom.GetMaterialResourceRequest) (mcom.GetMaterialResourceReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetMaterialResourceReply{}, err
	}

	if req.ResourceID == "" {
		return mcom.GetMaterialResourceReply{}, mcomErr.Error{Code: mcomErr.Code_RESOURCE_NOT_FOUND}
	}

	session := dm.newSession(ctx)

	resources, err := session.getMaterialResources(req.ResourceID)
	if err != nil {
		return mcom.GetMaterialResourceReply{}, err
	}
	if len(resources) == 0 {
		return mcom.GetMaterialResourceReply{}, mcomErr.Error{
			Code: mcomErr.Code_RESOURCE_NOT_FOUND,
		}
	}

	materialReplies := parseMaterialReplies(resources)

	sort.Slice(materialReplies, func(i, j int) bool { return materialReplies[i].Material.Type < materialReplies[j].Material.Type })

	return materialReplies, nil
}

func parseMaterialReplies(resources []models.MaterialResource) []mcom.MaterialReply {
	results := make([]mcom.MaterialReply, len(resources))
	for i, res := range resources {
		results[i] = mcom.MaterialReply{
			Material: mcom.Material{
				Type:            res.ProductType,
				ID:              res.ProductID,
				Grade:           res.Info.Grade,
				Status:          res.Status,
				Quantity:        res.Quantity,
				PlannedQuantity: res.Info.PlannedQuantity,
				Unit:            res.Info.Unit,
				LotNumber:       res.Info.LotNumber,
				ProductionTime:  res.Info.ProductionTime.Local(),
				ExpiryTime:      res.ExpiryTime.Time(),
				ResourceID:      res.ID,
				MinDosage:       res.Info.MinDosage,
				Inspections:     res.Info.Inspections,
				Remark:          res.Info.Remark,
				// todo set actual value.
				CarrierID:     "",
				Station:       res.Station,
				UpdatedAt:     res.UpdatedAt,
				UpdatedBy:     res.UpdatedBy,
				CreatedAt:     res.CreatedAt,
				CreatedBy:     res.CreatedBy,
				FeedRecordsID: res.FeedRecordsID,
			},
			Warehouse: mcom.Warehouse{
				ID:       res.WarehouseID,
				Location: res.WarehouseLocation,
			},
		}
	}
	return results
}

func (session *session) getMaterialResources(resourceID string) ([]models.MaterialResource, error) {
	var resources []models.MaterialResource
	err := session.db.Where(&models.MaterialResource{ID: resourceID}).Find(&resources).Error
	return resources, err
}

// GetMaterialResourceIdentity implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetMaterialResourceIdentity(ctx context.Context, req mcom.GetMaterialResourceIdentityRequest) (mcom.GetMaterialResourceIdentityReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetMaterialResourceIdentityReply{}, err
	}

	if req.ResourceID == "" || req.ProductType == "" {
		return mcom.GetMaterialResourceIdentityReply{}, mcomErr.Error{Code: mcomErr.Code_RESOURCE_NOT_FOUND}
	}

	getMaterialResourceReply, err := dm.GetMaterialResource(ctx, mcom.GetMaterialResourceRequest{ResourceID: req.ResourceID})
	if err != nil {
		return mcom.GetMaterialResourceIdentityReply{}, err
	}
	var res mcom.GetMaterialResourceIdentityReply
	var count int
	for _, material := range getMaterialResourceReply {
		if material.Material.Type == req.ProductType {
			res = mcom.GetMaterialResourceIdentityReply(material)
			count++
		}
	}
	if count == 0 {
		return mcom.GetMaterialResourceIdentityReply{}, mcomErr.Error{Code: mcomErr.Code_RESOURCE_NOT_FOUND}
	}
	if count == 1 {
		return res, nil
	}

	// count > 1 ： 資料庫會擋，基本上不可能發生
	return mcom.GetMaterialResourceIdentityReply{}, fmt.Errorf("duplicated product type")
}

func (dm *DataManager) ListMaterialResourceIdentities(ctx context.Context, req mcom.ListMaterialResourceIdentitiesRequest) (mcom.ListMaterialResourceIdentitiesReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListMaterialResourceIdentitiesReply{}, err
	}

	condition := make([][]string, len(req.Details))
	for i, r := range req.Details {
		condition[i] = []string{r.ResourceID, r.ProductType}
	}

	resources := []models.MaterialResource{}
	session := dm.newSession(ctx)
	if err := session.db.Where(`(id, product_type) IN ?`, condition).Find(&resources).Error; err != nil {
		return mcom.ListMaterialResourceIdentitiesReply{}, err
	}

	order := make(map[mcom.GetMaterialResourceIdentityRequest]int)
	for i, r := range req.Details {
		order[r] = i
	}

	res := make([]*mcom.MaterialReply, len(req.Details))

	parsedResources := parseMaterialReplies(resources)
	for _, r := range parsedResources {
		res[order[mcom.GetMaterialResourceIdentityRequest{
			ResourceID:  r.Material.ResourceID,
			ProductType: r.Material.Type,
		}]] = &mcom.MaterialReply{
			Material:  r.Material,
			Warehouse: r.Warehouse,
		}
	}

	return mcom.ListMaterialResourceIdentitiesReply{Replies: res}, nil
}

func (dm *DataManager) ListMaterialResources(ctx context.Context, req mcom.ListMaterialResourcesRequest) (mcom.ListMaterialResourcesReply, error) {
	session := dm.newSession(ctx)
	condition := func(db *gorm.DB) *gorm.DB {
		res := db.Model(&models.MaterialResource{}).
			Where(models.MaterialResource{
				ProductID: req.ProductID,
				Status:    req.Status,
			}).Where(`product_type = ?`, req.ProductType)
		if req.CreatedAt != 0 {
			res = res.Where(`created_at >= ?`, req.CreatedAt)
		}
		return res
	}

	dataCount, materialResources, err := listHandler[models.MaterialResource](&session, req, condition)
	if err != nil {
		return mcom.ListMaterialResourcesReply{}, err
	}

	returnValues := make([]mcom.MaterialReply, len(materialResources))
	for i, resource := range materialResources {
		returnValues[i] = mcom.MaterialReply{
			Material: mcom.Material{
				Type:            resource.ProductType,
				ID:              resource.ProductID,
				Grade:           resource.Info.Grade,
				Status:          resource.Status,
				Quantity:        resource.Quantity,
				PlannedQuantity: resource.Info.PlannedQuantity,
				Unit:            resource.Info.Unit,
				LotNumber:       resource.Info.LotNumber,
				ProductionTime:  resource.Info.ProductionTime,
				ExpiryTime:      resource.ExpiryTime.Time(),
				ResourceID:      resource.ID,
				UpdatedAt:       resource.UpdatedAt,
				UpdatedBy:       resource.UpdatedBy,
				CreatedAt:       resource.CreatedAt,
				CreatedBy:       resource.CreatedBy,
				MinDosage:       resource.Info.MinDosage,
				Inspections:     resource.Info.Inspections,
				Remark:          resource.Info.Remark,
				Station:         resource.Station,
				CarrierID:       "", // TODO: set value.
				FeedRecordsID:   resource.FeedRecordsID,
			},
			Warehouse: mcom.Warehouse{
				ID:       resource.WarehouseID,
				Location: resource.WarehouseLocation,
			},
		}
	}

	return mcom.ListMaterialResourcesReply{
		Resources: returnValues,
		PaginationReply: mcom.PaginationReply{
			AmountOfData: dataCount,
		},
	}, nil
}

// ListMaterialResourcesById implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListMaterialResourcesById(ctx context.Context, req mcom.ListMaterialResourcesByIdRequest) (mcom.ListMaterialResourcesByIdReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListMaterialResourcesByIdReply{}, err
	}

	res := make([]map[string]mcom.MaterialReply, len(req.ResourcesID))

	// #region initialize maps in slice
	for i := 0; i < len(res); i++ {
		res[i] = make(map[string]mcom.MaterialReply)
	}
	// #endregion initialize maps in slice

	session := dm.newSession(ctx)
	var resources []models.MaterialResource
	if err := session.db.Model(&models.MaterialResource{}).Where(`id IN ?`, req.ResourcesID).Find(&resources).Error; err != nil {
		return mcom.ListMaterialResourcesByIdReply{}, err
	}

	resourceIndexPair := make(map[string]int)
	for index, id := range req.ResourcesID {
		resourceIndexPair[id] = index
	}

	for _, resource := range resources {
		res[resourceIndexPair[resource.ID]][resource.ProductType] = mcom.MaterialReply{
			Material: mcom.Material{
				Type:            resource.ProductType,
				ID:              resource.ProductID,
				Grade:           resource.Info.Grade,
				Status:          resource.Status,
				Quantity:        resource.Quantity,
				PlannedQuantity: resource.Info.PlannedQuantity,
				Unit:            resource.Info.Unit,
				LotNumber:       resource.Info.LotNumber,
				ProductionTime:  resource.Info.ProductionTime.Local(),
				ExpiryTime:      resource.ExpiryTime.Time(),
				ResourceID:      resource.ID,
				MinDosage:       resource.Info.MinDosage,
				Inspections:     resource.Info.Inspections,
				Remark:          resource.Info.Remark,
				// todo set actual value.
				CarrierID:     "",
				Station:       resource.Station,
				UpdatedAt:     resource.UpdatedAt,
				UpdatedBy:     resource.UpdatedBy,
				CreatedAt:     resource.CreatedAt,
				CreatedBy:     resource.CreatedBy,
				FeedRecordsID: resource.FeedRecordsID,
			},
			Warehouse: mcom.Warehouse{
				ID:       resource.WarehouseID,
				Location: resource.WarehouseLocation,
			},
		}
	}
	return mcom.ListMaterialResourcesByIdReply{
		TypeResourcePairs: res,
	}, nil
}

// GetResourceWarehouse implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetResourceWarehouse(ctx context.Context, req mcom.GetResourceWarehouseRequest) (mcom.GetResourceWarehouseReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetResourceWarehouseReply{}, err
	}

	session := dm.newSession(ctx)
	resources, err := session.getMaterialResources(req.ResourceID)
	if err != nil {
		return mcom.GetResourceWarehouseReply{}, err
	}
	if len(resources) == 0 {
		return mcom.GetResourceWarehouseReply{}, mcomErr.Error{
			Code:    mcomErr.Code_RESOURCE_NOT_FOUND,
			Details: fmt.Sprintf("warehouse resource not found: %v", req.ResourceID),
		}
	}

	// 同ID的條碼應該在一樣的位置
	warehouseID := resources[0].WarehouseID
	warehouseLocation := resources[0].WarehouseLocation
	for _, resource := range resources {
		if resource.WarehouseID != warehouseID ||
			resource.WarehouseLocation != warehouseLocation {
			commonsCtx.Logger(ctx).Info("material resources with the same ID must have the same warehouse_id and warehouse_location",
				zap.String("resource_id", resource.ID))
		}
	}

	return mcom.GetResourceWarehouseReply{
		ID:       resources[0].WarehouseID,
		Location: resources[0].WarehouseLocation,
	}, nil
}

// WarehousingStock implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) WarehousingStock(ctx context.Context, req mcom.WarehousingStockRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)

	resources, err := session.listResources(req.ResourceIDs)
	if err != nil {
		return err
	}

	tx := dm.beginTx(ctx)
	defer tx.Rollback() // nolint: errcheck

	if err := tx.updateWarehouseStocks(req.Warehouse, resources); err != nil {
		return err
	}

	if err := tx.createResourceTransportRecords(req, commonsCtx.UserID(ctx), resources); err != nil {
		return err
	}
	return tx.Commit()
}

func (tx *txDataManager) updateWarehouseStocks(newWarehouse mcom.Warehouse, resources []models.MaterialResource) error {
	type variationKey struct {
		warehouseID       string
		warehouseLocation string
		productID         string
	}
	warehouseStockVariation := make(map[variationKey]decimal.Decimal)
	for _, resource := range resources {
		// reduce.
		oldWarehouseKey := variationKey{
			warehouseID:       resource.WarehouseID,
			warehouseLocation: resource.WarehouseLocation,
			productID:         resource.ProductID,
		}
		val, ok := warehouseStockVariation[oldWarehouseKey]
		if !ok {
			val = decimal.Zero
		}
		variation := resource.Quantity
		warehouseStockVariation[oldWarehouseKey] = val.Sub(variation)

		// add.
		newWarehouseKey := variationKey{
			warehouseID:       newWarehouse.ID,
			warehouseLocation: newWarehouse.Location,
			productID:         resource.ID,
		}
		val, ok = warehouseStockVariation[newWarehouseKey]
		if !ok {
			val = decimal.Zero
		}
		warehouseStockVariation[newWarehouseKey] = val.Add(variation)
	}
	// upsert.
	upsertStockData := make([]models.WarehouseStock, len(warehouseStockVariation))
	index := 0
	for key, val := range warehouseStockVariation {
		upsertStockData[index] = models.WarehouseStock{
			ID:        key.warehouseID,
			Location:  key.warehouseLocation,
			ProductID: key.productID,
			Quantity:  val,
		}
		index++
	}

	oids := make([]string, len(resources))
	for i := 0; i < len(resources); i++ {
		oids[i] = resources[i].OID
	}

	if err := tx.db.Model(&models.MaterialResource{}).Where(`oid in ?`, oids).
		Updates(map[string]interface{}{"warehouse_id": newWarehouse.ID, "warehouse_location": newWarehouse.Location}).Error; err != nil {
		return err
	}

	return tx.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
			{Name: "location"},
			{Name: "product_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"quantity": gorm.Expr("warehouse_stock.quantity+EXCLUDED.quantity"),
		}),
	}).Create(&upsertStockData).Error
}

func (tx *txDataManager) createResourceTransportRecords(
	req mcom.WarehousingStockRequest,
	createdBy string,
	resources []models.MaterialResource) error {
	rows := make([]models.ResourceTransportRecord, len(resources))
	for i, resource := range resources {
		rows[i] = models.ResourceTransportRecord{
			OldWarehouseID: resource.WarehouseID,
			OldLocation:    resource.WarehouseLocation,
			NewWarehouseID: req.Warehouse.ID,
			NewLocation:    req.Warehouse.Location,
			ResourceID:     resource.ID,
			Note:           "",
			CreatedBy:      createdBy,
		}
	}
	return tx.db.Create(&rows).Error
}

func (session *session) listResources(resourceIDs []string) ([]models.MaterialResource, error) {
	var res []models.MaterialResource
	err := session.db.Model(&models.MaterialResource{}).Where(`id in ?`, resourceIDs).Find(&res).Error
	if err != nil {
		return res, err
	}

	// all resources in the request should be found in material_resource table.
	return res, matchResources(res, resourceIDs)
}

func matchResources(res []models.MaterialResource, resourceIDs []string) error {
	distinctMap := make(map[string]bool)
	for _, resource := range res {
		distinctMap[resource.ID] = true
	}
	if len(distinctMap) != len(resourceIDs) {
		return mcomErr.Error{Code: mcomErr.Code_RESOURCE_NOT_FOUND}
	}
	return nil
}

// ListMaterialResourceStatus implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListMaterialResourceStatus(context.Context) (mcom.ListMaterialResourceStatusReply, error) {
	types := make([]string, len(resources.MaterialStatus_name)-1)
	for i := 1; ; i++ {
		k, ok := resources.MaterialStatus_name[int32(i)]
		if !ok {
			break
		}
		types[i-1] = k
	}
	sort.Slice(types, func(i, j int) bool {
		return types[i] < types[j]
	})
	return types, nil
}

// SplitMaterialResource implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) SplitMaterialResource(ctx context.Context, req mcom.SplitMaterialResourceRequest) (mcom.SplitMaterialResourceReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.SplitMaterialResourceReply{}, err
	}

	tx := dm.beginTx(ctx)
	defer tx.Rollback() // nolint: errcheck

	var sourceResource models.MaterialResource
	if err := tx.db.Where(&models.MaterialResource{
		ID:          req.ResourceID,
		ProductType: req.ProductType,
	}).Clauses(clause.Locking{
		Strength: "UPDATE",
		Options:  "NOWAIT",
	}).Take(&sourceResource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.SplitMaterialResourceReply{}, mcomErr.Error{Code: mcomErr.Code_RESOURCE_NOT_FOUND}
		}
		return mcom.SplitMaterialResourceReply{}, err
	}
	user := commonsCtx.UserID(ctx)

	// region InspectionRemarks.
	oldInspections := sourceResource.Info.Inspections
	newInspections, err := oldInspections.Split(req.InspectionIDs)
	if err != nil {
		return mcom.SplitMaterialResourceReply{}, err
	}
	// endregion InspectionRemarks.

	// region quantity.
	if req.Quantity.GreaterThanOrEqual(sourceResource.Quantity) || !req.Quantity.IsPositive() {
		return mcom.SplitMaterialResourceReply{}, mcomErr.Error{
			Code: mcomErr.Code_INVALID_NUMBER,
		}
	}

	newQuantity := req.Quantity
	oldQuantity := sourceResource.Quantity.Sub(newQuantity)
	// endregion quantity.

	info := sourceResource.Info
	info.Inspections = oldInspections
	if err := tx.db.Where(&models.MaterialResource{
		ID:          req.ResourceID,
		ProductType: req.ProductType,
	}).Updates(&models.MaterialResource{
		Info:      info,
		Quantity:  oldQuantity,
		UpdatedBy: user,
	}).
		Error; err != nil {
		return mcom.SplitMaterialResourceReply{}, err
	}

	resourceID := resourceGenerator()

	info.MinDosage = decimal.Zero
	info.Inspections = newInspections
	info.Remark = req.Remark
	if err := tx.db.Create(&models.MaterialResource{
		ID:                resourceID,
		ProductID:         sourceResource.ProductID,
		ProductType:       sourceResource.ProductType,
		Quantity:          newQuantity,
		Status:            sourceResource.Status,
		ExpiryTime:        sourceResource.ExpiryTime,
		Info:              info,
		WarehouseID:       sourceResource.WarehouseID,
		WarehouseLocation: sourceResource.WarehouseLocation,
		Station:           sourceResource.Station,
		UpdatedBy:         user,
		CreatedBy:         user,
		FeedRecordsID:     sourceResource.FeedRecordsID,
	}).Error; err != nil {
		return mcom.SplitMaterialResourceReply{}, err
	}

	if err := tx.Commit(); err != nil {
		return mcom.SplitMaterialResourceReply{}, err
	}

	return mcom.SplitMaterialResourceReply{
		NewResourceID: resourceID,
	}, nil
}
