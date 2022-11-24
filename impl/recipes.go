package impl

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/recipes"
)

// IsProductExisted implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) IsProductExisted(ctx context.Context, productID string) (bool, error) {
	if productID == "" {
		return false, mcomErr.Error{
			Code: mcomErr.Code_INSUFFICIENT_REQUEST,
		}
	}
	session := dm.newSession(ctx)
	return session.findProductID(productID)
}

func (session *session) findProductID(productID string) (bool, error) {
	var result models.RecipeProcessDefinition
	if err := session.db.Where(" product_id = ? ", productID).Take(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, mcomErr.Error{
				Code: mcomErr.Code_PRODUCT_ID_NOT_FOUND,
			}
		}
		return false, err
	}
	return true, nil
}

func getRecipeTool(t []*mcom.RecipeTool) []models.RecipeTool {
	tools := make([]models.RecipeTool, len(t))
	for i, v := range t {
		tools[i] = models.RecipeTool{
			Type:     v.Type,
			ID:       v.ID,
			Required: v.Required,
		}
	}
	return tools
}

func getRecipeMaterial(material []*mcom.RecipeMaterial) ([]models.RecipeMaterial, error) {
	rm := make([]models.RecipeMaterial, len(material))
	for i, v := range material {
		rm[i] = models.RecipeMaterial{
			ID:    v.Name,
			Grade: v.Grade,
			Value: models.RecipeMaterialParameter{
				High: v.Value.High,
				Mid:  v.Value.Mid,
				Low:  v.Value.Low,
				Unit: v.Value.Unit,
			},
			Site:             v.Site,
			RequiredRecipeID: v.RequiredRecipeID,
		}
	}
	return rm, nil
}

func parseRecipeProperty(property []*mcom.RecipeProperty) ([]models.RecipeProperty, error) {
	p := make([]models.RecipeProperty, len(property))
	for i, v := range property {
		p[i] = models.RecipeProperty{
			Name: v.Name,
			Param: models.RecipePropertyParameter{
				High: v.Param.High,
				Mid:  v.Param.Mid,
				Low:  v.Param.Low,
				Unit: v.Param.Unit,
			},
		}
	}
	return p, nil
}

func getProcStep(step []*mcom.RecipeProcessStep) ([]models.RecipeProcessStep, error) {
	procStep := make([]models.RecipeProcessStep, len(step))
	for i, v := range step {
		materials, err := getRecipeMaterial(v.Materials)
		if err != nil {
			return nil, err
		}
		controls, err := parseRecipeProperty(v.Controls)
		if err != nil {
			return nil, err
		}
		measurements, err := parseRecipeProperty(v.Measurements)
		if err != nil {
			return nil, err
		}

		procStep[i] = models.RecipeProcessStep{
			Materials:    materials,
			Controls:     controls,
			Measurements: measurements,
		}
	}
	return procStep, nil
}

func getOptionalFlow(opts []*mcom.RecipeOptionalFlow, procMap map[string]struct{}) ([]models.RecipeOptionalFlow, error) {
	optFlows := make([]models.RecipeOptionalFlow, len(opts))
	for i, v := range opts {
		for _, id := range v.OIDs {
			if _, ok := procMap[id]; !ok {
				return nil, mcomErr.Error{
					Code: mcomErr.Code_PROCESS_NOT_FOUND,
				}
			}
		}

		optFlows[i] = models.RecipeOptionalFlow{
			Name:           v.Name,
			Processes:      v.OIDs,
			MaxRepetitions: v.MaxRepetitions,
		}
	}
	return optFlows, nil
}

func getProcesses(processes []*mcom.Process, procMap map[string]struct{}) ([]models.RecipeProcess, error) {
	procs := make([]models.RecipeProcess, len(processes))
	for i, v := range processes {
		if _, ok := procMap[v.OID]; !ok {
			return nil, mcomErr.Error{
				Code: mcomErr.Code_PROCESS_NOT_FOUND,
			}
		}

		flows, err := getOptionalFlow(v.OptionalFlows, procMap)
		if err != nil {
			return nil, err
		}

		procs[i] = models.RecipeProcess{
			ReferenceOID:  v.OID,
			OptionalFlows: flows,
		}
	}
	return procs, nil
}

