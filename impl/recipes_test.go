package impl

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/recipes"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

const (
	testRecipeID         = "recipe_id"
	testRecipeIDNotExist = "bad_recipe_id"

	testProductID         = "product_id"
	testProductType       = "RUBBER"
	testProductIDNotExist = "bad_product_id"

	testProcessType           = "test"
	testProcessOID            = "3ca44172-8bc9-46c9-b4d6-1806dc97115c"
	testProcessFlowOID        = "4db44172-8bc9-46c9-b4d6-1806dc97115c"
	testProcessFlowOIDB       = "12344172-8bc9-46c9-b4d6-1806dc97115c"
	testProcessInspectionOID  = "5ab44172-8bc9-46c9-b4d6-1806dc97115c"
	testProcessName           = "process_name"
	testProcessFlowName       = "process_flow_name"
	testProcessFlowNameB      = "process_flow_B"
	testProcessInspectionName = "process_inspection_name"

	testRecipeToolID   = "tool_id"
	testRecipeToolType = "tool"

	testMaterialName  = "material_A"
	testMaterialGrade = "A"

	testRecipeSite             = "recipe_site"
	testRecipeRequiredRecipeID = "recipe_id"
	testRecipePropertyName     = "property_ame"

	testConfigStation = "gundam"

	testStanding = 0
	testExpiry   = 168
)

func createMockRecipeProcessDefinitionData(db *gorm.DB) error {
	if err := db.Create(&[]models.RecipeProcessDefinition{{
		OID:      testProcessOID,
		Name:     testProcessName,
		Type:     testProcessType,
		RecipeID: sql.NullString{String: testRecipeID, Valid: true},
		Configs: []models.RecipeProcessConfig{
			{
				Stations:  []string{testConfigStation},
				BatchSize: types.Decimal.RequireFromString("200"),
				Unit:      "kg",
			},
		},
		OutputProduct: models.OutputProduct{
			ID:   testProductID,
			Type: testProductType,
		},
		ProductValidPeriods: models.ProductValidPeriod{
			Standing: testStanding,
			Expiry:   testExpiry,
		},
	}, {
		OID:      testProcessFlowOID,
		Name:     testProcessFlowName,
		Type:     testProcessType,
		RecipeID: sql.NullString{String: testRecipeID, Valid: true},
		Configs: []models.RecipeProcessConfig{
			{
				Stations:  []string{testConfigStation},
				BatchSize: types.Decimal.RequireFromString("200"),
				Unit:      "kg",
			},
		},
		OutputProduct: models.OutputProduct{
			ID:   testProductID,
			Type: testProductType,
		},
		ProductValidPeriods: models.ProductValidPeriod{
			Standing: testStanding,
			Expiry:   testExpiry,
		},
	},
	}).Error; err != nil {
		return err
	}
	return nil
}

func deleteMockRecipeProcessDefinitionData(db *gorm.DB) error {
	if err := db.Where(` name = ? `, testProcessName).Delete(&models.RecipeProcessDefinition{}).Error; err != nil {
		return err
	}
	if err := db.Where(` name = ? `, testProcessFlowName).Delete(&models.RecipeProcessDefinition{}).Error; err != nil {
		return err
	}
	return nil
}

func Test_IsProductExisted(t *testing.T) {
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

	assert.NoError(deleteMockRecipeProcessDefinitionData(db))

	err = createMockRecipeProcessDefinitionData(db)
	if !assert.NoError(err) {
		return
	}

	{ // insufficient request.
		_, err := dm.IsProductExisted(ctx, "")
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // product id not found.
		_, err := dm.IsProductExisted(ctx, testProductIDNotExist)
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_PRODUCT_ID_NOT_FOUND,
		}, err)
	}
	{ // good case.
		result, err := dm.IsProductExisted(ctx, testProductID)
		assert.NoError(err)
		assert.True(result)
	}

	assert.NoError(deleteMockRecipeProcessDefinitionData(db))
}

