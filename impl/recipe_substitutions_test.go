package impl

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func TestDataManager_Substitutions(t *testing.T) {
	ctx, dm, db := initializeDB(t)
	assert := assert.New(t)
	assert.NoError(dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
		ProductID: models.ProductID{
			ID:    "Hello",
			Grade: "A",
		},
		Contents: []models.Substitution{{
			ID:         "Hi",
			Grade:      "S",
			Proportion: decimal.NewFromFloat(0.95),
		}, {
			ID:         "Bonjour",
			Grade:      "B",
			Proportion: decimal.NewFromFloat(1.1),
		}},
	}))
	{ // normal case (should pass)
		actual, err := dm.ListSubstitutions(ctx, mcom.ListSubstitutionsRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
		})
		assert.NoError(err)
		expected := mcom.ListSubstitutionsReply{
			Substitutions: []models.Substitution{{
				ID:         "Bonjour",
				Grade:      "B",
				Proportion: decimal.NewFromFloat(1.1),
			}, {
				ID:         "Hi",
				Grade:      "S",
				Proportion: decimal.NewFromFloat(0.95),
			}},
			UpdatedAt: actual.UpdatedAt,
			UpdatedBy: "",
		}
		assert.Equal(expected, actual)
	}
	{ // ListSubstitutions Code_INSUFFICIENT_REQUEST (should not pass)
		_, err := dm.ListSubstitutions(ctx, mcom.ListSubstitutionsRequest{
			ProductID: models.ProductID{},
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "ProductID is required",
		}, err)
	}
	{ // AddSubstitutions Code_INSUFFICIENT_REQUEST (should not pass)
		err := dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "Kitty",
			},
		})
		assert.ErrorIs(mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: "adding with empty contents is not allowed",
		}, err)
	}
	{ // add substitutions.
		assert.NoError(dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
			Contents: []models.Substitution{{
				ID:         "Addition1",
				Grade:      "A",
				Proportion: decimal.NewFromInt(1),
			}, {
				ID:         "Addition2",
				Grade:      "A",
				Proportion: decimal.NewFromInt(1),
			}},
		}))
		actual, err := dm.ListSubstitutions(ctx, mcom.ListSubstitutionsRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
		})
		assert.NoError(err)
		expected := mcom.ListSubstitutionsReply{
			Substitutions: []models.Substitution{{
				ID:         "Addition1",
				Grade:      "A",
				Proportion: decimal.NewFromInt(1),
			}, {
				ID:         "Addition2",
				Grade:      "A",
				Proportion: decimal.NewFromInt(1),
			}, {
				ID:         "Bonjour",
				Grade:      "B",
				Proportion: decimal.NewFromFloat(1.1),
			}, {
				ID:         "Hi",
				Grade:      "S",
				Proportion: decimal.NewFromFloat(0.95),
			}},
			UpdatedAt: actual.UpdatedAt,
			UpdatedBy: "",
		}
		assert.Equal(expected, actual)
	}
	{ // remove substitutions.
		assert.NoError(dm.DeleteSubstitutions(ctx, mcom.DeleteSubstitutionsRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
			Contents: []mcom.SubstitutionDeletionContent{{
				ProductID: models.ProductID{
					ID:    "Addition1",
					Grade: "A",
				},
			}, {
				ProductID: models.ProductID{
					ID:    "Addition2",
					Grade: "A",
				},
			}},
		}))
		actual, err := dm.ListSubstitutions(ctx, mcom.ListSubstitutionsRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
		})
		assert.NoError(err)
		expected := mcom.ListSubstitutionsReply{
			Substitutions: []models.Substitution{{
				ID:         "Bonjour",
				Grade:      "B",
				Proportion: decimal.NewFromFloat(1.1),
			}, {
				ID:         "Hi",
				Grade:      "S",
				Proportion: decimal.NewFromFloat(0.95),
			}},
			UpdatedAt: actual.UpdatedAt,
			UpdatedBy: "",
		}
		assert.Equal(expected, actual)
	}
	{ // update substitutions.
		assert.NoError(dm.UpdateSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
			Contents: []models.Substitution{{
				ID:         "Bonjour",
				Grade:      "B",
				Proportion: decimal.NewFromFloat(1.5),
			}},
		}))
		actual, err := dm.ListSubstitutions(ctx, mcom.ListSubstitutionsRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
		})
		assert.NoError(err)
		expected := mcom.ListSubstitutionsReply{
			Substitutions: []models.Substitution{{
				ID:         "Bonjour",
				Grade:      "B",
				Proportion: decimal.NewFromFloat(1.5),
			}},
			UpdatedAt: actual.UpdatedAt,
			UpdatedBy: "",
		}
		assert.Equal(expected, actual)
	}
	{ // case delete substitutions with empty grade.
		assert.NoError(dm.UpdateSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{
				ID:    "TestProduct",
				Grade: "",
			},
			Contents: []models.Substitution{{
				ID:         "TestSubstitution",
				Grade:      "",
				Proportion: decimal.NewFromFloat(0.5),
			}, {
				ID:         "TestSubstitution2",
				Grade:      "",
				Proportion: decimal.NewFromFloat(0.5),
			}},
		}))

		assert.NoError(dm.DeleteSubstitutions(ctx, mcom.DeleteSubstitutionsRequest{
			ProductID: models.ProductID{
				ID:    "TestProduct",
				Grade: "",
			},
			Contents: []mcom.SubstitutionDeletionContent{{
				ProductID: models.ProductID{
					ID:    "TestSubstitution",
					Grade: "",
				},
			}},
			DeleteAll: false,
		}))

		actual, err := dm.ListSubstitutions(ctx, mcom.ListSubstitutionsRequest{
			ProductID: models.ProductID{
				ID:    "TestProduct",
				Grade: "",
			},
		})
		assert.NoError(err)

		assert.Equal(mcom.ListSubstitutionsReply{
			Substitutions: []models.Substitution{{
				ID:         "TestSubstitution2",
				Grade:      "",
				Proportion: decimal.NewFromFloat(0.5),
			}},
			UpdatedAt: actual.UpdatedAt,
			UpdatedBy: "",
		}, actual)
	}
	{ // clear substitutions.
		assert.NoError(dm.DeleteSubstitutions(ctx, mcom.DeleteSubstitutionsRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
			DeleteAll: true,
		}))
		actual, err := dm.ListSubstitutions(ctx, mcom.ListSubstitutionsRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
		})
		assert.NoError(err)
		expected := mcom.ListSubstitutionsReply{Substitutions: []models.Substitution{}}
		assert.Equal(expected, actual)
	}
	{ // ListMultipleSubstitutions.
		assert.NoError(dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{
				ID:    "ProductA",
				Grade: "A",
			},
			Contents: []models.Substitution{{
				ID:         "SubstitutionB",
				Grade:      "B",
				Proportion: decimal.NewFromFloat(0.5),
			}, {
				ID:         "SubstitutionC",
				Grade:      "C",
				Proportion: decimal.NewFromFloat(0.8),
			}},
		}))
		assert.NoError(dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{
				ID:    "ProductB",
				Grade: "B",
			},
			Contents: []models.Substitution{{
				ID:         "SubstitutionD",
				Grade:      "D",
				Proportion: decimal.NewFromFloat(0.5),
			}, {
				ID:         "SubstitutionE",
				Grade:      "E",
				Proportion: decimal.NewFromFloat(0.8),
			}},
		}))
		assert.NoError(dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{
				ID:    "ProductC",
				Grade: "C",
			},
			Contents: []models.Substitution{{
				ID:         "SubstitutionF",
				Grade:      "F",
				Proportion: decimal.NewFromFloat(0.5),
			}, {
				ID:         "SubstitutionG",
				Grade:      "G",
				Proportion: decimal.NewFromFloat(0.8),
			}},
		}))
		actual, err := dm.ListMultipleSubstitutions(ctx, mcom.ListMultipleSubstitutionsRequest{
			ProductIDs: []models.ProductID{{
				ID:    "ProductC",
				Grade: "C",
			}, {
				ID:    "ProductA",
				Grade: "A",
			}},
		})
		assert.NoError(err)

		assert.Equal(mcom.ListMultipleSubstitutionsReply{Reply: map[models.ProductID]mcom.ListSubstitutionsReply{
			{
				ID:    "ProductC",
				Grade: "C",
			}: {
				Substitutions: models.Substitutions{{
					ID:         "SubstitutionF",
					Grade:      "F",
					Proportion: decimal.NewFromFloat(0.5),
				}, {
					ID:         "SubstitutionG",
					Grade:      "G",
					Proportion: decimal.NewFromFloat(0.8),
				}},
				UpdatedAt: actual.Reply[models.ProductID{
					ID:    "ProductC",
					Grade: "C",
				}].UpdatedAt,
				UpdatedBy: "",
			},
			{
				ID:    "ProductA",
				Grade: "A",
			}: {
				Substitutions: models.Substitutions{{
					ID:         "SubstitutionB",
					Grade:      "B",
					Proportion: decimal.NewFromFloat(0.5),
				}, {
					ID:         "SubstitutionC",
					Grade:      "C",
					Proportion: decimal.NewFromFloat(0.8),
				}},
				UpdatedAt: actual.Reply[models.ProductID{
					ID:    "ProductA",
					Grade: "A",
				}].UpdatedAt,
				UpdatedBy: "",
			},
		}}, actual)
	}
	{ // add substitutions with empty grade (should pass)
		assert.NoError(dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{ID: "ProductA"},
			Contents: []models.Substitution{{
				ID:         "NoGradeSubstitutionA",
				Grade:      "",
				Proportion: decimal.NewFromFloat(1.23),
			}},
		}))
	}
	{ // list product with empty grade.
		actual, err := dm.ListSubstitutions(ctx, mcom.ListSubstitutionsRequest{
			ProductID: models.ProductID{ID: "ProductA"},
		})
		assert.NoError(err)
		assert.Equal(mcom.ListSubstitutionsReply{
			Substitutions: []models.Substitution{{
				ID:         "NoGradeSubstitutionA",
				Grade:      "",
				Proportion: decimal.NewFromFloat(1.23),
			}},
			UpdatedAt: actual.UpdatedAt,
			UpdatedBy: "",
		}, actual)
	}
	{ // add duplicated substitutions in two requests (should not pass)
		assert.NoError(dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
			Contents: []models.Substitution{{
				ID:         "D",
				Grade:      "",
				Proportion: decimal.NewFromInt(1),
			}},
		}))

		assert.ErrorIs(dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
			Contents: []models.Substitution{{
				ID:         "D",
				Grade:      "",
				Proportion: decimal.Zero,
			}},
		}), mcomErr.Error{
			Code:    mcomErr.Code_SUBSTITUTION_ALREADY_EXISTS,
			Details: "ID: D, Grade: ",
		})
	}
	{ // add duplicated substitutions in one request (should not pass)
		assert.ErrorIs(dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{
				ID:    "Hello",
				Grade: "A",
			},
			Contents: []models.Substitution{
				{
					ID:         "DD",
					Grade:      "",
					Proportion: decimal.NewFromFloat(0.2),
				},
				{
					ID:         "DD",
					Grade:      "",
					Proportion: decimal.NewFromFloat(0.5),
				},
			},
		}), mcomErr.Error{
			Code:    mcomErr.Code_SUBSTITUTION_ALREADY_EXISTS,
			Details: "ID: DD, Grade: ",
		})
	}
	{ // add substitutions with empty grade again (should pass)
		assert.NoError(dm.AddSubstitutions(ctx, mcom.BasicSubstitutionRequest{
			ProductID: models.ProductID{ID: "ProductA"},
			Contents: []models.Substitution{{
				ID:         "NoGradeSubstitutionC",
				Grade:      "",
				Proportion: decimal.NewFromFloat(1.23),
			}},
		}))
	}

	assert.NoError(clearAllSubstitutionMapping(db))
	assert.NoError(dm.Close())
}

func clearAllSubstitutionMapping(db *gorm.DB) error {
	return newClearMaster(db, &models.SubstitutionMapping{}).Clear()
}
