package impl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func (session *session) getMaterialResourceInfo(condition *models.MaterialResource) (models.MaterialResource, error) {
	var results models.MaterialResource
	err := session.db.
		Where(condition).
		Take(&results).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.MaterialResource{}, mcomErr.Error{
				Code: mcomErr.Code_RESOURCE_NOT_FOUND,
			}
		}
		return models.MaterialResource{}, err
	}
	return results, nil
}

// CreateMaterialResources implement gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateMaterialResources(ctx context.Context, req mcom.CreateMaterialResourcesRequest, opts ...mcom.CreateMaterialResourcesOption) (reply mcom.CreateMaterialResourcesReply, err error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.CreateMaterialResourcesReply{}, err
	}

	userID := commonsCtx.UserID(ctx)
	o := mcom.ParseCreateMaterialResourcesOptions(opts)

	if len(opts) != 0 && (o.Warehouse.ID == "" || o.Warehouse.Location == "") {
		return mcom.CreateMaterialResourcesReply{}, mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing warehouse id or warehouse location",
		}
	}

	resources := parseMaterialResourceInsertData(req.Materials, userID, o.Warehouse)
	session := dm.newSession(ctx)
	reply, err = session.createMaterialResourcesAndStockIn(ctx, resources, o.Warehouse)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// resourceGenerator TODO: generate barcode id rules.
func resourceGenerator() string {
	return strings.ToUpper(xid.New().String())
}

func parseMaterialResources(resources []*models.MaterialResource, oidGetter materialResourceOIDGetter) ([]mcom.CreatedMaterialResource, error) {
	res := make([]mcom.CreatedMaterialResource, len(resources))
	for i, r := range resources {
		oid, err := oidGetter(r.ID, r.ProductType)
		if err != nil {
			return nil, err
		}

		res[i] = mcom.CreatedMaterialResource{
			ID:  r.ID,
			OID: oid,
		}
	}
	return res, nil
}

func parseMaterialResourceInsertData(resources []mcom.CreateMaterialResourcesRequestDetail, userID string, stock mcom.Warehouse) []*models.MaterialResource {
	materials := make([]*models.MaterialResource, len(resources))
	for i, res := range resources {
		resourceID := res.ResourceID
		if resourceID == "" {
			resourceID = resourceGenerator()
		}
		feedRecs := []string{}
		if res.FeedRecordID != "" {
			feedRecs = append(feedRecs, res.FeedRecordID)
		}
		materials[i] = &models.MaterialResource{
			ID:          resourceID,
			ProductID:   res.ID,
			ProductType: res.Type,
			Status:      res.Status,
			Quantity:    res.Quantity,
			ExpiryTime:  types.ToTimeNano(res.ExpiryTime),
			Info: models.MaterialInfo{
				Grade:           res.Grade,
				Unit:            res.Unit,
				LotNumber:       res.LotNumber,
				ProductionTime:  res.ProductionTime,
				MinDosage:       res.MinDosage,
				Inspections:     res.Inspections,
				PlannedQuantity: res.PlannedQuantity,
				Remark:          res.Remark,
			},
			Station:           res.Station,
			UpdatedBy:         userID,
			CreatedBy:         userID,
			WarehouseID:       stock.ID,
			WarehouseLocation: stock.Location,
			FeedRecordsID:     feedRecs,
		}
	}
	return materials
}

