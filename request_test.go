package mcom

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/roles"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func Test_GetBatchRequest(t *testing.T) {
	assert := assert.New(t)

	{ // missing number.
		assert.ErrorIs(
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "number should be positive",
			},
			GetBatchRequest{
				WorkOrder: "wo",
			}.CheckInsufficiency())
	}
	{ // missing workorder.
		assert.ErrorIs(
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "empty work order is not allowed",
			},
			GetBatchRequest{
				Number: 2,
			}.CheckInsufficiency())
	}
	{ // good request.
		assert.NoError(GetBatchRequest{
			WorkOrder: "wo",
			Number:    2,
		}.CheckInsufficiency())
	}
}

func Test_FeedRequest(t *testing.T) {
	assert := assert.New(t)

	{ // missing FeedContent.
		assert.ErrorIs(FeedRequest{
			Batch: BatchID{
				WorkOrder: "wo",
				Number:    2,
			},
			FeedContent: []FeedPerSite{},
		}.CheckInsufficiency(), mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST})
	}
	{ // missing BatchID. good case
		assert.NoError(FeedRequest{
			FeedContent: []FeedPerSite{FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "name",
						Index: 1,
					},
					Station: "station",
				},
				FeedAll:  false,
				Quantity: decimal.NewFromInt(5),
			}},
		}.CheckInsufficiency())
	}
	{ // invalid quantity.
		assert.ErrorIs(FeedRequest{
			Batch: BatchID{
				WorkOrder: "wo",
				Number:    2,
			},
			FeedContent: []FeedPerSite{FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "name",
						Index: 1,
					},
					Station: "station",
				},
				FeedAll:  false,
				Quantity: decimal.NewFromInt(-5),
			}},
		}.CheckInsufficiency(), mcomErr.Error{Code: mcomErr.Code_INVALID_NUMBER})
	}
	{ // good request.
		assert.NoError(FeedRequest{
			Batch: BatchID{
				WorkOrder: "wo",
				Number:    2,
			},
			FeedContent: []FeedPerSite{FeedPerSiteType1{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "name",
						Index: 1,
					},
					Station: "station",
				},
				FeedAll:  false,
				Quantity: decimal.NewFromInt(5),
			}},
		}.CheckInsufficiency())
	}
}

func Test_CreateAccountsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty request.
		var req CreateAccountsRequest
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "CreateAccounts: empty request",
		})
	}
	{ // empty element.
		req := CreateAccountsRequest{{}}
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		})
	}
	{ // missing password.
		req := CreateAccountsRequest{
			CreateAccountRequest{
				ID:    "id",
				Roles: []roles.Role{roles.Role_BEARER},
			},
		}
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		})
	}
	{ // good case.
		req := CreateAccountsRequest{
			CreateAccountRequest{
				ID:    "id",
				Roles: []roles.Role{roles.Role_BEARER},
			}.WithDefaultPassword(),
		}
		assert.NoError(req.CheckInsufficiency())

		req = CreateAccountsRequest{
			CreateAccountRequest{
				ID:    "id",
				Roles: []roles.Role{roles.Role_BEARER},
			}.WithSpecifiedPassword("pw"),
		}
		assert.NoError(req.CheckInsufficiency())

		pw := req[0].GetPassword()
		assert.Equal(pw, "pw")
	}

}

func Test_UpdateAccountRequest(t *testing.T) {
	assert := assert.New(t)
	// #region insufficient request
	{ // missing id.
		var req UpdateAccountRequest
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST})
	}
	{ // good case.
		req := UpdateAccountRequest{
			UserID: "user",
		}
		assert.NoError(req.CheckInsufficiency())
	}
	// #endregion insufficient request

	// #region options
	{
		opt := ParseUpdateAccountOptions([]UpdateAccountOption{ResetPassword()})
		assert.Equal(UpdateAccountOptions{
			ResetPassword: true,
		}, opt)
	}
	// #endregion options

	{ // other functions.
		var req UpdateAccountRequest
		assert.False(req.IsChangingPassword())
		assert.False(req.IsModifyingRoles())

		req.ChangePassword = &struct {
			NewPassword string
			OldPassword string
		}{
			NewPassword: "new",
			OldPassword: "old",
		}
		assert.True(req.IsChangingPassword())
		req.Roles = []roles.Role{roles.Role_BEARER}
		assert.True(req.IsModifyingRoles())
	}
}

