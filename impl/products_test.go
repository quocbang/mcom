package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/recipes"
)

const (
	testRecipeProductIDA = "product_A"
	testRecipeProductIDB = "product_B"
	testRecipeProductIDC = "product_C"

	testRecipeProductTypeA = "type_A"
	testRecipeProductTypeB = "type_B"
	testRecipeProductTypeC = "type_C"
)

func createMockRecipeProduetTypes(db *gorm.DB) error {
	if err := db.Create(&[]models.RecipeProcessDefinition{
		{
			OID:     "1995d6b5-b2e6-4a33-b2d3-34c2157a57a4",
			Name:    "123",
			Type:    testRecipeProductTypeA,
			Configs: []models.RecipeProcessConfig{},
			OutputProduct: models.OutputProduct{
				ID:   "123",
				Type: testRecipeProductTypeA,
			},
		},
		{
			OID:     "dcebf80c-d4cc-4d47-86a0-50fd92e73899",
			Name:    "456",
			Type:    testRecipeProductTypeB,
			Configs: []models.RecipeProcessConfig{},
			OutputProduct: models.OutputProduct{
				ID:   "456",
				Type: testRecipeProductTypeB,
			},
		},
		{
			OID:     "6d79a74d-7c09-44ae-908e-7f7fadb68484",
			Name:    "789",
			Type:    testRecipeProductTypeC,
			Configs: []models.RecipeProcessConfig{},
			OutputProduct: models.OutputProduct{
				ID:   "789",
				Type: testRecipeProductTypeC,
			},
		},
	}).Error; err != nil {
		return err
	}
	return db.Create([]*models.Recipe{{
		ID:          testRecipeProductIDA,
		ProductType: testRecipeProductTypeA,
		ProductID:   testRecipeProductIDA,
		Major:       "1",
		Minor:       "2",
		Stage:       recipes.StagePriority_TRIAL_PRODUCTION,
		Processes:   []models.RecipeProcess{},
	}, {
		ID:          testRecipeProductIDB,
		ProductType: testRecipeProductTypeB,
		ProductID:   testRecipeProductIDB,
		Major:       "1",
		Minor:       "2",
		Stage:       recipes.StagePriority_TRIAL_PRODUCTION,
		Processes:   []models.RecipeProcess{},
	}, {
		ID:          testRecipeProductIDC,
		ProductType: testRecipeProductTypeC,
		ProductID:   testRecipeProductIDC,
		Major:       "1",
		Minor:       "2",
		Stage:       recipes.StagePriority_TRIAL_PRODUCTION,
		Processes:   []models.RecipeProcess{},
	}}).Error
}

func deleteMockRecipeProductTypes(db *gorm.DB) error {
	if err := db.Where(` name IN ? `, []string{"123", "456", "789"}).
		Delete(&models.RecipeProcessDefinition{}).Error; err != nil {
		return err
	}
	return db.Where(` id IN ? `,
		[]string{testRecipeProductIDA, testRecipeProductIDB, testRecipeProductIDC}).
		Delete(&models.Recipe{}).Error
}

func createMockProductIDs(db *gorm.DB) error {
	return db.Create([]*models.RecipeProcessDefinition{
		{
			OID:     testProcessOID,
			Name:    testProcessName,
			Type:    testProcessType,
			Configs: []models.RecipeProcessConfig{},
			OutputProduct: models.OutputProduct{
				ID:   testRecipeProductIDA,
				Type: testProductType,
			},
		},
		{
			OID:     testProcessInspectionOID,
			Name:    testProcessInspectionName,
			Type:    testProcessType,
			Configs: []models.RecipeProcessConfig{},
			OutputProduct: models.OutputProduct{
				ID:   testRecipeProductIDA,
				Type: testProductType,
			},
		},
		{
			OID:     testProcessFlowOID,
			Name:    testProcessFlowName,
			Type:    testProcessType,
			Configs: []models.RecipeProcessConfig{},
			OutputProduct: models.OutputProduct{
				ID:   testRecipeProductIDB,
				Type: testProductType,
			},
		},
	}).Error
}