var commonRecipeProcessesInfo = []*mcom.Process{{
	OID: testProcessOID,
	OptionalFlows: []*mcom.RecipeOptionalFlow{{
		Name:           testProcessFlowName,
		OIDs:           []string{testProcessFlowOID},
		MaxRepetitions: 0,
	}},
}}

var procDef = mcom.ProcessDefinition{
	OID:  testProcessOID,
	Name: testProcessName,
	Type: testProcessType,
	Configs: []*mcom.RecipeProcessConfig{{
		Stations:  []string{"A", "B", "C"},
		BatchSize: types.Decimal.RequireFromString("600"),
		Unit:      "kg",
		Tools: []*mcom.RecipeTool{{
			Type:     testRecipeToolID,
			ID:       testRecipeToolType,
			Required: false,
		}},
		Steps: []*mcom.RecipeProcessStep{{
			Materials: []*mcom.RecipeMaterial{{
				Name:  testMaterialName,
				Grade: testMaterialGrade,
				Value: mcom.RecipeMaterialParameter{
					High: types.Decimal.RequireFromString("100"),
					Mid:  types.Decimal.RequireFromString("50"),
					Low:  types.Decimal.RequireFromString("0"),
					Unit: "kg",
				},
				Site:             testRecipeSite,
				RequiredRecipeID: testRecipeRequiredRecipeID,
			}},
			Controls: []*mcom.RecipeProperty{{
				Name: testRecipePropertyName,
				Param: &mcom.RecipePropertyParameter{
					High: types.Decimal.RequireFromString("100"),
					Mid:  types.Decimal.RequireFromString("50"),
					Low:  types.Decimal.RequireFromString("0"),
					Unit: "kg",
				},
			}},
			Measurements: []*mcom.RecipeProperty{{
				Name: testRecipePropertyName,
				Param: &mcom.RecipePropertyParameter{
					High: types.Decimal.RequireFromString("100"),
					Mid:  types.Decimal.RequireFromString("50"),
					Low:  types.Decimal.RequireFromString("0"),
					Unit: "kg",
				},
			}},
		}},
		CommonControls: []*mcom.RecipeProperty{},
		CommonProperties: []*mcom.RecipeProperty{
			{
				Name:  "commonPropertiesTest",
				Param: &mcom.RecipePropertyParameter{},
			},
		},
	}},
	Output: mcom.OutputProduct{
		ID:   testProductID,
		Type: testProductType,
	},
	ProductValidPeriod: mcom.ProductValidPeriodConfig{
		Standing: testStanding,
		Expiry:   testExpiry,
	},
}

var procFlowDef = mcom.ProcessDefinition{
	OID:  testProcessFlowOID,
	Name: testProcessFlowName,
	Type: testProcessType,
	Configs: []*mcom.RecipeProcessConfig{{
		Stations:  []string{"A", "B", "C"},
		BatchSize: types.Decimal.RequireFromString("600"),
		Unit:      "kg",
		Tools: []*mcom.RecipeTool{{
			Type:     testRecipeToolID,
			ID:       testRecipeToolType,
			Required: false,
		}},
		Steps: []*mcom.RecipeProcessStep{{
			Materials: []*mcom.RecipeMaterial{{
				Name:  testMaterialName,
				Grade: testMaterialGrade,
				Value: mcom.RecipeMaterialParameter{
					High: types.Decimal.RequireFromString("100"),
					Mid:  types.Decimal.RequireFromString("50"),
					Low:  types.Decimal.RequireFromString("0"),
					Unit: "kg",
				},
				Site:             testRecipeSite,
				RequiredRecipeID: testRecipeRequiredRecipeID,
			}},
			Controls: []*mcom.RecipeProperty{{
				Name: testRecipePropertyName,
				Param: &mcom.RecipePropertyParameter{
					High: types.Decimal.RequireFromString("100"),
					Mid:  types.Decimal.RequireFromString("50"),
					Low:  types.Decimal.RequireFromString("0"),
					Unit: "kg",
				},
			}},
			Measurements: []*mcom.RecipeProperty{{
				Name: testRecipePropertyName,
				Param: &mcom.RecipePropertyParameter{
					High: types.Decimal.RequireFromString("100"),
					Mid:  types.Decimal.RequireFromString("50"),
					Low:  types.Decimal.RequireFromString("0"),
					Unit: "kg",
				},
			}},
		}},
		CommonControls: []*mcom.RecipeProperty{},
		CommonProperties: []*mcom.RecipeProperty{
			{
				Name:  "commonPropertiesTest",
				Param: &mcom.RecipePropertyParameter{},
			},
		},
	}},
	Output: mcom.OutputProduct{
		ID:   testProductID,
		Type: testProductType,
	},
	ProductValidPeriod: mcom.ProductValidPeriodConfig{
		Standing: testStanding,
		Expiry:   testExpiry,
	},
}