func parseRecipeInsertData(recipe mcom.Recipe, procMap map[string]struct{}) (models.Recipe, error) {
	procs, err := getProcesses(recipe.Processes, procMap)
	if err != nil {
		return models.Recipe{}, err
	}

	return models.Recipe{
		ID:          recipe.ID,
		ProductType: recipe.Product.Type,
		ProductID:   recipe.Product.ID,
		Major:       recipe.Version.Major,
		Minor:       recipe.Version.Minor,
		Stage:       recipes.StagePriority(recipes.StagePriority_value[recipe.Version.Stage]),
		ReleasedAt:  recipe.Version.ReleasedAt,
		Processes:   procs,
	}, nil
}

func getProcessConfigs(configs []*mcom.RecipeProcessConfig) ([]models.RecipeProcessConfig, error) {
	processConfigs := make([]models.RecipeProcessConfig, len(configs))
	for i, config := range configs {
		// where config.BatchSize == nil means that the batch size is defined by user.
		steps, err := getProcStep(config.Steps)
		if err != nil {
			return nil, err
		}

		controls, err := parseRecipeProperty(config.CommonControls)
		if err != nil {
			return nil, err
		}

		props, err := parseRecipeProperty(config.CommonProperties)
		if err != nil {
			return nil, err
		}

		processConfigs[i] = models.RecipeProcessConfig{
			Stations:         config.Stations,
			BatchSize:        config.BatchSize,
			Unit:             config.Unit,
			Tools:            getRecipeTool(config.Tools),
			Steps:            steps,
			CommonControls:   controls,
			CommonProperties: props,
		}
	}

	return processConfigs, nil
}

// create all RecipeProcessDefinition and return OIDs mapping for compare recipe reference.
func (tx *txDataManager) createProcessDefinition(recipeID string, processes []mcom.ProcessDefinition) (map[string]struct{}, error) {
	procs := make([]models.RecipeProcessDefinition, len(processes))
	OIDs := make(map[string]struct{}, len(processes))
	for i, proc := range processes {
		OIDs[proc.OID] = struct{}{}

		configs, err := getProcessConfigs(proc.Configs)
		if err != nil {
			return nil, err
		}

		procs[i] = models.RecipeProcessDefinition{
			OID:      proc.OID,
			Name:     proc.Name,
			Type:     proc.Type,
			Configs:  configs,
			RecipeID: sql.NullString{String: recipeID, Valid: true},
			OutputProduct: models.OutputProduct{
				ID:   proc.Output.ID,
				Type: proc.Output.Type,
			},
			ProductValidPeriods: models.ProductValidPeriod{
				Standing: uint16(proc.ProductValidPeriod.Standing),
				Expiry:   uint16(proc.ProductValidPeriod.Expiry),
			},
		}
	}

	if err := tx.db.Create(&procs).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return nil, mcomErr.Error{
				Code: mcomErr.Code_PROCESS_ALREADY_EXISTS,
			}
		}
		return nil, err
	}
	return OIDs, nil
}

// CreateRecipes implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) CreateRecipes(ctx context.Context, req mcom.CreateRecipesRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	tx := dm.beginTx(ctx)
	defer tx.Rollback() // nolint: errcheck

	recipeInsertData := make([]models.Recipe, len(req.Recipes))
	for i, recipe := range req.Recipes {
		procMap, err := tx.createProcessDefinition(recipe.ID, recipe.ProcessDefinitions)
		if err != nil {
			return err
		}

		recipeInfo, err := parseRecipeInsertData(recipe, procMap)
		if err != nil {
			return err
		}
		recipeInsertData[i] = recipeInfo
	}

	if err := tx.db.Create(&recipeInsertData).Error; err != nil {
		if IsPqError(err, UniqueViolation) {
			return mcomErr.Error{
				Code: mcomErr.Code_RECIPE_ALREADY_EXISTS,
			}
		}
		return err
	}
	return tx.Commit()
}

