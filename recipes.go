package mcom

import (
	"time"

	"github.com/shopspring/decimal"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// RecipeMaterialParameter definition.
type RecipeMaterialParameter struct {
	High *decimal.Decimal
	Mid  *decimal.Decimal
	Low  *decimal.Decimal
	Unit string
}

// RecipeMaterial definition.
type RecipeMaterial struct {
	Name             string
	Grade            string
	Value            RecipeMaterialParameter
	Site             string
	RequiredRecipeID string
}

// RecipePropertyParameter definition.
type RecipePropertyParameter struct {
	High *decimal.Decimal
	Mid  *decimal.Decimal
	Low  *decimal.Decimal
	Unit string
}

// RecipeProperty definition.
type RecipeProperty struct {
	Name  string
	Param *RecipePropertyParameter
}

// RecipeTool definition.
type RecipeTool struct {
	Type     string
	ID       string
	Required bool
}

// RecipeProcessStep definition.
type RecipeProcessStep struct {
	Materials    []*RecipeMaterial
	Controls     []*RecipeProperty
	Measurements []*RecipeProperty
}

// RecipeProcessConfig definition.
type RecipeProcessConfig struct {
	Stations         []string
	BatchSize        *decimal.Decimal
	Unit             string
	Tools            []*RecipeTool
	Steps            []*RecipeProcessStep
	CommonControls   []*RecipeProperty
	CommonProperties []*RecipeProperty
}

// OutputProduct definition.
type OutputProduct struct {
	ID   string
	Type string
}

// ProcessDefinition definition.
type ProcessDefinition struct {
	OID                string
	Name               string
	Type               string
	Configs            []*RecipeProcessConfig
	Output             OutputProduct
	ProductValidPeriod ProductValidPeriodConfig
}

// RecipeOptionalFlowEntity definition.
type RecipeOptionalFlowEntity struct {
	Name           string
	Processes      []ProcessDefinition
	MaxRepetitions int32
}

// ProcessEntity definition.
type ProcessEntity struct {
	Info          ProcessDefinition
	OptionalFlows []*RecipeOptionalFlowEntity
}

// RecipeVersion definition.
type RecipeVersion struct {
	Major      string `validate:"required"`
	Minor      string
	Stage      string `validate:"required"`
	ReleasedAt types.TimeNano
}

// GetRecipeReply definition.
type GetRecipeReply struct {
	ID        string
	Product   Product
	Version   RecipeVersion
	Processes []*ProcessEntity
}

// ListRecipesByProductRequest definition.
type ListRecipesByProductRequest struct {
	ProductID    string
	orderRequest OrderRequest
}

func (req ListRecipesByProductRequest) WithOrder(o ...Order) ListRecipesByProductRequest {
	req.orderRequest = OrderRequest{OrderBy: o}
	return req
}

// NeedOrder implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListRecipesByProductRequest) NeedOrder() bool {
	return len(req.orderRequest.OrderBy) != 0
}

// ValidateOrder implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListRecipesByProductRequest) ValidateOrder() error {
	for _, order := range req.orderRequest.OrderBy {
		if !isValidName[models.Recipe](order.Name) {
			return mcomErr.Error{Code: mcomErr.Code_BAD_REQUEST, Details: "order: invalid field name"}
		}
	}
	return nil
}