var procFlowB = mcom.ProcessDefinition{
	OID:  testProcessFlowOIDB,
	Name: testProcessFlowNameB,
	Type: testProcessType,
	Configs: []*mcom.RecipeProcessConfig{{
		Stations:  []string{"A", "B", "C"},
		BatchSize: types.Decimal.RequireFromString("600"),
		Unit:      "kg",
		Tools: []*mcom.RecipeTool{{
			Type:     testRecipeToolID,
			ID:       testRecipeToolType,
			Required: false,
		}},
		Steps: []*mcom.RecipeProcessStep{{
			Materials: []*mcom.RecipeMaterial{{
				Name:  testMaterialName,
				Grade: testMaterialGrade,
				Value: mcom.RecipeMaterialParameter{
					High: types.Decimal.RequireFromString("100"),
					Mid:  types.Decimal.RequireFromString("50"),
					Low:  types.Decimal.RequireFromString("0"),
					Unit: "kg",
				},
				Site:             testRecipeSite,
				RequiredRecipeID: testRecipeRequiredRecipeID,
			}},
			Controls: []*mcom.RecipeProperty{{
				Name: testRecipePropertyName,
				Param: &mcom.RecipePropertyParameter{
					High: types.Decimal.RequireFromString("100"),
					Mid:  types.Decimal.RequireFromString("50"),
					Low:  types.Decimal.RequireFromString("0"),
					Unit: "kg",
				},
			}},
			Measurements: []*mcom.RecipeProperty{{
				Name: testRecipePropertyName,
				Param: &mcom.RecipePropertyParameter{
					High: types.Decimal.RequireFromString("100"),
					Mid:  types.Decimal.RequireFromString("50"),
					Low:  types.Decimal.RequireFromString("0"),
					Unit: "kg",
				},
			}},
		}},
		CommonControls: []*mcom.RecipeProperty{},
	}},
	Output: mcom.OutputProduct{
		ID:   testProductID,
		Type: testProductType,
	},
	ProductValidPeriod: mcom.ProductValidPeriodConfig{
		Standing: testStanding,
		Expiry:   testExpiry,
	},
}

var commonProcessDefinition = []mcom.ProcessDefinition{
	procDef,
	procFlowDef,
}