func getProcessOIDs(recipes []models.Recipe) []string {
	processIDs := []string{}
	for _, recipe := range recipes {
		for _, proc := range recipe.Processes {
			processIDs = append(processIDs, proc.ReferenceOID)
			for _, flow := range proc.OptionalFlows {
				processIDs = append(processIDs, flow.Processes...)
			}
		}
	}
	return processIDs
}

// DeleteRecipe implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) DeleteRecipe(ctx context.Context, req mcom.DeleteRecipeRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	tx := dm.beginTx(ctx)
	defer tx.db.Rollback()

	recipes := []models.Recipe{}
	if err := tx.db.Where(` id IN ? `, req.IDs).Find(&recipes).Error; err != nil {
		return err
	}

	processIDs := getProcessOIDs(recipes)
	if err := tx.db.Where(` oid IN ? `, processIDs).Delete(&models.RecipeProcessDefinition{}).Error; err != nil {
		return err
	}

	if err := tx.db.Where(` id IN ? `, req.IDs).Delete(&models.Recipe{}).Error; err != nil {
		return err
	}

	return tx.Commit()
}

func parseRecipeTools(recipeTools []models.RecipeTool) []*mcom.RecipeTool {
	tools := make([]*mcom.RecipeTool, len(recipeTools))
	for i, v := range recipeTools {
		tools[i] = &mcom.RecipeTool{
			Type:     v.Type,
			ID:       v.ID,
			Required: v.Required,
		}
	}
	return tools
}

func parseRecipeMaterials(recipeMtrls []models.RecipeMaterial) []*mcom.RecipeMaterial {
	mtrls := make([]*mcom.RecipeMaterial, len(recipeMtrls))
	for i, v := range recipeMtrls {
		mtrls[i] = &mcom.RecipeMaterial{
			Name:  v.ID,
			Grade: v.Grade,
			Value: mcom.RecipeMaterialParameter{
				High: v.Value.High,
				Mid:  v.Value.Mid,
				Low:  v.Value.Low,
				Unit: v.Value.Unit,
			},
			Site:             v.Site,
			RequiredRecipeID: v.RequiredRecipeID,
		}
	}
	return mtrls
}

func parseRecipeProperties(recipeProp []models.RecipeProperty) []*mcom.RecipeProperty {
	prop := make([]*mcom.RecipeProperty, len(recipeProp))
	for i, v := range recipeProp {
		prop[i] = &mcom.RecipeProperty{
			Name: v.Name,
			Param: &mcom.RecipePropertyParameter{
				High: v.Param.High,
				Mid:  v.Param.Mid,
				Low:  v.Param.Low,
				Unit: v.Param.Unit,
			},
		}
	}
	return prop
}

func parseRecipeProcessStep(procSteps []models.RecipeProcessStep) []*mcom.RecipeProcessStep {
	steps := make([]*mcom.RecipeProcessStep, len(procSteps))
	for i, v := range procSteps {
		steps[i] = &mcom.RecipeProcessStep{
			Materials:    parseRecipeMaterials(v.Materials),
			Controls:     parseRecipeProperties(v.Controls),
			Measurements: parseRecipeProperties(v.Measurements),
		}
	}
	return steps
}

func parseProcessConfigs(procConfigs models.RecipeProcessConfigs) []*mcom.RecipeProcessConfig {
	configs := make([]*mcom.RecipeProcessConfig, len(procConfigs))
	for i, v := range procConfigs {
		configs[i] = &mcom.RecipeProcessConfig{
			Stations:         v.Stations,
			BatchSize:        v.BatchSize,
			Unit:             v.Unit,
			Tools:            parseRecipeTools(v.Tools),
			Steps:            parseRecipeProcessStep(v.Steps),
			CommonControls:   parseRecipeProperties(v.CommonControls),
			CommonProperties: parseRecipeProperties(v.CommonProperties),
		}
	}
	return configs
}