func (session *session) createMaterialResourcesAndStockIn(
	ctx context.Context,
	resources []*models.MaterialResource,
	warehouse mcom.Warehouse) ([]mcom.CreatedMaterialResource, error) {
	tx := session.beginTx()
	defer tx.db.Rollback()

	// create resource.
	row := tx.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
			{Name: "product_type"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			// NOTE: we have to set the space at the end for the distinction
			// between the update clause and the subquery clause.
			"quantity":        gorm.Expr("material_resource.quantity+EXCLUDED.quantity "),
			"feed_records_id": gorm.Expr("material_resource.feed_records_id||EXCLUDED.feed_records_id"),
		}),
		Where: clause.Where{
			Exprs: []clause.Expression{
				clause.Expr{SQL: "material_resource.product_id=EXCLUDED.product_id"},
			},
		},
	}).Create(resources)
	if err := row.Error; err != nil {
		return nil, err
	}
	if row.RowsAffected != int64(len(resources)) {
		detail := ""
		for _, resource := range resources {
			detail += resource.Describe()
		}
		return nil, mcomErr.Error{
			Code:    mcomErr.Code_RESOURCE_EXISTED,
			Details: fmt.Sprintf("resources to create: \n%s", detail),
		}
	}

	// stock in.
	if needStockIn(warehouse) {
		if err := tx.stockIn(resources, warehouse); err != nil {
			return nil, err
		}
	}

	oidGetter, err := tx.newMaterialResourceOIDGetter(resources)
	if err != nil {
		return nil, err
	}
	replies, err := parseMaterialResources(resources, oidGetter)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return replies, nil
}

type materialResourceOIDGetter func(resourceID, productType string) (oid string, err error)

func (tx *txDataManager) newMaterialResourceOIDGetter(rs []*models.MaterialResource) (materialResourceOIDGetter, error) {
	type key struct {
		ResourceID  string
		ProductType string
	}

	whereClause := make([][]string, len(rs))
	for i, r := range rs {
		whereClause[i] = []string{r.ID, r.ProductType}
	}

	var resources []*models.MaterialResource
	if err := tx.db.Where("(id, product_type) IN ?", whereClause).Find(&resources).Error; err != nil {
		return nil, err
	}

	m := make(map[key]string, len(rs))
	for _, r := range resources {
		k := key{
			ResourceID:  r.ID,
			ProductType: r.ProductType,
		}
		m[k] = r.OID
	}
	return func(resourceID, productType string) (string, error) {
		k := key{
			ResourceID:  resourceID,
			ProductType: productType,
		}
		oid, ok := m[k]
		if !ok {
			return "", fmt.Errorf("the resource oid is not found: resourceID=%s, productType=%s", resourceID, productType)
		}
		return oid, nil
	}, nil
}

func needStockIn(warehouse mcom.Warehouse) bool {
	return warehouse.ID != "" && warehouse.Location != ""
}

func (tx *txDataManager) stockIn(
	resources []*models.MaterialResource,
	warehouse mcom.Warehouse,
) error {
	type id struct {
		ID        string
		Location  string
		ProductID string
	}

	preResourceStockIn := make([]*models.WarehouseStock, len(resources))
	for i, resource := range resources {
		quantity := resource.Quantity
		preResourceStockIn[i] = &models.WarehouseStock{
			ID:        warehouse.ID,
			Location:  warehouse.Location,
			ProductID: resource.ProductID,
			Quantity:  quantity,
		}
	}

	idQuantityPair := make(map[id]decimal.Decimal)
	for _, stock := range preResourceStockIn {
		id := id{
			ID:        stock.ID,
			Location:  stock.Location,
			ProductID: stock.ProductID,
		}
		val, ok := idQuantityPair[id]
		if ok {
			idQuantityPair[id] = val.Add(stock.Quantity)
		} else {
			idQuantityPair[id] = stock.Quantity
		}
	}

	resourceStockIn := []models.WarehouseStock{}

	for k, v := range idQuantityPair {
		resourceStockIn = append(resourceStockIn, models.WarehouseStock{
			ID:        k.ID,
			Location:  k.Location,
			ProductID: k.ProductID,
			Quantity:  v,
		})
	}

	if err := tx.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
			{Name: "location"},
			{Name: "product_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"quantity": gorm.Expr("warehouse_stock.quantity+EXCLUDED.quantity"),
		}),
	}).Create(resourceStockIn).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			detail := ""
			for _, stock := range resourceStockIn {
				detail += stock.Describe()
			}
			return mcomErr.Error{
				Code:    mcomErr.Code_RESOURCE_EXISTED,
				Details: fmt.Sprintf("resources stock in to create:\n%s", detail),
			}
		}
		return err
	}
	return nil
}