func Test_SignOutRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing token.
		var req SignOutRequest
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		})
	}
	{ // good case.
		req := SignOutRequest{Token: "hello"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_DeleteAccountRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing user id.
		var req DeleteAccountRequest
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "missing user id",
		})
	}
	{ // good case.
		req := DeleteAccountRequest{ID: "rockefeller"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateBatchRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty work order.
		req := CreateBatchRequest{Number: 3}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "empty work order is not allowed",
			})
	}
	{ // missing batch number.
		req := CreateBatchRequest{WorkOrder: "wo"}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "number should be greater than zero",
			})
	}
	{ // good case.
		req := CreateBatchRequest{WorkOrder: "wo", Number: 3}
		assert.NoError(req.CheckInsufficiency())
	}

}
func Test_UpdateBatchRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty work order.
		req := UpdateBatchRequest{Number: 1}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "empty work order or empty number",
			})
	}
	{ // empty batch number.
		req := UpdateBatchRequest{WorkOrder: "wo"}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "empty work order or empty number",
			})
	}
	{ // good case.
		req := UpdateBatchRequest{
			WorkOrder: "wo",
			Number:    2,
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListBatchesRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty work order.
		req := ListBatchesRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "empty work order is not allowed",
			},
		)
	}
	{ // good case.
		req := ListBatchesRequest{WorkOrder: "wo"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListCarriersRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing department oid.
		req := ListCarriersRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: "department OID is required",
			})
	}
	{ // good case.
		req := ListCarriersRequest{
			DepartmentOID: "department oid",
		}.WithPagination(PaginationRequest{
			PageCount:      1,
			ObjectsPerPage: 1,
		}).WithOrder([]Order{
			{
				Name:       "id_prefix",
				Descending: false,
			},
		}...,
		)
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetCarrierRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing carrier id.
		req := GetCarrierRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "carrier ID is required",
			})
	}
	{ // good case.
		req := GetCarrierRequest{
			ID: "carrier id",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_UpdateCarrierRequest(t *testing.T) {
	assert := assert.New(t)

	{ // missing carruer id.
		req := UpdateCarrierRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "id is required",
			})
	}
	{ // missing action.
		req := UpdateCarrierRequest{
			ID: "carrier id",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing action",
			})
	}
	{ // good case.
		req := UpdateCarrierRequest{
			ID: "carrier id",
			Action: BindResources{
				ResourcesID: []string{"A"},
			},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateCarrierRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing department oid.
		req := CreateCarrierRequest{
			IDPrefix: "AA",
			Quantity: 20,
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "illegal CreateCarrierRequest",
			})
	}
	{ // missing id prefix.
		req := CreateCarrierRequest{
			DepartmentOID: "department oid",
			Quantity:      20,
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "illegal CreateCarrierRequest",
			})
	}
	{ // illegal id prefix.
		req := CreateCarrierRequest{
			DepartmentOID: "department oid",
			IDPrefix:      "AAAAA",
			Quantity:      20,
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "illegal CreateCarrierRequest",
			})
	}
	{ // missing quantity.
		req := CreateCarrierRequest{
			DepartmentOID: "department oid",
			IDPrefix:      "AA",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "illegal CreateCarrierRequest",
			})
	}
	{ // good case.
		req := CreateCarrierRequest{
			DepartmentOID: "department oid",
			IDPrefix:      "AA",
			Quantity:      20,
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_DeleteCarrierRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing carrier id.
		req := DeleteCarrierRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "id is required",
			})
	}
	{ // good case.
		req := DeleteCarrierRequest{ID: "AA0001"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_DeleteDepartmentRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing department id.
		req := DeleteDepartmentRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := DeleteDepartmentRequest{DepartmentID: "department"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreatePackRecordRequest(t *testing.T) {
	assert := assert.New(t)
	{ // good case.
		req := CreatePackRecordsRequest{}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateProductionPlanRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing department id.
		req := CreateProductionPlanRequest{
			Date: time.Now(),
			Product: Product{
				ID:   "id",
				Type: "type",
			},
			Quantity: decimal.NewFromInt(20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing product id.
		req := CreateProductionPlanRequest{
			Date: time.Now(),
			Product: Product{
				Type: "type",
			},
			DepartmentOID: "department oid",
			Quantity:      decimal.NewFromInt(20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing product type.
		req := CreateProductionPlanRequest{
			Date: time.Now(),
			Product: Product{
				ID: "id",
			},
			DepartmentOID: "department oid",
			Quantity:      decimal.NewFromInt(20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing date.
		req := CreateProductionPlanRequest{
			Product: Product{
				ID:   "id",
				Type: "type",
			},
			DepartmentOID: "department oid",
			Quantity:      decimal.NewFromInt(20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // negative quantity.
		req := CreateProductionPlanRequest{
			Date: time.Now(),
			Product: Product{
				ID:   "id",
				Type: "type",
			},
			DepartmentOID: "department oid",
			Quantity:      decimal.NewFromInt(-20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "the Quantity is <= 0",
			})
	}
	{ // good case.
		req := CreateProductionPlanRequest{
			Date: time.Now(),
			Product: Product{
				ID:   "id",
				Type: "type",
			},
			DepartmentOID: "department oid",
			Quantity:      decimal.NewFromInt(20),
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListProductPlansRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing department oid.
		req := ListProductPlansRequest{
			Date:        time.Now(),
			ProductType: "type",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing date.
		req := ListProductPlansRequest{
			DepartmentOID: "department oid",
			ProductType:   "type",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing product type.
		req := ListProductPlansRequest{
			DepartmentOID: "department oid",
			Date:          time.Now(),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := ListProductPlansRequest{
			DepartmentOID: "department oid",
			Date:          time.Now(),
			ProductType:   "type",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListProductIDsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing type.
		req := ListProductIDsRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := ListProductIDsRequest{
			Type:          "type",
			IsLastProcess: false,
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListProductGroupsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing department oid.
		req := ListProductGroupsRequest{
			Type: "type",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing type.
		req := ListProductGroupsRequest{
			DepartmentOID: "department oid",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := ListProductGroupsRequest{
			DepartmentOID: "department oid",
			Type:          "type",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListProductTypesRequest(t *testing.T) {
	assert := assert.New(t)
	{ // good case.
		req := ListProductTypesRequest{}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetRecipeRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing id.
		req := GetRecipeRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := GetRecipeRequest{ID: "id"}
		assert.NoError(req.CheckInsufficiency())
		assert.True(req.NeedProcesses())
	}
	{ // good case.
		req := GetRecipeRequest{ID: "id"}.WithoutProcesses()
		assert.NoError(req.CheckInsufficiency())
		assert.False(req.NeedProcesses())
	}
}

func Test_CreateRecipesRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing recipes.
		req := CreateRecipesRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateRecipesRequest.Recipes' Error:Field validation for 'Recipes' failed on the 'min' tag",
			})
	}
	{ // missing recipe id.
		req := CreateRecipesRequest{
			Recipes: []Recipe{{
				Product: Product{ID: "id", Type: "type"},
				Version: RecipeVersion{Major: "1", Minor: "1", Stage: "1"},
				Processes: []*Process{{
					OID:           "OID",
					OptionalFlows: []*RecipeOptionalFlow{},
				}},
				ProcessDefinitions: []ProcessDefinition{{
					OID:     "oid",
					Name:    "name",
					Type:    "type",
					Configs: []*RecipeProcessConfig{},
					Output: OutputProduct{
						ID:   "id",
						Type: "type",
					},
				}},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateRecipesRequest.Recipes[0].ID' Error:Field validation for 'ID' failed on the 'required' tag",
			})
	}
	{ // missing recipe process.
		req := CreateRecipesRequest{
			Recipes: []Recipe{{
				ID:      "id",
				Product: Product{ID: "id", Type: "type"},
				Version: RecipeVersion{Major: "1", Minor: "1", Stage: "1"},
				ProcessDefinitions: []ProcessDefinition{{
					OID:     "oid",
					Name:    "name",
					Type:    "type",
					Configs: []*RecipeProcessConfig{},
					Output: OutputProduct{
						ID:   "id",
						Type: "type",
					},
				}},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateRecipesRequest.Recipes[0].Processes' Error:Field validation for 'Processes' failed on the 'min' tag",
			})
	}
	{ // missing recipe process definition.
		req := CreateRecipesRequest{
			Recipes: []Recipe{{
				ID:      "id",
				Product: Product{ID: "id", Type: "type"},
				Version: RecipeVersion{Major: "1", Minor: "1", Stage: "1"},
				Processes: []*Process{{
					OID:           "OID",
					OptionalFlows: []*RecipeOptionalFlow{},
				}},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateRecipesRequest.Recipes[0].ProcessDefinitions' Error:Field validation for 'ProcessDefinitions' failed on the 'min' tag",
			})
	}
	{ // missing product id.
		req := CreateRecipesRequest{
			Recipes: []Recipe{{
				ID:      "id",
				Product: Product{Type: "type"},
				Version: RecipeVersion{Major: "1", Minor: "1", Stage: "1"},
				Processes: []*Process{{
					OID:           "OID",
					OptionalFlows: []*RecipeOptionalFlow{},
				}},
				ProcessDefinitions: []ProcessDefinition{{
					OID:     "oid",
					Name:    "name",
					Type:    "type",
					Configs: []*RecipeProcessConfig{},
					Output: OutputProduct{
						ID:   "id",
						Type: "type",
					},
				}},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateRecipesRequest.Recipes[0].Product.ID' Error:Field validation for 'ID' failed on the 'required' tag",
			})
	}
	{ // missing product type.
		req := CreateRecipesRequest{
			Recipes: []Recipe{{
				ID:      "id",
				Product: Product{ID: "id"},
				Version: RecipeVersion{Major: "1", Minor: "1", Stage: "1"},
				Processes: []*Process{{
					OID:           "OID",
					OptionalFlows: []*RecipeOptionalFlow{},
				}},
				ProcessDefinitions: []ProcessDefinition{{
					OID:     "oid",
					Name:    "name",
					Type:    "type",
					Configs: []*RecipeProcessConfig{},
					Output: OutputProduct{
						ID:   "id",
						Type: "type",
					},
				}},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateRecipesRequest.Recipes[0].Product.Type' Error:Field validation for 'Type' failed on the 'required' tag",
			})
	}
	{ // missing version.
		req := CreateRecipesRequest{
			Recipes: []Recipe{{
				ID:      "id",
				Product: Product{ID: "id", Type: "type"},
				Processes: []*Process{{
					OID:           "OID",
					OptionalFlows: []*RecipeOptionalFlow{},
				}},
				ProcessDefinitions: []ProcessDefinition{{
					OID:     "oid",
					Name:    "name",
					Type:    "type",
					Configs: []*RecipeProcessConfig{},
					Output: OutputProduct{
						ID:   "id",
						Type: "type",
					},
				}},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateRecipesRequest.Recipes[0].Version.Major' Error:Field validation for 'Major' failed on the 'required' tag\nKey: 'CreateRecipesRequest.Recipes[0].Version.Stage' Error:Field validation for 'Stage' failed on the 'required' tag",
			})
	}
	{ // good case : missing released at
		req := CreateRecipesRequest{
			Recipes: []Recipe{{
				ID:      "id",
				Product: Product{ID: "id", Type: "type"},
				Version: RecipeVersion{Major: "1", Minor: "1", Stage: "1"},
				Processes: []*Process{{
					OID:           "OID",
					OptionalFlows: []*RecipeOptionalFlow{},
				}},
				ProcessDefinitions: []ProcessDefinition{{
					OID:     "oid",
					Name:    "name",
					Type:    "type",
					Configs: []*RecipeProcessConfig{},
					Output: OutputProduct{
						ID:   "id",
						Type: "type",
					},
				}},
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
	{ // good case.
		req := CreateRecipesRequest{
			Recipes: []Recipe{{
				ID:      "id",
				Product: Product{ID: "id", Type: "type"},
				Version: RecipeVersion{Major: "1", Minor: "1", Stage: "1", ReleasedAt: types.ToTimeNano(time.Now())},
				Processes: []*Process{{
					OID:           "OID",
					OptionalFlows: []*RecipeOptionalFlow{},
				}},
				ProcessDefinitions: []ProcessDefinition{{
					OID:     "oid",
					Name:    "name",
					Type:    "type",
					Configs: []*RecipeProcessConfig{},
					Output: OutputProduct{
						ID:   "id",
						Type: "type",
					},
				}},
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_DeleteRecipeRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing ids.
		req := DeleteRecipeRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := DeleteRecipeRequest{IDs: []string{"hello"}}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetCollectRecordRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing work order.
		req := GetCollectRecordRequest{Sequence: 2}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing sequence.
		req := GetCollectRecordRequest{WorkOrder: "wo"}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := GetCollectRecordRequest{WorkOrder: "wo", Sequence: 2}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListRecordsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing  date.
		req := ListRecordsRequest{DepartmentID: "id"}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing  department id.
		req := ListRecordsRequest{Date: time.Now()}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := ListRecordsRequest{
			Date:         time.Now(),
			DepartmentID: "id",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateCollectRecordRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing work order.
		req := CreateCollectRecordRequest{
			Sequence:    2,
			LotNumber:   "2",
			Station:     "station",
			ResourceOID: "resource oid",
			Quantity:    decimal.NewFromInt(20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing sequence.
		req := CreateCollectRecordRequest{
			WorkOrder:   "wo",
			LotNumber:   "2",
			Station:     "station",
			ResourceOID: "resource oid",
			Quantity:    decimal.NewFromInt(20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing lot number.
		req := CreateCollectRecordRequest{
			WorkOrder:   "wo",
			Sequence:    2,
			Station:     "station",
			ResourceOID: "resource oid",
			Quantity:    decimal.NewFromInt(20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing station.
		req := CreateCollectRecordRequest{
			WorkOrder:   "wo",
			Sequence:    2,
			LotNumber:   "2",
			ResourceOID: "resource oid",
			Quantity:    decimal.NewFromInt(20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing resource oid.
		req := CreateCollectRecordRequest{
			WorkOrder: "wo",
			Sequence:  2,
			LotNumber: "2",
			Station:   "station",
			Quantity:  decimal.NewFromInt(20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // invalid quantity.
		req := CreateCollectRecordRequest{
			WorkOrder:   "wo",
			Sequence:    2,
			LotNumber:   "2",
			Station:     "station",
			ResourceOID: "resource oid",
			Quantity:    decimal.NewFromInt(-20),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "the Quantity is <= 0",
			})
	}
	{ // good case.
		req := CreateCollectRecordRequest{
			WorkOrder:   "wo",
			Sequence:    2,
			LotNumber:   "2",
			Station:     "station",
			ResourceOID: "resource oid",
			Quantity:    decimal.NewFromInt(20),
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateMaterialResourcesRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty request.
		req := CreateMaterialResourcesRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "there is no resource to add",
			})
	}
	{ // good case.
		req := CreateMaterialResourcesRequest{
			Materials: []CreateMaterialResourcesRequestDetail{{
				Type:            "type",
				ID:              "id",
				Grade:           "A",
				Status:          resources.MaterialStatus_AVAILABLE,
				Quantity:        decimal.NewFromInt(20),
				PlannedQuantity: decimal.NewFromInt(20),
				Unit:            "kg",
				LotNumber:       "2",
				ProductionTime:  time.Now(),
				ExpiryTime:      time.Now(),
				ResourceID:      "resource id",
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
	{ // bad case: quantity/planned quantity is less than 0
		req := CreateMaterialResourcesRequest{
			Materials: []CreateMaterialResourcesRequestDetail{{
				Type:            "type",
				ID:              "id",
				Grade:           "A",
				Status:          resources.MaterialStatus_AVAILABLE,
				Quantity:        decimal.NewFromInt(-1),
				PlannedQuantity: decimal.NewFromInt(20),
				Unit:            "kg",
				LotNumber:       "2",
				ProductionTime:  time.Now(),
				ExpiryTime:      time.Now(),
				ResourceID:      "resource id",
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INVALID_NUMBER,
			})
	}
	{ // bad case: both quantity and planned quantity equal 0
		req := CreateMaterialResourcesRequest{
			Materials: []CreateMaterialResourcesRequestDetail{{
				Type:            "type",
				ID:              "id",
				Grade:           "A",
				Status:          resources.MaterialStatus_AVAILABLE,
				Quantity:        decimal.Zero,
				PlannedQuantity: decimal.Zero,
				Unit:            "kg",
				LotNumber:       "2",
				ProductionTime:  time.Now(),
				ExpiryTime:      time.Now(),
				ResourceID:      "resource id",
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INVALID_NUMBER,
			})
	}
}

func Test_MaterialResourceBindRequest(t *testing.T) {
	assert := assert.New(t)
	{ // over than 1 shared sites.
		req := MaterialResourceBindRequest{
			Station: "",
			Details: []MaterialBindRequestDetail{{}, {}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing station or only one shared site can be targeted",
			})
	}
	{ // missing site name.
		req := MaterialResourceBindRequest{
			Station: "station",
			Details: []MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{},
				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing SiteID",
			})
	}
	{ // missing bind type.
		req := MaterialResourceBindRequest{
			Station: "station",
			Details: []MaterialBindRequestDetail{{
				Site: models.SiteID{
					Name:  "site",
					Index: 0,
				},
				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing bind type",
			})
	}
	{ // set UnspecifiedQuantity with wrong bind type.
		req := MaterialResourceBindRequest{
			Station: "station",
			Details: []MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_COLLECTION_BIND,
				Site: models.SiteID{
					Name:  "site",
					Index: 0,
				},
				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    nil,
					ResourceID:  "A",
					ProductType: "A",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "unspecified quantity only applicable to slot bind or queue bind",
		})
	}
	{ // zero quantity is bad request.
		req := MaterialResourceBindRequest{
			Station: "station",
			Details: []MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "site",
					Index: 0,
				},

				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromFloat32(0),
					ResourceID:  "A",
					ProductType: "A",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "quantity cannot be less than 0 or not equal to nil",
		})
	}
	{ // invalid resources count.
		req := MaterialResourceBindRequest{
			Station: "station",
			Details: []MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "site",
					Index: 0,
				},

				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    nil,
					ResourceID:  "A",
					ProductType: "A",
				}, {
					Material: models.Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    nil,
					ResourceID:  "B",
					ProductType: "B",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "should contain only one resource",
		})
	}
	{ // good case.
		req := MaterialResourceBindRequest{
			Station: "station",
			Details: []MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "site",
					Index: 0,
				},

				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
	{ // good case unspecified quantity.
		req := MaterialResourceBindRequest{
			Station: "station",
			Details: []MaterialBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "site",
					Index: 0,
				},

				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    nil,
					ResourceID:  "A",
					ProductType: "A",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_MaterialResourceBindRequestV2(t *testing.T) {
	assert := assert.New(t)
	{ // missing bind type
		req := MaterialResourceBindRequestV2{
			Details: []MaterialBindRequestDetailV2{{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "site",
						Index: 0,
					},
					Station: "station",
				},
				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'MaterialResourceBindRequestV2.Details[0].Type' Error:Field validation for 'Type' failed on the 'gt' tag",
			})
	}
	{ // set UnspecifiedQuantity with wrong bind type
		req := MaterialResourceBindRequestV2{
			Details: []MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_COLLECTION_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "site",
						Index: 0,
					},
					Station: "station",
				},
				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    nil,
					ResourceID:  "A",
					ProductType: "A",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "unspecified quantity only applicable to slot bind or queue bind",
		})
	}
	{ // invalid resources count
		req := MaterialResourceBindRequestV2{
			Details: []MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "site",
						Index: 0,
					},
					Station: "station",
				},
				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    nil,
					ResourceID:  "A",
					ProductType: "A",
				}, {
					Material: models.Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    nil,
					ResourceID:  "B",
					ProductType: "B",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code:    mcomErr.Code_BAD_REQUEST,
			Details: "should contain only one resource",
		})
	}
	{ // good case
		req := MaterialResourceBindRequestV2{
			Details: []MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "site",
						Index: 0,
					},
					Station: "station",
				},
				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
	{ // good case unspecified quantity
		req := MaterialResourceBindRequestV2{
			Details: []MaterialBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  "site",
						Index: 0,
					},
					Station: "station",
				},
				Resources: []BindMaterialResource{{
					Material: models.Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    nil,
					ResourceID:  "A",
					ProductType: "A",
				}},
				Option: models.BindOption{},
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListSiteMaterialsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing station
		req := ListSiteMaterialsRequest{
			Site: models.SiteID{
				Name:  "site",
				Index: 0,
			},
		}
		assert.ErrorIs(req.CheckInsufficiency(), mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "Key: 'ListSiteMaterialsRequest.Station' Error:Field validation for 'Station' failed on the 'required' tag",
		})
	}
	{ // good case.
		req := ListSiteMaterialsRequest{
			Station: "station",
			Site: models.SiteID{
				Name:  "site",
				Index: 0,
			},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateSharedSiteRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing department id.
		req := CreateSharedSiteRequest{
			SiteInfo: SiteInformation{
				Name:       "site",
				Index:      0,
				Type:       sites.Type_SLOT,
				SubType:    sites.SubType_MATERIAL,
				Limitation: []string{},
			},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing type.
		req := CreateSharedSiteRequest{
			SiteInfo: SiteInformation{
				Name:       "site",
				Index:      0,
				SubType:    sites.SubType_MATERIAL,
				Limitation: []string{},
			},
			DepartmentID: "dep",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing sub type.
		req := CreateSharedSiteRequest{
			SiteInfo: SiteInformation{
				Name:       "site",
				Index:      0,
				Type:       sites.Type_SLOT,
				Limitation: []string{},
			},
			DepartmentID: "dep",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := CreateSharedSiteRequest{
			SiteInfo: SiteInformation{
				Name:       "site",
				Index:      0,
				Type:       sites.Type_SLOT,
				SubType:    sites.SubType_MATERIAL,
				Limitation: []string{},
			},
			DepartmentID: "dep",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_UpdateSharedSiteRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing site name.
		req := UpdateSharedSiteRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing action.
		req := UpdateSharedSiteRequest{
			TargetSite: SiteID{
				Name:  "site",
				Index: 0,
			},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := UpdateSharedSiteRequest{
			TargetSite: SiteID{
				Name:  "site",
				Index: 0,
			},
			AddToStations:      []string{"station"},
			RemoveFromStations: []string{},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_DeleteSharedSiteRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing site name.
		req := DeleteSharedSiteRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := DeleteSharedSiteRequest{
			TargetSite: SiteID{
				Name:  "site",
				Index: 0,
			},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListStationsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing department oid.
		req := ListStationsRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := ListStationsRequest{
			DepartmentOID: "oid",
		}.WithPagination(PaginationRequest{
			PageCount:      1,
			ObjectsPerPage: 3,
		}).WithOrder(Order{
			Name:       "id",
			Descending: true,
		})
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetStationRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing id.
		req := GetStationRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := GetStationRequest{
			ID: "id",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_StationGroupRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing id.
		req := StationGroupRequest{
			Stations: []string{"station"},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing stations.
		req := StationGroupRequest{
			ID: "id",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := StationGroupRequest{
			ID:       "id",
			Stations: []string{"station"},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_DeleteStationGroupRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing group id.
		req := DeleteStationGroupRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := DeleteStationGroupRequest{
			GroupID: "id",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetSiteRequest(t *testing.T) {
	assert := assert.New(t)
	{ // good case.
		req := GetSiteRequest{
			StationID: "station",
			SiteName:  "site",
			SiteIndex: 1,
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateStationRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing id.
		req := CreateStationRequest{
			DepartmentOID: "oid",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "the ID or departmentOID is empty",
			})
	}
	{ // missing department oid.
		req := CreateStationRequest{
			ID: "id",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "the ID or departmentOID is empty",
			})
	}
	{ // good case.
		req := CreateStationRequest{
			ID:            "id",
			DepartmentOID: "oid",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_UpdateStationRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing id.
		req := UpdateStationRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := UpdateStationRequest{ID: "id"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_DeleteStationRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing station id.
		req := DeleteStationRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := DeleteStationRequest{StationID: "station"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateUsersRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty user.
		req := CreateUsersRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "there is no user to create",
			})
	}
	{ // missing user id.
		req := CreateUsersRequest{
			Users: []User{{
				DepartmentID: "dep id",
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "the user id or the department id is empty",
			})
	}
	{ // missing user department.
		req := CreateUsersRequest{
			Users: []User{{
				ID: "id",
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "the user id or the department id is empty",
			})
	}
	{ // good case.
		req := CreateUsersRequest{
			Users: []User{{
				ID:           "id",
				DepartmentID: "dep id",
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_UpdateUserRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing id.
		req := UpdateUserRequest{
			Account: "account",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "the user id is empty or one of the department id or leave date or account is empty",
			})
	}
	{ // missing others.
		req := UpdateUserRequest{
			ID: "id",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "the user id is empty or one of the department id or leave date or account is empty",
			})
	}
	{ // good case.
		req := UpdateUserRequest{
			ID:      "id",
			Account: "account",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_UpdateDepartmentRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing old id.
		req := UpdateDepartmentRequest{NewID: "new"}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing new id.
		req := UpdateDepartmentRequest{OldID: "old"}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := UpdateDepartmentRequest{
			NewID: "new",
			OldID: "old",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_DeleteUserRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing id.
		req := DeleteUserRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := DeleteUserRequest{ID: "id"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListUnauthorizedUsersRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing department oid.
		req := ListUnauthorizedUsersRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := ListUnauthorizedUsersRequest{DepartmentOID: "oid"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_SignInStationRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing station.
		req := SignInStationRequest{
			Site: models.SiteID{Name: "site"},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing station",
			})
	}
	{ // good case.
		req := SignInStationRequest{
			Station: "station",
			Site:    models.SiteID{Name: "site"},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_SignOutStationRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing station.
		req := SignOutStationRequest{
			Site: models.SiteID{Name: "site"},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing station",
			})
	}
	{ // good case.
		req := SignOutStationRequest{
			Station: "station",
			Site:    models.SiteID{Name: "site"},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_WarehousingStockRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty resource ids.
		req := WarehousingStockRequest{
			Warehouse: Warehouse{
				ID:       "w",
				Location: "01",
			},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing required request",
			})
	}
	{ // missing warehouse id.
		req := WarehousingStockRequest{
			Warehouse: Warehouse{
				Location: "01",
			},
			ResourceIDs: []string{"rid"},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing required request",
			})
	}
	{ // missing warehouse location.
		req := WarehousingStockRequest{
			Warehouse: Warehouse{
				ID: "w",
			},
			ResourceIDs: []string{"rid"},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "missing required request",
			})
	}
	{ // missing resource id.
		req := WarehousingStockRequest{
			Warehouse: Warehouse{
				ID:       "w",
				Location: "01",
			},
			ResourceIDs: []string{""},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "some resourceID is empty",
			})
	}
	{ // good case.
		req := WarehousingStockRequest{
			Warehouse: Warehouse{
				ID:       "w",
				Location: "01",
			},
			ResourceIDs: []string{"rid"},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetResourceWarehouseRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing resource id.
		req := GetResourceWarehouseRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "empty resourceID",
			})
	}
	{ // good case.
		req := GetResourceWarehouseRequest{ResourceID: "id"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetMaterialResourceRequest(t *testing.T) {
	assert := assert.New(t)
	{ // good case.
		req := GetMaterialResourceRequest{ResourceID: "id"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetMaterialResourceIdentityRequest(t *testing.T) {
	assert := assert.New(t)
	{ // good case.
		req := GetMaterialResourceIdentityRequest{
			ResourceID:  "id",
			ProductType: "type",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListMaterialResourceIdentitiesRequest(t *testing.T) {
	assert := assert.New(t)
	{ // good case.
		req := ListMaterialResourceIdentitiesRequest{
			Details: []GetMaterialResourceIdentityRequest{{
				ResourceID:  "id",
				ProductType: "type",
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListMaterialResourcesRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing pagination request ObjectsPerPage.
		req := ListMaterialResourcesRequest{ProductType: "type"}.WithPagination(PaginationRequest{PageCount: 20})
		assert.ErrorIs(req.ValidatePagination(),
			mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'PaginationRequest.ObjectsPerPage' Error:Field validation for 'ObjectsPerPage' failed on the 'required' tag",
			})
	}
	{ // missing pagination request PageCount.
		req := ListMaterialResourcesRequest{ProductType: "type"}.WithPagination(PaginationRequest{ObjectsPerPage: 6})
		assert.ErrorIs(req.ValidatePagination(),
			mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'PaginationRequest.PageCount' Error:Field validation for 'PageCount' failed on the 'required' tag",
			})
	}
	{ // not invalid field name.
		req := ListMaterialResourcesRequest{ProductType: "type"}.WithOrder(Order{
			Name:       "invalidName",
			Descending: false,
		})
		assert.ErrorIs(req.ValidateOrder(),
			mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST,
				Details: "order: invalid field name",
			})
	}
	{ // good case.
		req := ListMaterialResourcesRequest{ProductType: "type"}
		assert.NoError(req.CheckInsufficiency())

		assert.Len(req.GetOrder(), 0)
	}
	{ // good case with pagination.
		req := ListMaterialResourcesRequest{ProductType: "type"}.WithPagination(PaginationRequest{PageCount: 20, ObjectsPerPage: 6})
		assert.NoError(req.ValidatePagination())
		assert.Equal(6, req.GetLimit())
		assert.Equal(19*6, req.GetOffset())
		assert.True(req.NeedPagination())
	}
	{ // good case with specified order.
		req := ListMaterialResourcesRequest{ProductType: "type"}.WithOrder(Order{
			Name:       "created_at",
			Descending: false,
		})
		assert.NoError(req.ValidateOrder())

		assert.Equal(req.GetOrder(), []Order{{
			Name:       "created_at",
			Descending: false,
		}})
	}
}

func Test_GetWorkOrderRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing id.
		req := GetWorkOrderRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := GetWorkOrderRequest{ID: "id"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateWorkOrdersRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty request.
		req := CreateWorkOrdersRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing process oid.
		req := CreateWorkOrdersRequest{WorkOrders: []CreateWorkOrder{{
			RecipeID:        "rid",
			ProcessName:     "pname",
			ProcessType:     "ptype",
			DepartmentOID:   "doid",
			BatchesQuantity: NewQuantityPerBatch([]decimal.Decimal{decimal.Zero}),
			Date:            time.Now(),
		}}}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing recipe id.
		req := CreateWorkOrdersRequest{WorkOrders: []CreateWorkOrder{{
			ProcessOID:      "oid",
			ProcessName:     "pname",
			ProcessType:     "ptype",
			DepartmentOID:   "doid",
			BatchesQuantity: NewQuantityPerBatch([]decimal.Decimal{decimal.Zero}),
			Date:            time.Now(),
		}}}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing process name.
		req := CreateWorkOrdersRequest{WorkOrders: []CreateWorkOrder{{
			ProcessOID:      "oid",
			RecipeID:        "rid",
			ProcessType:     "ptype",
			DepartmentOID:   "doid",
			BatchesQuantity: NewQuantityPerBatch([]decimal.Decimal{decimal.Zero}),
			Date:            time.Now(),
		}}}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing process type.
		req := CreateWorkOrdersRequest{WorkOrders: []CreateWorkOrder{{
			ProcessOID:      "oid",
			RecipeID:        "rid",
			ProcessName:     "pname",
			DepartmentOID:   "doid",
			BatchesQuantity: NewQuantityPerBatch([]decimal.Decimal{decimal.Zero}),
			Date:            time.Now(),
		}}}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing department oid.
		req := CreateWorkOrdersRequest{WorkOrders: []CreateWorkOrder{{
			ProcessOID:      "oid",
			RecipeID:        "rid",
			ProcessName:     "pname",
			ProcessType:     "ptype",
			BatchesQuantity: NewQuantityPerBatch([]decimal.Decimal{decimal.Zero}),
			Date:            time.Now(),
		}}}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing date.
		req := CreateWorkOrdersRequest{WorkOrders: []CreateWorkOrder{{
			ProcessOID:      "oid",
			RecipeID:        "rid",
			ProcessName:     "pname",
			ProcessType:     "ptype",
			DepartmentOID:   "doid",
			BatchesQuantity: NewQuantityPerBatch([]decimal.Decimal{decimal.Zero}),
		}}}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing batches quantity.
		req := CreateWorkOrdersRequest{WorkOrders: []CreateWorkOrder{{
			ProcessOID:    "oid",
			RecipeID:      "rid",
			ProcessName:   "pname",
			ProcessType:   "ptype",
			DepartmentOID: "doid",
			Date:          time.Now(),
		}}}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := CreateWorkOrdersRequest{WorkOrders: []CreateWorkOrder{{
			ProcessOID:      "oid",
			RecipeID:        "rid",
			ProcessName:     "pname",
			ProcessType:     "ptype",
			DepartmentOID:   "doid",
			BatchesQuantity: NewQuantityPerBatch([]decimal.Decimal{decimal.Zero}),
			Date:            time.Now(),
		}}}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_UpdateWorkOrdersRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing id.
		req := UpdateWorkOrdersRequest{Orders: []UpdateWorkOrder{{}}}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'UpdateWorkOrdersRequest.Orders[0].ID' Error:Field validation for 'ID' failed on the 'required' tag",
			})
	}
	{ // good case.
		req := UpdateWorkOrdersRequest{
			Orders: []UpdateWorkOrder{
				{
					ID: "id",
				},
			},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListWorkOrdersRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing station.
		req := ListWorkOrdersRequest{Date: time.Now()}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing date.
		req := ListWorkOrdersRequest{Station: "station"}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := ListWorkOrdersRequest{
			Date:    time.Now(),
			Station: "station",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListRecipesByProductRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing product id.
		req := ListRecipesByProductRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := ListRecipesByProductRequest{ProductID: "id"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListMaterialResourcesByIdRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty resources id.
		req := ListMaterialResourcesByIdRequest{
			ResourcesID: []string{},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ListMaterialResourcesByIdRequest.ResourcesID' Error:Field validation for 'ResourcesID' failed on the 'min' tag",
			})
	}
	{ // nil resources id.
		req := ListMaterialResourcesByIdRequest{
			ResourcesID: nil,
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ListMaterialResourcesByIdRequest.ResourcesID' Error:Field validation for 'ResourcesID' failed on the 'required' tag",
			})
	}
	{ // good case.
		req := ListMaterialResourcesByIdRequest{
			ResourcesID: []string{"123", "456"},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListWorkOrdersByDurationRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing department and station.
		req := ListWorkOrdersByDurationRequest{
			Since: time.Now(),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing since.
		req := ListWorkOrdersByDurationRequest{
			Until: time.Now(),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ListWorkOrdersByDurationRequest.Since' Error:Field validation for 'Since' failed on the 'required' tag",
			})
	}
	{ // invalid time value.
		req := ListWorkOrdersByDurationRequest{
			Since: time.Now().AddDate(0, 0, 1),
			Until: time.Now(),
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_BAD_REQUEST,
				Details: "invalid time value",
			})
	}
	{ // good case with department id and station.
		req := ListWorkOrdersByDurationRequest{
			DepartmentID: "department",
			Station:      "station",
			Since:        time.Now(),
		}
		assert.NoError(req.CheckInsufficiency())
	}
	{ // good case with department id.
		req := ListWorkOrdersByDurationRequest{
			DepartmentID: "department",
			Since:        time.Now(),
		}
		assert.NoError(req.CheckInsufficiency())
	}
	{ // good case with station.
		req := ListWorkOrdersByDurationRequest{
			Station: "station",
			Since:   time.Now(),
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetProcessDefinitionRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing recipe id.
		req := GetProcessDefinitionRequest{
			ProcessName: "process name",
			ProcessType: "process type",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing process name.
		req := GetProcessDefinitionRequest{
			RecipeID:    "recipe id",
			ProcessType: "process type",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing process type.
		req := GetProcessDefinitionRequest{
			RecipeID:    "recipe id",
			ProcessName: "process name",
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := GetProcessDefinitionRequest{
			RecipeID:    "recipe id",
			ProcessName: "process name",
			ProcessType: "process type",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListWorkOrdersByIDsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing id.
		req := ListWorkOrdersByIDsRequest{Station: "station"}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ListWorkOrdersByIDsRequest.IDs' Error:Field validation for 'IDs' failed on the 'required' tag",
			})
	}
	{ // good case.
		req := ListWorkOrdersByIDsRequest{
			IDs: []string{"id"},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}
func Test_SplitMaterialResourceRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing resource id.
		req := SplitMaterialResourceRequest{ProductType: "type"}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing product type.
		req := SplitMaterialResourceRequest{ResourceID: "id"}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := SplitMaterialResourceRequest{
			ResourceID:  "id",
			ProductType: "type",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListStationIDsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // good case.
		req := ListStationIDsRequest{
			DepartmentOID: "123",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetToolResourceRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing.
		req := GetToolResourceRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'GetToolResourceRequest.ResourceID' Error:Field validation for 'ResourceID' failed on the 'required' tag",
			})
	}
	{ // good case.
		req := GetToolResourceRequest{ResourceID: "id"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ToolResourceBindRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing details.
		req := ToolResourceBindRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ToolResourceBindRequest.Details' Error:Field validation for 'Details' failed on the 'gt' tag",
			})
	}
	{ // missing bind type
		req := ToolResourceBindRequest{
			Station: "",
			Details: []ToolBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND,
				Site: models.SiteID{
					Name: "name",
				},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ToolResourceBindRequest.Details[0].Type' Error:Field validation for 'Type' failed on the 'eq=1101|eq=1110' tag",
			})
	}
	{ // invalid bind type
		req := ToolResourceBindRequest{
			Station: "",
			Details: []ToolBindRequestDetail{{
				Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
				Site: models.SiteID{
					Name:  "site",
					Index: 0,
				},
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ToolResourceBindRequestV2(t *testing.T) {
	assert := assert.New(t)
	{ // missing details
		req := ToolResourceBindRequestV2{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ToolResourceBindRequestV2.Details' Error:Field validation for 'Details' failed on the 'gt' tag",
			})
	}
	{ // missing bind type
		req := ToolResourceBindRequestV2{
			Details: []ToolBindRequestDetailV2{{
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name: "name",
					},
					Station: "station",
				},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ToolResourceBindRequestV2.Details[0].Type' Error:Field validation for 'Type' failed on the 'required' tag",
			})
	}
	{ // invalid bind type
		req := ToolResourceBindRequestV2{
			Details: []ToolBindRequestDetailV2{{
				Type: bindtype.BindType_RESOURCE_BINDING_CONTAINER_BIND,
				Site: models.UniqueSite{
					SiteID: models.SiteID{
						Name: "name",
					},
					Station: "station",
				},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ToolResourceBindRequestV2.Details[0].Type' Error:Field validation for 'Type' failed on the 'eq=1101|eq=1110' tag",
			})
	}
	{ // empty detail.
		req := ToolResourceBindRequest{Station: "test", Details: []ToolBindRequestDetail{}}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ToolResourceBindRequest.Details' Error:Field validation for 'Details' failed on the 'gt' tag",
			})
	}
	{ // good case.
		req := ToolResourceBindRequest{
			Station: "",
			Details: []ToolBindRequestDetail{
				{
					Type: bindtype.BindType_RESOURCE_BINDING_SLOT_BIND,
					Site: models.SiteID{
						Name:  "site",
						Index: 0,
					},
				},
			},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListToolResourcesRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing resources id.
		req := ListToolResourcesRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ListToolResourcesRequest.ResourcesID' Error:Field validation for 'ResourcesID' failed on the 'required' tag",
			})
	}
	{ // missing resource id: empty slice.
		req := ListToolResourcesRequest{
			ResourcesID: []string{},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ListToolResourcesRequest.ResourcesID' Error:Field validation for 'ResourcesID' failed on the 'gt' tag",
			})
	}
	{ // missing resource id: empty string.
		req := ListToolResourcesRequest{
			ResourcesID: []string{"tool", ""},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ListToolResourcesRequest.ResourcesID[1]' Error:Field validation for 'ResourcesID[1]' failed on the 'required' tag",
			})
	}
	{ // good case.
		req := ListToolResourcesRequest{
			ResourcesID: []string{"tool"},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_SetStationConfigurationRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing station id.
		req := SetStationConfigurationRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'SetStationConfigurationRequest.StationID' Error:Field validation for 'StationID' failed on the 'required' tag",
			})
	}
	{ // good case.
		req := SetStationConfigurationRequest{StationID: "station"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetStationConfigurationRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing station id.
		req := GetStationConfigurationRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'GetStationConfigurationRequest.StationID' Error:Field validation for 'StationID' failed on the 'required' tag",
			})
	}
	{ // good case.
		req := GetStationConfigurationRequest{StationID: "station"}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_BindRecordsCheckRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing site info
		req := BindRecordsCheckRequest{
			Resources: []models.UniqueMaterialResource{{
				ResourceID:  "id",
				ProductType: "type",
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'BindRecordsCheckRequest.Site.Station' Error:Field validation for 'Station' failed on the 'required' tag",
			})
	}
	{ // missing resources.
		req := BindRecordsCheckRequest{
			Site: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "site",
					Index: 0,
				},
				Station: "station",
			},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'BindRecordsCheckRequest.Resources' Error:Field validation for 'Resources' failed on the 'required' tag",
			})
	}
	{ // empty resources.
		req := BindRecordsCheckRequest{
			Site: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "site",
					Index: 0,
				},
				Station: "station",
			},
			Resources: []models.UniqueMaterialResource{},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'BindRecordsCheckRequest.Resources' Error:Field validation for 'Resources' failed on the 'gt' tag",
			})
	}
	{ // good case.
		req := BindRecordsCheckRequest{
			Site: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "site",
					Index: 0,
				},
				Station: "station",
			},
			Resources: []models.UniqueMaterialResource{{
				ResourceID:  "id",
				ProductType: "type",
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateLimitaryHourRequest(t *testing.T) {
	assert := assert.New(t)

	{ // missing Limitary-Hour
		req := CreateLimitaryHourRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // missing Limitary ProductType.
		req := CreateLimitaryHourRequest{
			LimitaryHour: []LimitaryHour{{
				LimitaryHour: LimitaryHourParameter{
					Min: 0,
					Max: 1,
				},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // MIN or MAX should not be less than 0.
		req := CreateLimitaryHourRequest{
			LimitaryHour: []LimitaryHour{{
				ProductType: "CARBON RING",
				LimitaryHour: LimitaryHourParameter{
					Min: -1,
					Max: -1,
				},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // MIN must be less than MAX.
		req := CreateLimitaryHourRequest{
			LimitaryHour: []LimitaryHour{{
				ProductType: "CARBON RING",
				LimitaryHour: LimitaryHourParameter{
					Min: 1,
					Max: 0,
				},
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code: mcomErr.Code_INSUFFICIENT_REQUEST,
			})
	}
	{ // good case.
		req := CreateLimitaryHourRequest{
			LimitaryHour: []LimitaryHour{{
				ProductType: "CARBON RING",
				LimitaryHour: LimitaryHourParameter{
					Min: 0,
					Max: 1,
				},
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_SignOutStationsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing sites.
		req := SignOutStationsRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'SignOutStationsRequest.Sites' Error:Field validation for 'Sites' failed on the 'min' tag",
			})
	}
	{ // good case.
		req := SignOutStationsRequest{
			Sites: []models.UniqueSite{{}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_GetLimitaryHourRequest(t *testing.T) {
	assert := assert.New(t)
	{ // missing product type
		req := GetLimitaryHourRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "product type is required",
			})
	}
	{ // good case
		req := GetLimitaryHourRequest{
			ProductType: "product type",
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListAssociatedStationsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty request
		req := ListAssociatedStationsRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ListAssociatedStationsRequest.Site.Station' Error:Field validation for 'Station' failed on the 'required' tag",
			})
	}
	{ // missing station id
		req := ListAssociatedStationsRequest{
			Site: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "site",
					Index: 0,
				},
			},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'ListAssociatedStationsRequest.Site.Station' Error:Field validation for 'Station' failed on the 'required' tag",
			})
	}
	{ // good case
		req := ListAssociatedStationsRequest{
			Site: models.UniqueSite{
				SiteID: models.SiteID{
					Name:  "site",
					Index: 0,
				},
				Station: "station",
			},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_CreateBlobResourceRecordRequest(t *testing.T) {
	assert := assert.New(t)
	{ // empty request
		req := CreateBlobResourceRecordRequest{}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateBlobResourceRecordRequest.Details' Error:Field validation for 'Details' failed on the 'gt' tag",
			})
	}
	{ // missing blob uri
		req := CreateBlobResourceRecordRequest{
			Details: []CreateBlobResourceRecordDetail{{
				Resources:     []string{},
				Station:       "station",
				DateTime:      12345,
				ContainerName: "container",
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateBlobResourceRecordRequest.Details[0].BlobURI' Error:Field validation for 'BlobURI' failed on the 'required' tag",
			})
	}
	{ // missing station
		req := CreateBlobResourceRecordRequest{
			Details: []CreateBlobResourceRecordDetail{{
				BlobURI:       "uri",
				Resources:     []string{},
				DateTime:      12345,
				ContainerName: "container",
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateBlobResourceRecordRequest.Details[0].Station' Error:Field validation for 'Station' failed on the 'required' tag",
			})
	}
	{ // missing date time
		req := CreateBlobResourceRecordRequest{
			Details: []CreateBlobResourceRecordDetail{{
				BlobURI:       "uri",
				Resources:     []string{},
				Station:       "station",
				ContainerName: "container",
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateBlobResourceRecordRequest.Details[0].DateTime' Error:Field validation for 'DateTime' failed on the 'required' tag",
			})
	}
	{ // missing container name
		req := CreateBlobResourceRecordRequest{
			Details: []CreateBlobResourceRecordDetail{{
				BlobURI:   "uri",
				Resources: []string{},
				Station:   "station",
				DateTime:  12345,
			}},
		}
		assert.ErrorIs(req.CheckInsufficiency(),
			mcomErr.Error{
				Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
				Details: "Key: 'CreateBlobResourceRecordRequest.Details[0].ContainerName' Error:Field validation for 'ContainerName' failed on the 'required' tag",
			})
	}
	{ // good case
		req := CreateBlobResourceRecordRequest{
			Details: []CreateBlobResourceRecordDetail{{
				BlobURI:       "uri",
				Resources:     []string{},
				Station:       "station",
				DateTime:      12345,
				ContainerName: "container",
			}},
		}
		assert.NoError(req.CheckInsufficiency())
	}
}

func Test_ListBlobURIsRequest(t *testing.T) {
	assert := assert.New(t)
	{ // good case
		req := ListBlobURIsRequest{}
		assert.NoError(req.CheckInsufficiency())
	}
}