func (session *session) getRecipeProcessDefinition(oids []string) ([]models.RecipeProcessDefinition, error) {
	proc := []models.RecipeProcessDefinition{}
	if err := session.db.
		Where(` oid IN ? `, oids).
		Order(` product_id `).
		Find(&proc).Error; err != nil {
		return nil, err
	}
	return proc, nil
}

func (session *session) parseProcessDefinition(procOIDs []string) ([]mcom.ProcessDefinition, error) {
	procs, err := session.getRecipeProcessDefinition(procOIDs)
	if err != nil {
		return nil, err
	}
	if len(procs) == 0 {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, mcomErr.Error{
				Code: mcomErr.Code_PROCESS_NOT_FOUND,
			}
		}
	}

	flows := make([]mcom.ProcessDefinition, len(procs))
	for i, v := range procs {
		flows[i] = mcom.ProcessDefinition{
			OID:     v.OID,
			Name:    v.Name,
			Type:    v.Type,
			Configs: parseProcessConfigs(v.Configs),
			Output: mcom.OutputProduct{
				ID:   v.OutputProduct.ID,
				Type: v.OutputProduct.Type,
			},
			ProductValidPeriod: mcom.ProductValidPeriodConfig{
				Standing: time.Duration(v.ProductValidPeriods.Standing),
				Expiry:   time.Duration(v.ProductValidPeriods.Expiry),
			},
		}
	}
	return flows, nil
}

func (session *session) parseRecipeOptionalFlow(recipeOpts []models.RecipeOptionalFlow) ([]*mcom.RecipeOptionalFlowEntity, error) {
	opts := make([]*mcom.RecipeOptionalFlowEntity, len(recipeOpts))
	for i, v := range recipeOpts {
		processes, err := session.parseProcessDefinition(v.Processes)
		if err != nil {
			return nil, err
		}
		opts[i] = &mcom.RecipeOptionalFlowEntity{
			Name:           v.Name,
			Processes:      processes,
			MaxRepetitions: v.MaxRepetitions,
		}
	}
	return opts, nil
}

func (session *session) parseProcessesEntity(recipeProcesses models.RecipeProcesses) ([]*mcom.ProcessEntity, error) {
	oids := make([]string, len(recipeProcesses))
	for i, v := range recipeProcesses {
		oids[i] = v.ReferenceOID
	}

	procs, err := session.getRecipeProcessDefinition(oids)
	if err != nil {
		return nil, err
	}

	procEnts := make([]*mcom.ProcessEntity, len(procs))
	for i, v := range procs {
		flows, err := session.parseRecipeOptionalFlow(recipeProcesses[i].OptionalFlows)
		if err != nil {
			return nil, err
		}

		procEnts[i] = &mcom.ProcessEntity{
			Info: mcom.ProcessDefinition{
				OID:     v.OID,
				Name:    v.Name,
				Type:    v.Type,
				Configs: parseProcessConfigs(v.Configs),
				Output: mcom.OutputProduct{
					ID:   v.OutputProduct.ID,
					Type: v.OutputProduct.Type,
				},
				ProductValidPeriod: mcom.ProductValidPeriodConfig{
					Standing: time.Duration(v.ProductValidPeriods.Standing),
					Expiry:   time.Duration(v.ProductValidPeriods.Expiry),
				},
			},
			OptionalFlows: flows,
		}
	}
	return procEnts, nil
}

func (session *session) parseRecipes(recipeResult []models.Recipe) ([]mcom.GetRecipeReply, error) {
	repliedRecipes := make([]mcom.GetRecipeReply, 0, len(recipeResult))
	for _, v := range recipeResult {
		procs, err := session.parseProcessesEntity(v.Processes)
		if err != nil {
			return nil, err
		}

		if len(procs) == 0 {
			continue
		}

		repliedRecipes = append(repliedRecipes, mcom.GetRecipeReply{
			ID: v.ID,
			Product: mcom.Product{
				ID:   v.ProductID,
				Type: v.ProductType,
			},
			Version: mcom.RecipeVersion{
				Major:      v.Major,
				Minor:      v.Minor,
				Stage:      recipes.StagePriority_name[int32(v.Stage)],
				ReleasedAt: v.ReleasedAt,
			},
			Processes: procs,
		})
	}
	return repliedRecipes, nil
}

