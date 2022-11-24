package mcom

import (
	"github.com/shopspring/decimal"

	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/workorder"
)

type BatchQuantity interface {
	Detail() models.BatchQuantityDetails
}

func NewQuantityPerBatch(batchesQuantity []decimal.Decimal) quantityPerBatch {
	return quantityPerBatch{
		BatchQuantityType:  workorder.BatchSize_PER_BATCH_QUANTITIES,
		QuantityForBatches: batchesQuantity,
	}
}

type quantityPerBatch models.BatchQuantityDetails

func (q quantityPerBatch) Detail() models.BatchQuantityDetails {
	return models.BatchQuantityDetails(q)
}

func NewFixedQuantity(batchCount uint, planQuantity decimal.Decimal) fixedQuantity {
	return fixedQuantity{
		BatchQuantityType: workorder.BatchSize_FIXED_QUANTITY,
		FixedQuantity: &models.FixedQuantity{
			BatchCount:   batchCount,
			PlanQuantity: planQuantity,
		},
	}
}

type fixedQuantity models.BatchQuantityDetails

func (q fixedQuantity) Detail() models.BatchQuantityDetails {
	return models.BatchQuantityDetails(q)
}

func NewPlanQuantity(batchCount uint, quantity decimal.Decimal) planQuantity {
	return planQuantity{
		BatchQuantityType: workorder.BatchSize_PLAN_QUANTITY,
		PlanQuantity: &models.PlanQuantity{
			BatchCount:   batchCount,
			PlanQuantity: quantity,
		},
	}
}

type planQuantity models.BatchQuantityDetails

func (q planQuantity) Detail() models.BatchQuantityDetails {
	return models.BatchQuantityDetails(q)
}
