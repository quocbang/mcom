package impl

import (
	"context"
	"sort"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

// ListProductTypes implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListProductTypes(ctx context.Context, req mcom.ListProductTypesRequest) (mcom.ListProductTypesReply, error) {
	// todo : 現階段並沒有進行部門篩選，一律對全部的RecipeProcessDefinition進行查找。
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListProductTypesReply{}, err
	}

	types := []string{}
	session := dm.newSession(ctx)
	if err := session.db.Model(&models.RecipeProcessDefinition{}).
		Select(`product_type`).
		Group(`product_type`).
		Find(&types).Error; err != nil {
		return nil, err
	}

	sort.Slice(types, func(i, j int) bool {
		return types[i] < types[j]
	})

	return types, nil
}

// ListProductIDs implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListProductIDs(ctx context.Context, req mcom.ListProductIDsRequest) (mcom.ListProductIDsReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListProductIDsReply{}, err
	}

	session := dm.newSession(ctx)
	productIDs := mcom.ListProductIDsReply{}
	if req.IsLastProcess {
		recipes := []models.Recipe{}
		if err := session.db.Where(` product_type = ? `, req.Type).Find(&recipes).Error; err != nil {
			return nil, err
		}
		for _, v := range recipes {
			productIDs = append(productIDs, v.ProductID)
		}
	} else {
		processes := []models.RecipeProcessDefinition{}
		if err := session.db.Distinct("product_id").Where(` product_type = ? `, req.Type).Find(&processes).Error; err != nil {
			return nil, err
		}
		for _, v := range processes {
			productIDs = append(productIDs, v.OutputProduct.ID)
		}
	}

	sort.Slice(productIDs, func(i, j int) bool {
		return productIDs[i] < productIDs[j]
	})

	return productIDs, nil
}

func parseProductGroup(productGroup []models.ProductGroup) mcom.ListProductGroupsReply {
	products := make([]mcom.ProductGroup, len(productGroup))
	for i, v := range productGroup {
		products[i] = mcom.ProductGroup{
			ID:       v.ProductID,
			Children: mcom.ListProductIDsReply(v.Children),
		}
	}

	// sort by ID.
	sort.Slice(products, func(i, j int) bool {
		return products[i].ID < products[j].ID
	})

	return mcom.ListProductGroupsReply{
		Products: products,
	}
}

// ListProductGroups implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListProductGroups(ctx context.Context, req mcom.ListProductGroupsRequest) (mcom.ListProductGroupsReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListProductGroupsReply{}, err
	}

	session := dm.newSession(ctx)
	productGroup := []models.ProductGroup{}
	if err := session.db.Where(` department_id = ? AND product_type = ? `, req.DepartmentOID, req.Type).
		Find(&productGroup).Error; err != nil {
		return mcom.ListProductGroupsReply{}, err
	}
	return parseProductGroup(productGroup), nil
}
