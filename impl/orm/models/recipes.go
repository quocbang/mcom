package models

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/utils/recipes"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// RecipeTool definition.
type RecipeTool struct {
	Type string `json:"type"`
	ID   string `json:"id"`

	// Required means the tool is required for production.
	Required bool `json:"required"`
}

// RecipeMaterialParameter definition.
type RecipeMaterialParameter struct {
	// High may be nil.
	High *decimal.Decimal `json:"high,omitempty"`
	// Mid may be nil.
	Mid *decimal.Decimal `json:"mid,omitempty"`
	// Low may be nil.
	Low *decimal.Decimal `json:"low,omitempty"`

	Unit string `json:"unit"`
}

// RecipeMaterial definition.
type RecipeMaterial struct {
	ID    string `json:"name"`
	Grade string `json:"grade"`

	Value RecipeMaterialParameter `json:"value"`

	// Site is the site where this material is required from.
	Site string `json:"site"`

	// RequiredRecipeID we are required the material is made by specific recipe.
	// The value is relative to Recipe.ID.
	RequiredRecipeID string `json:"required_recipe_id"`
}

// Substitution Definition.
type Substitution struct {
	ID         string
	Grade      string
	Proportion decimal.Decimal
}

// ProductID definition.
type ProductID struct {
	ID    string
	Grade string
}

func (id ProductID) getSubstitutionMapping(db *gorm.DB) (SubstitutionMapping, error) {
	var res SubstitutionMapping
	if err := db.Where(`id = ? AND grade = ?`, id.ID, id.Grade).Take(&res).Error; err != nil {
		// has no substitutions.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SubstitutionMapping{
				ID:            id.ID,
				Grade:         id.Grade,
				Substitutions: []Substitution{},
			}, nil
		}
		return SubstitutionMapping{}, err
	}
	return res, nil
}

// ListSubstitutions returns a substitution slice according to the RecipeMaterial.
func (id ProductID) ListSubstitutions(db *gorm.DB) (ListSubstitutionsReply, error) {
	substitutionMapping, err := id.getSubstitutionMapping(db)
	if err != nil {
		return ListSubstitutionsReply{}, err
	}

	// sort by id then by grade.
	res := substitutionMapping.Substitutions
	sort.Slice(res, func(i, j int) bool {
		if res[i].ID == res[j].ID {
			return res[i].Grade < res[j].Grade
		}
		return res[i].ID < res[j].ID
	})

	return ListSubstitutionsReply{
		Substitutions: res,
		UpdatedAt:     substitutionMapping.UpdatedAt,
		UpdatedBy:     substitutionMapping.UpdatedBy,
	}, nil
}

type ListSubstitutionsReply struct {
	Substitutions []Substitution
	UpdatedAt     types.TimeNano
	UpdatedBy     string
}

// SetSubstitutions replaces the substitutions by specified contents.
func (id ProductID) SetSubstitutions(ctx context.Context, db *gorm.DB, contents Substitutions) error {
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
			{Name: "grade"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"substitutions": contents,
		}),
	}).Create(&SubstitutionMapping{
		ID:            id.ID,
		Grade:         id.Grade,
		Substitutions: contents,
	}).Error
}

// AddSubstitutions adds new substitutions.
func (id ProductID) AddSubstitutions(ctx context.Context, db *gorm.DB, contents Substitutions) error {
	substitutionMapping, err := id.getSubstitutionMapping(db)
	if err != nil {
		return err
	}

	target := substitutionMapping.Substitutions
	contents.ForEach(func(s Substitution) {
		if target.Any(func(ts Substitution) bool {
			return ts.ID == s.ID && ts.Grade == s.Grade
		}) {
			err = mcomErr.Error{
				Code:    mcomErr.Code_SUBSTITUTION_ALREADY_EXISTS,
				Details: fmt.Sprintf("ID: %s, Grade: %s", s.ID, s.Grade),
			}
		}
		target = target.Append(s)
	})
	if err != nil {
		return err
	}

	updatedBy := commonsCtx.UserID(ctx)
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
			{Name: "grade"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"substitutions": target.Distinct(),
			"updated_by":    updatedBy,
		}),
	}).Create(&SubstitutionMapping{
		ID:            id.ID,
		Grade:         id.Grade,
		Substitutions: target.Distinct(),
		UpdatedBy:     updatedBy,
	}).Error
}