func Test_Recipe(t *testing.T) {
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

	assert.NoError(newClearMaster(db, models.RecipeProcessDefinition{}, models.Recipe{}).Clear())
	releasedAt := types.ToTimeNano(time.Now())

	{ // CreateRecipes: nil recipes.
		err := dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
			Recipes: nil,
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "Key: 'CreateRecipesRequest.Recipes' Error:Field validation for 'Recipes' failed on the 'min' tag",
		}, err)
	}
	{ // CreateRecipes: insufficient request.
		err := dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
			Recipes: []mcom.Recipe{{
				ID: "",
				Product: mcom.Product{
					ID:   testProductID,
					Type: testProductType,
				},
				Version:            mcom.RecipeVersion{},
				Processes:          []*mcom.Process{},
				ProcessDefinitions: commonProcessDefinition,
			}},
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "Key: 'CreateRecipesRequest.Recipes[0].ID' Error:Field validation for 'ID' failed on the 'required' tag\nKey: 'CreateRecipesRequest.Recipes[0].Version.Major' Error:Field validation for 'Major' failed on the 'required' tag\nKey: 'CreateRecipesRequest.Recipes[0].Version.Stage' Error:Field validation for 'Stage' failed on the 'required' tag\nKey: 'CreateRecipesRequest.Recipes[0].Processes' Error:Field validation for 'Processes' failed on the 'min' tag",
		}, err)
	}
	{ // CreateRecipes: process not found.
		err := dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
			Recipes: []mcom.Recipe{{
				ID: testRecipeID,
				Product: mcom.Product{
					ID:   testProductID,
					Type: testProductType,
				},
				Version:            mcom.RecipeVersion{Major: "1", Minor: "2", Stage: recipes.StagePriority_UNSPECIFIED.String(), ReleasedAt: releasedAt},
				Processes:          commonRecipeProcessesInfo,
				ProcessDefinitions: []mcom.ProcessDefinition{commonProcessDefinition[1]},
			}},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_PROCESS_NOT_FOUND,
		}, err)
	}
	{ // CreateRecipes: process flow not found, missing process flow in processes.
		err := dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
			Recipes: []mcom.Recipe{{
				ID: testRecipeID,
				Product: mcom.Product{
					ID:   testProductID,
					Type: testProductType,
				},
				Version:            mcom.RecipeVersion{Major: "1", Minor: "2", Stage: "UNSPECIFIED", ReleasedAt: releasedAt},
				Processes:          commonRecipeProcessesInfo,
				ProcessDefinitions: []mcom.ProcessDefinition{commonProcessDefinition[0]},
			}},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_PROCESS_NOT_FOUND,
		}, err)
	}
	{ // CreateRecipes: good case.
		err := dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
			Recipes: []mcom.Recipe{{
				ID: testRecipeID,
				Product: mcom.Product{
					ID:   testProductID,
					Type: testProductType,
				},
				Version:            mcom.RecipeVersion{Major: "1", Minor: "2", Stage: "UNSPECIFIED", ReleasedAt: releasedAt},
				Processes:          commonRecipeProcessesInfo,
				ProcessDefinitions: commonProcessDefinition,
			}},
		})
		assert.NoError(err)
	}
	{ // CreateRecipes: recipe already exists.
		err = dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
			Recipes: []mcom.Recipe{{
				ID: testRecipeID,
				Product: mcom.Product{
					ID:   testProductID,
					Type: testProductType,
				},
				Version: mcom.RecipeVersion{Major: "1", Minor: "2", Stage: "UNSPECIFIED", ReleasedAt: releasedAt},
				Processes: []*mcom.Process{{
					OID: testProcessFlowOIDB,
				}},
				ProcessDefinitions: []mcom.ProcessDefinition{procFlowB},
			}},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_RECIPE_ALREADY_EXISTS,
		}, err)
	}
	{ // CreateRecipes: process already exists.
		err := dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
			Recipes: []mcom.Recipe{{
				ID: testRecipeID,
				Product: mcom.Product{
					ID:   testProductID,
					Type: testProductType,
				},
				Version:            mcom.RecipeVersion{Major: "1", Minor: "2", Stage: "UNSPECIFIED", ReleasedAt: releasedAt},
				Processes:          commonRecipeProcessesInfo,
				ProcessDefinitions: commonProcessDefinition,
			}},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_PROCESS_ALREADY_EXISTS,
		}, err)
	}
	{ // ListRecipesByProduct: insufficient request.
		_, err := dm.ListRecipesByProduct(ctx, mcom.ListRecipesByProductRequest{ProductID: ""})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // ListRecipesByProduct: good case.
		result, err := dm.ListRecipesByProduct(ctx, mcom.ListRecipesByProductRequest{ProductID: testProductID})
		assert.NoError(err)
		assert.Equal(mcom.ListRecipesByProductReply{
			Recipes: []mcom.GetRecipeReply{
				{
					ID: testRecipeID,
					Product: mcom.Product{
						ID:   testProductID,
						Type: testProductType,
					},
					Version: mcom.RecipeVersion{Major: "1", Minor: "2", Stage: "UNSPECIFIED", ReleasedAt: releasedAt},
					Processes: []*mcom.ProcessEntity{{
						Info: procDef,
						OptionalFlows: []*mcom.RecipeOptionalFlowEntity{{
							Name:           testProcessFlowName,
							Processes:      []mcom.ProcessDefinition{procFlowDef},
							MaxRepetitions: 0,
						}},
					}},
				},
			}}, result)
	}
	{ // ListRecipesByProduct: good case. test order
		recipeIDs := []string{"a1", "a2", "a3", "b1", "b2"}
		processOIDs := []string{
			"fe7a3ad1-2ca5-4621-a769-5bd100e8f1b3",
			"f7715831-9532-4c12-b3df-9f01cde43a5b",
			"e146dfc8-4855-4aa0-acc6-b6b4ffbb4679",
			"4f49901d-dcd8-49d7-9419-030a36a9162d",
			"a67a70ff-6f82-4d97-afaa-bdeedcc082d0",
		}
		reqs := make([]mcom.Recipe, 5)
		for i, id := range recipeIDs {
			reqs[i] = mcom.Recipe{
				ID: id,
				Product: mcom.Product{
					ID:   testProductID,
					Type: testProductType,
				},
				Version: mcom.RecipeVersion{Major: "1", Minor: "2", Stage: "UNSPECIFIED", ReleasedAt: releasedAt},
				Processes: []*mcom.Process{
					{
						OID: processOIDs[i],
						OptionalFlows: []*mcom.RecipeOptionalFlow{
							{
								Name:           testProcessFlowName,
								OIDs:           []string{processOIDs[i]},
								MaxRepetitions: 0,
							},
						},
					},
				},
				ProcessDefinitions: []mcom.ProcessDefinition{
					{
						OID:  processOIDs[i],
						Name: testProcessName,
						Type: testProductType,
						Configs: []*mcom.RecipeProcessConfig{
							{
								Stations:  []string{testStationA},
								BatchSize: types.Decimal.NewFromInt32(20),
								Unit:      "kg",
								Steps: []*mcom.RecipeProcessStep{{
									Materials: []*mcom.RecipeMaterial{{
										Name:  testMaterialName,
										Grade: testMaterialGrade,
										Value: mcom.RecipeMaterialParameter{
											High: types.Decimal.NewFromInt32(50),
											Mid:  types.Decimal.NewFromInt32(10),
											Low:  types.Decimal.NewFromInt32(5),
											Unit: "kg",
										},
										Site:             testRecipeSite,
										RequiredRecipeID: recipeIDs[i],
									}},
									Controls: []*mcom.RecipeProperty{{
										Name: testRecipePropertyName,
										Param: &mcom.RecipePropertyParameter{
											High: types.Decimal.NewFromInt32(50),
											Mid:  types.Decimal.NewFromInt32(10),
											Low:  types.Decimal.NewFromInt32(5),
											Unit: "kg",
										},
									}},
									Measurements: []*mcom.RecipeProperty{
										{
											Name: testRecipePropertyName,
											Param: &mcom.RecipePropertyParameter{
												High: types.Decimal.NewFromInt32(50),
												Mid:  types.Decimal.NewFromInt32(10),
												Low:  types.Decimal.NewFromInt32(5),
												Unit: "kg",
											},
										},
									},
								}},
							},
						},
						Output: mcom.OutputProduct{
							ID:   testProductID,
							Type: testProductType,
						},
						ProductValidPeriod: mcom.ProductValidPeriodConfig{
							Standing: testStanding,
							Expiry:   testExpiry,
						},
					},
				},
			}
		}
		assert.NoError(dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
			Recipes: reqs,
		}))

		rep, err := dm.ListRecipesByProduct(ctx, mcom.ListRecipesByProductRequest{
			ProductID: testProductID,
		}.WithOrder(mcom.Order{
			Name:       "id",
			Descending: false,
		}))
		assert.NoError(err)

		actual := make([]string, len(rep.Recipes))
		for i, rec := range rep.Recipes {
			actual[i] = rec.ID
		}

		assert.Equal([]string{"a1", "a2", "a3", "b1", "b2", "recipe_id"}, actual)
	}
	{ // GetRecipe: insufficient request.
		_, err := dm.GetRecipe(ctx, mcom.GetRecipeRequest{ID: ""})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // GetRecipe: recipe not found.
		_, err := dm.GetRecipe(ctx, mcom.GetRecipeRequest{ID: testRecipeIDNotExist})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_RECIPE_NOT_FOUND,
		}, err)
	}
	{ // GetRecipe: good case.
		result, err := dm.GetRecipe(ctx, mcom.GetRecipeRequest{ID: testRecipeID})
		assert.NoError(err)
		assert.Equal(mcom.GetRecipeReply{
			ID: testRecipeID,
			Product: mcom.Product{
				ID:   testProductID,
				Type: testProductType,
			},
			Version: mcom.RecipeVersion{Major: "1", Minor: "2", Stage: "UNSPECIFIED", ReleasedAt: releasedAt},
			Processes: []*mcom.ProcessEntity{{
				Info: procDef,
				OptionalFlows: []*mcom.RecipeOptionalFlowEntity{{
					Name:           testProcessFlowName,
					Processes:      []mcom.ProcessDefinition{procFlowDef},
					MaxRepetitions: 0,
				}},
			}},
		}, result)
	}
	{ // GetRecipe: good case without processes.
		result, err := dm.GetRecipe(ctx, mcom.GetRecipeRequest{ID: testRecipeID}.WithoutProcesses())
		assert.NoError(err)
		assert.Equal(mcom.GetRecipeReply{
			ID: testRecipeID,
			Product: mcom.Product{
				ID:   testProductID,
				Type: testProductType,
			},
			Version: mcom.RecipeVersion{Major: "1", Minor: "2", Stage: "UNSPECIFIED", ReleasedAt: releasedAt},
		}, result)
	}
	{ // DeleteRecipe: insufficient request.
		err := dm.DeleteRecipe(ctx, mcom.DeleteRecipeRequest{
			IDs: []string{},
		})
		assert.ErrorIs(mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}, err)
	}
	{ // DeleteRecipe: good case.
		err := dm.DeleteRecipe(ctx, mcom.DeleteRecipeRequest{
			IDs: []string{testRecipeID},
		})
		assert.NoError(err)
	}
}

