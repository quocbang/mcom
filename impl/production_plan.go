package impl

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	pbWorkOrder "gitlab.kenda.com.tw/kenda/commons/v2/proto/golang/mes/v2/workorder"
	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/workorder"
)

func (session *session) createProductPlan(
	ctx context.Context,
	departmentID,
	productID string,
	productType string,
	date time.Time,
	dayQuantity decimal.Decimal,
	userID string,
) error {
	if err := session.db.Create(&models.ProductionPlan{
		DepartmentID: departmentID,
		PlanDate:     date,
		ProductionPlanProduct: models.ProductionPlanProduct{
			ID:   productID,
			Type: productType,
		},
		Quantity:  dayQuantity,
		UpdatedAt: 0,
		UpdatedBy: userID,
		CreatedAt: 0,
		CreatedBy: userID,
	}).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return mcomErr.Error{
				Code: mcomErr.Code_PRODUCTION_PLAN_EXISTED,
			}
		}
		return err
	}
	return nil
}

// CreateProductPlan implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateProductPlan(ctx context.Context, req mcom.CreateProductionPlanRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	session := dm.newSession(ctx)
	// get department data.
	department, err := session.getDepartmentInfo(req.DepartmentOID)
	if err != nil {
		return err
	}

	planDate, err := parseDate(req.Date)
	if err != nil {
		return err
	}

	// create production plan data.
	return session.createProductPlan(
		ctx,
		department.ID,
		req.Product.ID,
		req.Product.Type,
		planDate,
		req.Quantity,
		commonsCtx.UserID(ctx),
	)
}

func (session *session) getProductPlans(departmentID string, date time.Time, productType string) ([]models.ProductionPlan, error) {
	plans := []models.ProductionPlan{}
	if err := session.db.Where(` department_id = ? AND plan_date = ? AND product_type = ? `, departmentID, date, productType).
		Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

type getProductPlanWeekQuantityRequest struct {
	DepartmentID string
	Product      models.ProductionPlanProduct
	StartDate    time.Time
	EndDate      time.Time
}

func (session *session) getProductPlanWeekQuantity(req *getProductPlanWeekQuantityRequest) (decimal.Decimal, error) {
	var results []models.ProductionPlan
	if err := session.db.Where(` department_id = ? AND product_id = ? AND product_type = ? AND plan_date BETWEEN ? AND ? `,
		req.DepartmentID, req.Product.ID, req.Product.Type, req.StartDate, req.EndDate).
		Find(&results).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return decimal.Zero, err
	}

	weekQuantity := decimal.Zero
	for _, r := range results {
		quantity := r.Quantity
		weekQuantity = weekQuantity.Add(quantity)
	}
	return weekQuantity, nil
}

func (session *session) getProductPlanQuantity(departmentID string, product models.ProductionPlanProduct, planDate time.Time) (decimal.Decimal, error) {
	startDate, endDate := getWeekFirstAndLastDate(planDate)
	productPlanWeekly, err := session.getProductPlanWeekQuantity(&getProductPlanWeekQuantityRequest{
		DepartmentID: departmentID,
		Product:      product,
		StartDate:    startDate,
		EndDate:      endDate,
	})
	if err != nil {
		return decimal.Zero, err
	}

	return productPlanWeekly, nil
}

// TODO: 暫時沒有庫存資料, 後期補上實作
func (session *session) getProductStock() (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (session *session) getReservedQuantity(departmentID string, product models.ProductionPlanProduct, date time.Time) (decimal.Decimal, error) {
	processes := []*models.RecipeProcessDefinition{}
	if err := session.db.Where(` product_id = ? AND product_type = ? `, product.ID, product.Type).Find(&processes).Error; err != nil {
		return decimal.Zero, err
	}

	planDate, err := parseDate(date)
	if err != nil {
		return decimal.Zero, err
	}

	resvQty := decimal.Zero
	for _, proc := range processes {
		workOrders := []models.WorkOrder{}
		if err := session.db.
			Where(` department_id = ? AND recipe_id = ? AND process_name = ? AND process_type = ? AND reserved_date = ? AND status != ? `,
				departmentID, proc.RecipeID, proc.Name, proc.Type, planDate, pbWorkOrder.Status_SKIPPED).
			Find(&workOrders).Error; err != nil {
			return decimal.Zero, err
		}

		for _, w := range workOrders {
			switch w.Information.BatchQuantityType {
			case workorder.BatchSize_PER_BATCH_QUANTITIES:
				for _, qty := range w.Information.QuantityForBatches {
					resvQty = resvQty.Add(qty)
				}
			case workorder.BatchSize_FIXED_QUANTITY:
				resvQty = resvQty.Add(w.Information.FixedQuantity.PlanQuantity)
			case workorder.BatchSize_PLAN_QUANTITY:
				resvQty = resvQty.Add(w.Information.PlanQuantity.PlanQuantity)
			}
		}
	}
	return resvQty, nil
}

func (session *session) parseProductPlans(plans []models.ProductionPlan) (mcom.ListProductPlansReply, error) {
	productPlans := make([]mcom.ProductPlan, len(plans))
	for i, v := range plans {
		productPlanWeekly, err := session.getProductPlanQuantity(v.DepartmentID, v.ProductionPlanProduct, v.PlanDate)
		if err != nil {
			return mcom.ListProductPlansReply{}, err
		}

		stock, err := session.getProductStock()
		if err != nil {
			return mcom.ListProductPlansReply{}, err
		}

		reservedQuantity, err := session.getReservedQuantity(v.DepartmentID, v.ProductionPlanProduct, v.PlanDate)
		if err != nil {
			return mcom.ListProductPlansReply{}, err
		}

		productPlans[i] = mcom.ProductPlan{
			ProductID: v.ProductionPlanProduct.ID,
			Quantity: mcom.ProductPlanQuantity{
				Daily:    v.Quantity,
				Week:     productPlanWeekly,
				Stock:    stock,
				Reserved: reservedQuantity,
			},
		}
	}

	// sort by product id.
	sort.Slice(productPlans, func(i, j int) bool {
		return productPlans[i].ProductID < productPlans[j].ProductID
	})

	return mcom.ListProductPlansReply{
		ProductPlans: productPlans,
	}, nil
}

// ListProductPlans implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListProductPlans(ctx context.Context, req mcom.ListProductPlansRequest) (mcom.ListProductPlansReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListProductPlansReply{}, err
	}

	planDate, err := parseDate(req.Date)
	if err != nil {
		return mcom.ListProductPlansReply{}, err
	}

	session := dm.newSession(ctx)
	plans, err := session.getProductPlans(req.DepartmentOID, planDate, req.ProductType)
	if err != nil {
		return mcom.ListProductPlansReply{}, err
	}

	return session.parseProductPlans(plans)
}