// UpdateSubstitutions updates the proportion of substitutions by material id and grade.
func (id ProductID) UpdateSubstitutions(ctx context.Context, db *gorm.DB, contents Substitutions) error {
	substitutionMapping, err := id.getSubstitutionMapping(db)
	if err != nil {
		return err
	}

	target := substitutionMapping.Substitutions
	var res Substitutions
	target.ForEach(func(s Substitution) {
		for _, content := range contents {
			if content.ID == s.ID && content.Grade == s.Grade {
				res = res.Append(Substitution{
					ID:         content.ID,
					Grade:      content.Grade,
					Proportion: content.Proportion,
				})
			} else {
				res = res.Append(s)
			}
		}
	})

	updatedBy := commonsCtx.UserID(ctx)
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
			{Name: "grade"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"substitutions": res,
			"updated_by":    updatedBy,
		}),
	}).Create(&SubstitutionMapping{
		ID:            id.ID,
		Grade:         id.Grade,
		Substitutions: res,
		UpdatedBy:     updatedBy,
	}).Error
}

// RemoveSubstitutions removes specified substitutions (ID+Grade)
func (id ProductID) RemoveSubstitutions(ctx context.Context, db *gorm.DB, contents Substitutions) error {
	substitutionMapping, err := id.getSubstitutionMapping(db)
	if err != nil {
		return err
	}
	target := substitutionMapping.Substitutions

	if len(target) == 0 {
		return nil
	}

	updatedBy := commonsCtx.UserID(ctx)

	if target.RemoveAll(func(s Substitution) bool {
		return contents.Any(func(ss Substitution) bool {
			return ss.ID == s.ID && ss.Grade == s.Grade
		})
	}) != 0 {
		return db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "id"},
				{Name: "grade"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"substitutions": target,
				"updated_by":    updatedBy,
			}),
		}).Create(&SubstitutionMapping{
			ID:            id.ID,
			Grade:         id.Grade,
			Substitutions: target,
			UpdatedBy:     updatedBy,
		}).Error
	}
	return nil
}

// ClearSubstitutions clears all substitutions.
func (id ProductID) ClearSubstitutions(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Delete(&SubstitutionMapping{
		ID:    id.ID,
		Grade: id.Grade,
	}).Error
}

// SubstitutionMapping Definition.
type SubstitutionMapping struct {
	ID            string         `gorm:"type:text;primaryKey"`
	Grade         string         `gorm:"type:text;primaryKey"`
	Substitutions Substitutions  `gorm:"type:jsonb;default:'[]';not null"`
	UpdatedAt     types.TimeNano `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy     string         `gorm:"type:text;not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (SubstitutionMapping) TableName() string {
	return "substitution_mapping"
}

// RecipePropertyParameter definition.
type RecipePropertyParameter struct {
	// High may be nil.
	High *decimal.Decimal `json:"high,omitempty"`
	// Mid may be nil.
	Mid *decimal.Decimal `json:"mid,omitempty"`
	// Low may be nil.
	Low *decimal.Decimal `json:"min,omitempty"`

	Unit string `json:"unit"`
}

// RecipeProperty sees https://gitlab.kenda.com.tw/kenda/commons/-/blob/master/src/proto/dm/rs/recipe.proto#L84
// message Property for details.
type RecipeProperty struct {
	// Name sees https://gitlab.kenda.com.tw/kenda/commons/-/blob/master/src/proto/dm/rs/enums.proto#L337
	// enum PropertyName value mapping.
	Name  string                  `json:"name"`
	Param RecipePropertyParameter `json:"param"`
}

// RecipeProcessStep definition.
type RecipeProcessStep struct {
	Materials    []RecipeMaterial `json:"materials"`
	Controls     []RecipeProperty `json:"controls"`
	Measurements []RecipeProperty `json:"measurements"`
}