func Test_GetProcessDefinition(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	assert := assert.New(t)
	cm := newClearMaster(db, &models.Recipe{}, &models.RecipeProcessDefinition{})
	assert.NoError(cm.Clear())
	{ // record not found.
		_, err := dm.GetProcessDefinition(ctx, mcom.GetProcessDefinitionRequest{
			RecipeID:    "not found",
			ProcessName: "not found",
			ProcessType: "not found",
		})
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_PROCESS_NOT_FOUND})
	}
	{ // good case.
		assert.NoError(dm.CreateRecipes(ctx, mcom.CreateRecipesRequest{
			Recipes: []mcom.Recipe{{
				ID: testRecipeID,
				Product: mcom.Product{
					ID:   testProductID,
					Type: testProductType,
				},
				Version:            mcom.RecipeVersion{Major: "1", Minor: "2", Stage: "UNSPECIFIED", ReleasedAt: types.ToTimeNano(time.Now())},
				Processes:          commonRecipeProcessesInfo,
				ProcessDefinitions: commonProcessDefinition,
			}},
		}))

		actual, err := dm.GetProcessDefinition(ctx, mcom.GetProcessDefinitionRequest{
			RecipeID:    testRecipeID,
			ProcessName: testProcessName,
			ProcessType: testProcessType,
		})
		assert.NoError(err)
		expected := mcom.GetProcessDefinitionReply{
			ProcessDefinition: procDef,
		}
		assert.Equal(expected, actual)
	}
	assert.NoError(cm.Clear())
}