// GetOrderFunc implements gitlab.kenda.com.tw/kenda/mcom Orderable interface.
func (req ListRecipesByProductRequest) GetOrder() []Order {
	return req.orderRequest.OrderBy
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListRecipesByProductRequest) CheckInsufficiency() error {
	if req.ProductID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// ListRecipesByProductReply definition.
type ListRecipesByProductReply struct {
	Recipes []GetRecipeReply
}

// GetRecipeRequest definition.
type GetRecipeRequest struct {
	ID          string
	noProcesses bool
}

func (req GetRecipeRequest) WithoutProcesses() GetRecipeRequest {
	req.noProcesses = true
	return req
}

func (req GetRecipeRequest) NeedProcesses() bool {
	return !req.noProcesses
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetRecipeRequest) CheckInsufficiency() error {
	if req.ID == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

type GetProcessDefinitionRequest struct {
	RecipeID    string
	ProcessName string
	ProcessType string
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req GetProcessDefinitionRequest) CheckInsufficiency() error {
	if req.RecipeID == "" || req.ProcessName == "" || req.ProcessType == "" {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

type GetProcessDefinitionReply struct {
	ProcessDefinition
}

// RecipeOptionalFlow definition.
type RecipeOptionalFlow struct {
	Name           string
	OIDs           []string
	MaxRepetitions int32
}

// Process definition.
type Process struct {
	OID           string
	OptionalFlows []*RecipeOptionalFlow
}

// Recipe definition.
type Recipe struct {
	ID      string        `validate:"required"`
	Product Product       `validate:"required,dive"`
	Version RecipeVersion `validate:"required,dive"`

	Processes          []*Process          `validate:"min=1"`
	ProcessDefinitions []ProcessDefinition `validate:"min=1"`
}

// CreateRecipesRequest definition.
type CreateRecipesRequest struct {
	Recipes []Recipe `validate:"min=1,dive"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateRecipesRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{
			Code:    mcomErr.Code_INSUFFICIENT_REQUEST,
			Details: err.Error(),
		}
	}
	return nil
}

// DeleteRecipeRequest definition.
type DeleteRecipeRequest struct {
	// recipe IDs.
	IDs []string
}

// ListSubstitutionsRequest definition.
type ListSubstitutionsRequest struct {
	ProductID models.ProductID
}

// ListMultipleSubstitutionsRequest definition.
type ListMultipleSubstitutionsRequest struct {
	ProductIDs []models.ProductID
}

// ListSubstitutionsReply definition.
type ListSubstitutionsReply struct {
	Substitutions []models.Substitution
	UpdatedAt     types.TimeNano
	UpdatedBy     string
}

// ListMultipleSubstitutionsReply definition.
type ListMultipleSubstitutionsReply struct {
	Reply map[models.ProductID]ListSubstitutionsReply
}

// DeleteSubstitutionsRequest definition.
type DeleteSubstitutionsRequest struct {
	ProductID models.ProductID
	Contents  SubstitutionDeletionContents
	DeleteAll bool
}

// SubstitutionDeletionContent definition.
type SubstitutionDeletionContent struct {
	ProductID models.ProductID
}

// SubstitutionDeletionContents definition.
type SubstitutionDeletionContents []SubstitutionDeletionContent

// ToSubstitution converts mcom.SubstitutionDeletionContent to models.Substitution.
func (sdc SubstitutionDeletionContent) ToSubstitution() models.Substitution {
	return models.Substitution{
		ID:    sdc.ProductID.ID,
		Grade: sdc.ProductID.Grade,
	}
}

// ToSubstitutions converts mcom.SubstitutionDeletionContents to models.Substitutions.
func (sdcs SubstitutionDeletionContents) ToSubstitutions() models.Substitutions {
	res := models.NewSubstitutions([]models.Substitution{})
	for _, sdc := range sdcs {
		res = res.Append(sdc.ToSubstitution())
	}
	return res
}

// BasicSubstitutionRequest definition.
type BasicSubstitutionRequest struct {
	ProductID models.ProductID
	Contents  []models.Substitution
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req DeleteRecipeRequest) CheckInsufficiency() error {
	if len(req.IDs) == 0 {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST}
	}
	return nil
}

// GetMaterialExtendDateRequest definition.
type GetMaterialExtendDateRequest struct {
	MaterialID string
}

// GetMaterialExtendDateReply definition.
type GetMaterialExtendDateReply time.Duration

// ProductValidPeriodConfig definition.
type ProductValidPeriodConfig struct {
	Standing time.Duration
	Expiry   time.Duration
}