func deleteMockProductIDs(db *gorm.DB) error {
	return db.Where(` name IN ?`, []string{testProcessName, testProcessInspectionName, testProcessFlowName}).Delete(&models.RecipeProcessDefinition{}).Error
}

func Test_Product(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}
	defer dm.Close()

	assert.NoError(deleteMockRecipeProductTypes(db))
	assert.NoError(deleteMockProductIDs(db))

	err = createMockRecipeProduetTypes(db)
	if !assert.NoError(err) {
		return
	}

	err = createMockProductIDs(db)
	if !assert.NoError(err) {
		return
	}

	{ // ListProductTypes: empty DepartmentOID is allowed.
		result, err := dm.ListProductTypes(ctx, mcom.ListProductTypesRequest{})
		assert.NoError(err)
		expected := mcom.ListProductTypesReply{
			testRecipeProductTypeA,
			testRecipeProductTypeB,
			testRecipeProductTypeC,
			"RUBBER",
		}

		assert.NoError(err)
		assert.ElementsMatch(expected, result)
		{ // ListProductTypes: empty DepartmentOID is allowed.
			_, err := dm.ListProductTypes(ctx, mcom.ListProductTypesRequest{})
			assert.NoError(err)
		}
		{ // ListProductTypes: good case.
			result, err := dm.ListProductTypes(ctx, mcom.ListProductTypesRequest{DepartmentOID: testDepartmentID})

			expected := mcom.ListProductTypesReply{
				testRecipeProductTypeA,
				testRecipeProductTypeB,
				testRecipeProductTypeC,
				"RUBBER",
			}

			assert.NoError(err)
			assert.ElementsMatch(expected, result)
		}
		{ // ListProductIDs: insufficient request.
			_, err := dm.ListProductIDs(ctx, mcom.ListProductIDsRequest{})
			assert.ErrorIs(mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			}, err)
		}
		{ // ListProductIDs: good case.
			result, err := dm.ListProductIDs(ctx, mcom.ListProductIDsRequest{
				Type: testProductType,
			})

			expected := mcom.ListProductIDsReply{
				testRecipeProductIDA,
				testRecipeProductIDB,
			}

			assert.NoError(err)
			assert.ElementsMatch(expected, result)
		}

		assert.NoError(deleteMockRecipeProductTypes(db))
		assert.NoError(deleteMockProductIDs(db))
	}
}

func createMockProductGroup(db *gorm.DB) error {
	return db.Create([]*models.ProductGroup{{
		ProductID:    testRecipeProductIDA,
		ProductType:  testRecipeProductTypeA,
		DepartmentID: testDepartmentID,
		Children:     []string{testRecipeProductIDB, testRecipeProductIDC},
	}}).Error
}

func deleteMockProductGroup(db *gorm.DB) error {
	return db.Where(` product_id = ? `, testRecipeProductIDA).Delete(&models.ProductGroup{}).Error
}

func Test_ListProductGroups(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	dm, db, err := newTestDataManager()
	if !assert.NoError(err) {
		return
	}
	if !assert.NotNil(dm) {
		return
	}
	defer dm.Close()

	assert.NoError(deleteMockProductGroup(db))

	err = createMockProductGroup(db)
	if !assert.NoError(err) {
		return
	}

	{ // ListProductGroups: insufficient request.
		_, err := dm.ListProductGroups(ctx, mcom.ListProductGroupsRequest{})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // ListProductGroups: good case.
		result, err := dm.ListProductGroups(ctx, mcom.ListProductGroupsRequest{
			DepartmentOID: testDepartmentID,
			Type:          testRecipeProductTypeA,
		})
		assert.NoError(err)
		assert.Equal(mcom.ListProductGroupsReply{
			Products: []mcom.ProductGroup{{
				ID:       testRecipeProductIDA,
				Children: []string{testRecipeProductIDB, testRecipeProductIDC},
			}},
		}, result)
	}
	assert.NoError(deleteMockProductGroup(db))
}