// ListRecipesByProduct implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListRecipesByProduct(ctx context.Context, req mcom.ListRecipesByProductRequest) (mcom.ListRecipesByProductReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListRecipesByProductReply{}, err
	}

	oids := []string{}
	session := dm.newSession(ctx)
	if err := session.db.
		Model(&models.RecipeProcessDefinition{}).
		Select(`oid`).
		Where(`product_id = ?`, req.ProductID).
		Scan(&oids).Error; err != nil {
		return mcom.ListRecipesByProductReply{}, err
	}

	_, result, err := listHandler[models.Recipe](&session, req, func(d *gorm.DB) *gorm.DB {
		return d.Raw(`
	SELECT id,product_id,product_type,released_at,major,minor,stage,processes FROM `+
			(&models.Recipe{}).TableName()+` ,
	JSONB_ARRAY_ELEMENTS(processes) AS array_element
	WHERE (array_element ->> 'reference_oid')::uuid IN ? GROUP BY id ORDER BY id
`, oids)
	})
	if err != nil {
		return mcom.ListRecipesByProductReply{}, err
	}
	res, err := session.parseRecipes(result)
	return mcom.ListRecipesByProductReply(mcom.ListRecipesByProductReply{Recipes: res}), err
}

func (session *session) parseSingleRecipe(recipeResult models.Recipe, needProcesses bool) (mcom.GetRecipeReply, error) {
	var procs []*mcom.ProcessEntity
	if needProcesses {
		var err error
		procs, err = session.parseProcessesEntity(recipeResult.Processes)
		if err != nil {
			return mcom.GetRecipeReply{}, err
		}
	}

	return mcom.GetRecipeReply{
		ID: recipeResult.ID,
		Product: mcom.Product{
			ID:   recipeResult.ProductID,
			Type: recipeResult.ProductType,
		},
		Version: mcom.RecipeVersion{
			Major:      recipeResult.Major,
			Minor:      recipeResult.Minor,
			Stage:      recipes.StagePriority_name[int32(recipeResult.Stage)],
			ReleasedAt: recipeResult.ReleasedAt,
		},
		Processes: procs,
	}, nil
}

// GetRecipe implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetRecipe(ctx context.Context, req mcom.GetRecipeRequest) (mcom.GetRecipeReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetRecipeReply{}, err
	}

	result := models.Recipe{}
	session := dm.newSession(ctx)
	if err := session.db.Where(&models.Recipe{ID: string(req.ID)}).
		Take(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetRecipeReply{}, mcomErr.Error{
				Code: mcomErr.Code_RECIPE_NOT_FOUND,
			}
		}
		return mcom.GetRecipeReply{}, err
	}
	return session.parseSingleRecipe(result, req.NeedProcesses())
}

func (dm *DataManager) GetProcessDefinition(ctx context.Context, req mcom.GetProcessDefinitionRequest) (mcom.GetProcessDefinitionReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetProcessDefinitionReply{}, err
	}

	session := dm.newSession(ctx)
	var process models.RecipeProcessDefinition
	if err := session.db.
		Where(`recipe_id = ? AND name = ? AND type = ?`, req.RecipeID, req.ProcessName, req.ProcessType).
		Take(&process).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetProcessDefinitionReply{}, mcomErr.Error{Code: mcomErr.Code_PROCESS_NOT_FOUND}
		}
		return mcom.GetProcessDefinitionReply{}, err
	}

	return mcom.GetProcessDefinitionReply{ProcessDefinition: mcom.ProcessDefinition{
		OID:     process.OID,
		Name:    process.Name,
		Type:    process.Type,
		Configs: parseProcessConfigs(process.Configs),
		Output:  mcom.OutputProduct(process.OutputProduct),
		ProductValidPeriod: mcom.ProductValidPeriodConfig(mcom.ProductValidPeriodConfig{
			Standing: time.Duration(process.ProductValidPeriods.Standing),
			Expiry:   time.Duration(process.ProductValidPeriods.Expiry),
		}),
	}}, nil
}