// RecipeProcessConfig defines the needs in the process, e.g. where the station
// the operator works in, what tools we need, what materials we will feed, how
// to work right etc.
type RecipeProcessConfig struct {
	// Stations mean where to execute the process. Values are relative to Station.ID or
	// StationGroup.ID.
	Stations  []string         `json:"stations"`
	BatchSize *decimal.Decimal `json:"batch_size"`
	Unit      string           `json:"unit"`
	// Tools are needs in production.
	Tools            []RecipeTool        `json:"tools"`
	Steps            []RecipeProcessStep `json:"steps"`
	CommonControls   []RecipeProperty    `json:"common_controls"`
	CommonProperties []RecipeProperty    `json:"common_properties"`
}

// RecipeProcessConfigs definition.
type RecipeProcessConfigs []RecipeProcessConfig

// Scan implements database/sql Scanner interface.
func (cs *RecipeProcessConfigs) Scan(src interface{}) error {
	return ScanJSON(src, cs)
}

// Value implements database/sql/driver Valuer interface.
func (cs RecipeProcessConfigs) Value() (driver.Value, error) {
	return json.Marshal(cs)
}

// Scan implements database/sql Scanner interface.
func (cs *ProductValidPeriod) Scan(src interface{}) error {
	return ScanJSON(src, cs)
}

// Value implements database/sql/driver Valuer interface.
func (cs ProductValidPeriod) Value() (driver.Value, error) {
	return json.Marshal(cs)
}

// OutputProduct definition.
type OutputProduct struct {
	ID   string `gorm:"column:product_id;type:text;not null;index:idx_output_product_id"`
	Type string `gorm:"column:product_type;type:text;not null"`
}

// ProductValidPeriod definition
type ProductValidPeriod struct {
	Standing uint16 `json:"standing"`
	Expiry   uint16 `json:"expiry"`
}

// RecipeProcessDefinition definition.
type RecipeProcessDefinition struct {
	OID      string               `gorm:"column:oid;type:uuid;primaryKey"`
	RecipeID sql.NullString       `gorm:"type:text;uniqueIndex:unique_process_definition"`
	Name     string               `gorm:"type:text;not null;uniqueIndex:unique_process_definition"`
	Type     string               `gorm:"not null;uniqueIndex:unique_process_definition"`
	Configs  RecipeProcessConfigs `gorm:"type:json;not null"`
	OutputProduct
	ProductValidPeriods ProductValidPeriod `gorm:"column:product_valid_period;type:jsonb;default:'{}'"` // unit: hours.
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (RecipeProcessDefinition) TableName() string {
	return "recipe_process_definition"
}

// RecipeOptionalFlow definition.
type RecipeOptionalFlow struct {
	Name string `json:"name"`
	// Processes are OIDs relative to RecipeProcessDefinition.OID.
	Processes []string `json:"processes"`
	// MaxRepetitions is the maximum times of the process executed.
	MaxRepetitions int32 `json:"max_repetitions"`
}

// RecipeProcess is what the process is.
type RecipeProcess struct {
	// ReferenceOID is the OID relative to RecipeProcessDefinition.OID.
	ReferenceOID string `json:"reference_oid"`
	// OptionalFlows are how to handle major defects on the product. The user.
	// could choose some of them to handle.
	OptionalFlows []RecipeOptionalFlow `json:"optional_flows,omitempty"`
}

// RecipeProcesses definition.
type RecipeProcesses []RecipeProcess

// Scan implements database/sql Scanner interface.
func (ps *RecipeProcesses) Scan(src interface{}) error {
	return ScanJSON(src, ps)
}

// Value implements database/sql/driver Valuer interface.
func (ps RecipeProcesses) Value() (driver.Value, error) {
	return json.Marshal(ps)
}

// Recipe definition.
type Recipe struct {
	ID string `gorm:"type:text;primaryKey"`

	ProductType string `gorm:"not null"`
	ProductID   string `gorm:"type:text;not null"`

	Major string                `gorm:"type:text;not null"`
	Minor string                `gorm:"type:text"`
	Stage recipes.StagePriority `gorm:"not null"`

	ReleasedAt types.TimeNano  `gorm:"released_at;not null"`
	Processes  RecipeProcesses `gorm:"type:jsonb;not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (Recipe) TableName() string {
	return "recipe"
}
