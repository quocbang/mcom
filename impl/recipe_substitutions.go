package impl

import (
	"context"
	"fmt"
	"sort"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

// ListSubstitutions implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListSubstitutions(ctx context.Context, req mcom.ListSubstitutionsRequest) (mcom.ListSubstitutionsReply, error) {
	if err := checkInsufficientRequest(req.ProductID.ID, req.ProductID.Grade); err != nil {
		return mcom.ListSubstitutionsReply{}, err
	}
	rep, err := req.ProductID.ListSubstitutions(dm.db)
	return mcom.ListSubstitutionsReply(rep), err
}

// ListMultipleSubstitutions implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListMultipleSubstitutions(ctx context.Context, req mcom.ListMultipleSubstitutionsRequest) (mcom.ListMultipleSubstitutionsReply, error) {
	var substitutionMappings []models.SubstitutionMapping
	whereClause := make([][]string, len(req.ProductIDs))
	for i, productID := range req.ProductIDs {
		whereClause[i] = []string{productID.ID, productID.Grade}
	}
	if err := dm.db.WithContext(ctx).Where(`(id, grade) IN ?`, whereClause).Find(&substitutionMappings).Error; err != nil {
		return mcom.ListMultipleSubstitutionsReply{}, err
	}

	productSubstitutionMap := make(map[models.ProductID]mcom.ListSubstitutionsReply)
	for _, subMap := range substitutionMappings {
		id := models.ProductID{ID: subMap.ID, Grade: subMap.Grade}
		if _, ok := productSubstitutionMap[id]; ok {
			return mcom.ListMultipleSubstitutionsReply{}, fmt.Errorf("violation of pk constraint in product substitution")
		}

		sortedVal := subMap.Substitutions
		sort.Slice(sortedVal, func(i, j int) bool {
			if sortedVal[i].ID == sortedVal[j].ID {
				return sortedVal[i].Grade < sortedVal[j].Grade
			}
			return sortedVal[i].ID < sortedVal[j].ID
		})

		productSubstitutionMap[id] = mcom.ListSubstitutionsReply{
			Substitutions: sortedVal,
			UpdatedAt:     subMap.UpdatedAt,
			UpdatedBy:     subMap.UpdatedBy,
		}
	}

	return mcom.ListMultipleSubstitutionsReply{
		Reply: productSubstitutionMap,
	}, nil
}

// DeleteSubstitutions implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) DeleteSubstitutions(ctx context.Context, req mcom.DeleteSubstitutionsRequest) error {
	if err := checkInsufficientRequest(req.ProductID.ID, req.ProductID.Grade); err != nil {
		return err
	}
	if req.DeleteAll {
		return req.ProductID.ClearSubstitutions(ctx, dm.db)
	}
	return req.ProductID.RemoveSubstitutions(ctx, dm.db, req.Contents.ToSubstitutions())
}

// UpdateSubstitutions implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) UpdateSubstitutions(ctx context.Context, req mcom.BasicSubstitutionRequest) error {
	if err := checkInsufficientRequest(req.ProductID.ID, req.ProductID.Grade); err != nil {
		return err
	}
	if len(req.Contents) == 0 {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "updating with empty contens is not allowed"}
	}
	return req.ProductID.SetSubstitutions(ctx, dm.db, req.Contents)
}

// AddSubstitutions implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) AddSubstitutions(ctx context.Context, req mcom.BasicSubstitutionRequest) error {
	if err := checkInsufficientRequest(req.ProductID.ID, req.ProductID.Grade); err != nil {
		return err
	}
	if len(req.Contents) == 0 {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "adding with empty contents is not allowed"}
	}
	return req.ProductID.AddSubstitutions(ctx, dm.db, req.Contents)
}

func checkInsufficientRequest(id, grade string) error {
	if id == "" {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "ProductID is required",
		}
	}
	return nil
}
