package mcom

import (
	"context"
)

type pdaMethods interface {
	// GetMaterial gets materials information to be control or release.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_RESOURCE_NOT_FOUND
	GetMaterial(context.Context, GetMaterialRequest) (GetMaterialReply, error)

	// ListControlReasons gets control reason list.
	ListControlReasons(context.Context) (ListControlReasonsReply, error)

	// UpdateMaterial materials control or release.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_RESOURCE_NOT_FOUND
	UpdateMaterial(context.Context, UpdateMaterialRequest) error

	// ListChangeableStatus gets the statuses can be changed.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_RESOURCE_NOT_FOUND
	ListChangeableStatus(context.Context, ListChangeableStatusRequest) (ListChangeableStatusReply, error)

	// ListControlAreas operates area.
	ListControlAreas(context.Context) (ListControlAreasReply, error)

	// GetMaterialExtendDate gets extendable days for material.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_RESOURCE_NOT_FOUND
	GetMaterialExtendDate(context.Context, GetMaterialExtendDateRequest) (GetMaterialExtendDateReply, error)
}

type cloudMethods interface {
	// CreateBlobResourceRecord adds a new blob record to db.
	//
	// CreateBlobResourceRecord needs the following required input:
	//  - BlobURI
	//  - Station
	//  - DateTime
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	CreateBlobResourceRecord(context.Context, CreateBlobResourceRecordRequest) error

	// ListBlobURIs lists blob URIs according to specified conditions.
	ListBlobURIs(context.Context, ListBlobURIsRequest) (ListBlobURIsReply, error)
}

// DataManager is for data access.
//
// The returned parameters which are time.Time type are in local time.
//
// Methods may return an error that is able to be assert as Error struct,
// which we call USER_ERROR, in gitlab.kenda.com.tw/kenda/mcom/errors package,
// it is useful to invoke 'As' function, in gitlab.kenda.com.tw/kenda/mcom/errors
// package, like `e, ok := errors.As(err)`.
type DataManager interface {
	pdaMethods
	cloudMethods
	// Close closes the connection.
	Close() error

	// GetTokenInfo returns the information of the token.
	// The following input arguments are required:
	//  - Token
	//
	// The returned USER_ERROR would be as below:
	//  - Code_USER_UNKNOWN_TOKEN
	GetTokenInfo(context.Context, GetTokenInfoRequest) (GetTokenInfoReply, error)

	// SignIn needs the following required input:
	//  - Name value(employee id)
	//  - Password
	// If ADUser is true, the Name value is AD account, otherwise, is MES account.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_DEPARTMENT_NOT_FOUND
	//
	// The token is expired in 8 hours in default.
	// SignIn also returns whether the user needs to change the password or not.
	SignIn(context.Context, SignInRequest, ...SignInOption) (SignInReply, error)

	// SignOut signs out.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_USER_UNKNOWN_TOKEN
	SignOut(context.Context, SignOutRequest) error

	// CreateUsers needs the following required input:
	//  - the length of the Users field is at least 1
	//  - all of elements are required:
	//    - ID
	//    - DepartmentID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_USER_ALREADY_EXISTS
	CreateUsers(context.Context, CreateUsersRequest) error

	// UpdateUser needs the following required input:
	//  - ID
	//  - one of Account, DepartmentID and LeaveDate field
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_DEPARTMENT_NOT_FOUND
	//  - Code_USER_NOT_FOUND
	UpdateUser(context.Context, UpdateUserRequest) error

	// DeleteUser needs the following required input:
	//  - ID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	DeleteUser(context.Context, DeleteUserRequest) error

	// CreateDepartments needs the following required input:
	//  - the length of the CreateDepartmentsRequest field is at least 1
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_DEPARTMENT_ALREADY_EXISTS
	CreateDepartments(context.Context, CreateDepartmentsRequest) error

	// DeleteDepartment needs the following required input:
	//  - DepartmentID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	DeleteDepartment(context.Context, DeleteDepartmentRequest) error

	// UpdateDepartment renames the department id.
	// The following input arguments are required:
	//  - OldID
	//  - NewID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_DEPARTMENT_NOT_FOUND
	UpdateDepartment(context.Context, UpdateDepartmentRequest) error

	// ListAllDepartment lists all departments.
	//
	// the reply will be ordered by department id.
	ListAllDepartment(context.Context) (ListAllDepartmentReply, error)

	// ListStations lists the stations by specified department oid.
	//
	// For slot site, if:
	//  - there is nothing in the site, all of fields of the Slot field of the SiteContent struct is zero-value.
	//  - one thing is in the site, the corresponding field of the Slot field is a non-zero-value.
	//
	// For container, collection, queue, and colqueue sites, if:
	//  - there is nothing in the site, the length of the corresponding field of the SiteContent struct is 0.
	//  - there are something in the site, the length of the corresponding field is greater than 0.
	//
	// this function is able to be paginated. 🗐
	//
	// this function is orderable with following fields:
	//  - "id" : station id
	//
	// ListStations needs the following required input:
	//  - DepartmentOID
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//
	// If there is no station, it will return a 0-length slice.
	ListStations(context.Context, ListStationsRequest) (ListStationsReply, error)

	// ListStationIDs lists station IDs according to specified department oid.
	// the reply will be ordered by id.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	ListStationIDs(context.Context, ListStationIDsRequest) (ListStationIDsReply, error)

	// GetStation returns the station by specified station id.
	//
	// GetStation needs the following required input:
	//  - ID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_NOT_FOUND
	GetStation(context.Context, GetStationRequest) (GetStationReply, error)

	// CreateStation needs the following required input:
	//  - ID
	//  - DepartmentOID
	// others are optional.
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_ALREADY_EXISTS
	// Notice that the default station state is SHUTDOWN.
	CreateStation(context.Context, CreateStationRequest) error

	// UpdateStation needs the following required input:
	//  - ID
	// others are optional.
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_NOT_FOUND
	UpdateStation(context.Context, UpdateStationRequest) error

	// DeleteStation deletes the specified station and the sites that belong to it.
	// DeleteStation needs the following required input:
	//  - StationID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_NOT_FOUND
	DeleteStation(context.Context, DeleteStationRequest) error

	// CreateStationGroup needs the following required input:
	//  - all fields in request
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_GROUP_ALREADY_EXISTS
	CreateStationGroup(context.Context, StationGroupRequest) error

	// UpdateStationGroup needs the following required input:
	//  - all fields in request
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_GROUP_ID_NOT_FOUND
	UpdateStationGroup(context.Context, StationGroupRequest) error

	// DeleteStationGroup needs the following required input:
	//  - GroupID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	DeleteStationGroup(context.Context, DeleteStationGroupRequest) error

	// SetStationConfiguration sets configs of the station.
	//
	// SetStationConfiguration needs the following required input:
	//  - StationID
	//
	// Notice that all value(contains zero value) in the request will be updated to db.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	SetStationConfiguration(context.Context, SetStationConfigurationRequest) error

	// GetStationConfiguration queries the specified station configs.
	//
	// GetStationConfiguration needs the following required input:
	//  - StationID
	//
	// this method return a empty reply while station config not found.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	GetStationConfiguration(context.Context, GetStationConfigurationRequest) (GetStationConfigurationReply, error)

	// GetSite
	//
	// GetSite needs the following required input:
	//  - StationID
	//  - SiteName
	//  - SiteIndex (0 is allowed)
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_SITE_NOT_FOUND
	GetSite(context.Context, GetSiteRequest) (GetSiteReply, error)

	// ListAssociatedStations gets associated stations according to specified site.
	//
	// This method will NOT check if the specified site exists or not
	//
	// ListAssociatedStations needs the following required input:
	//  - Site
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	ListAssociatedStations(context.Context, ListAssociatedStationsRequest) (ListAssociatedStationsReply, error)

	// SignInStation is required:
	//  - Station
	//  - Site : operator sites only (Type: Slot, Sub Type: Operator)
	//  - WorkDate
	//  - Group
	//
	// Options to set:
	//  - WithVerifyWorkDateHandler: verify if the work date is valid. Default is passed.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_USER_NOT_FOUND_OR_BAD_PASSWORD
	//  - Code_STATION_NOT_FOUND
	//  - Code_RESOURCE_NOT_FOUND
	//  - Code_BAD_WORK_DATE
	//  - Code_PREVIOUS_USER_NOT_SIGNED_OUT
	SignInStation(context.Context, SignInStationRequest, ...SignInStationOption) error

	// SignOutStation is required:
	//  - Station
	//  - Site : operator sites only (Type: Slot, Sub Type: Operator)
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_USER_NOT_FOUND_OR_BAD_PASSWORD
	//  - Code_STATION_NOT_FOUND
	//  - Code_RESOURCE_NOT_FOUND
	//  - Code_USER_HAS_NOT_SIGNED_IN
	//  - Code_STATION_OPERATOR_NOT_MATCH
	SignOutStation(context.Context, SignOutStationRequest) error

	// SignOutStations needs the following required input:
	//  - Sites
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	SignOutStations(context.Context, SignOutStationsRequest) error

	// CreateProductPlan creates a product plan for production.
	//
	// @Param: req - all fields in req are required.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_DEPARTMENT_ID_NOT_FOUND
	//  - Code_PRODUCTION_PLAN_EXISTED
	CreateProductPlan(context.Context, CreateProductionPlanRequest) error

	// ListProductPlans returns specified product type production planning on.
	// the specified date.
	// ListProductPlansReply.ProductPlans are ordered by product id.
	//
	// ListProductPlans needs the following required input:
	//  - DepartmentOID
	//  - Date
	//  - ProductType
	ListProductPlans(context.Context, ListProductPlansRequest) (ListProductPlansReply, error)

	// IsProductExisted needs the following required input:
	//  - productID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_PRODUCT_ID_NOT_FOUND
	IsProductExisted(ctx context.Context, productID string) (bool, error)

	// ListFeedRecords only lists the materials which have been fed.
	// Information about operators and time spent is not in the reply.
	//
	// ListFeedRecords needs the following required input:
	//  - Date
	//  - DepartmentID
	// others are optional.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	ListFeedRecords(context.Context, ListRecordsRequest) (ListFeedRecordReply, error)

	// GetCollectRecord needs the following required input:
	//  - all field in GetCollectRecordRequest
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_RECORD_NOT_FOUND
	GetCollectRecord(context.Context, GetCollectRecordRequest) (GetCollectRecordReply, error)

	// ListCollectRecords needs the following required input:
	//  - Date
	//  - DepartmentID
	// others are optional.
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	// the reply will be sorted by product id.
	ListCollectRecords(context.Context, ListRecordsRequest) (ListCollectRecordsReply, error)

	// CreateCollectRecord needs the following required input:
	//  -  WorkOrder
	//  -  Sequence
	//  -  LotNumber
	//  -  Station
	//  -  ResourceOID
	//  -  Quantity must greater than 0
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_RECORD_ALREADY_EXISTS
	CreateCollectRecord(context.Context, CreateCollectRecordRequest) error

	// CreateWorkOrders needs the following required input:
	//  - the length of the CreateWorkOrdersRequest field is at least 1
	//  - all of elements are required:
	//    - ProcessOID
	//    - RecipeID
	//    - DepartmentOID
	//    - Date
	//    - BatchesQuantity
	//
	// others are optional.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_PROCESS_NOT_FOUND
	CreateWorkOrders(context.Context, CreateWorkOrdersRequest) (CreateWorkOrdersReply, error)

	// UpdateWorkOrders modifies the work order information, excluding ID(for.
	// update condition)
	//
	// Notice that this function will not update zero values like empty string or 0 integer
	//
	// UpdateWorkOrders needs the following required input:
	//  - ID
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	UpdateWorkOrders(context.Context, UpdateWorkOrdersRequest) error

	// Deprecated: ListWorkOrders : Use ListWorkOrdersByDuration instead.
	//
	// ListWorkOrders gets work orders in ascending order of date and sequence.
	// the reply will be ordered by reserved sequence.
	//
	// ListWorkOrders needs the following required input:
	//  - Date
	//  - Station
	// ID is optional.
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	ListWorkOrders(context.Context, ListWorkOrdersRequest) (ListWorkOrdersReply, error)

	// ListWorkOrdersByDuration lists work orders in ascending order of date and sequence.
	//
	// ListWorkOrdersByDuration needs the following required input:
	//  - Since
	//  - one of Station or DepartmentID
	// ID is optional.
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_BAD_REQUEST
	ListWorkOrdersByDuration(context.Context, ListWorkOrdersByDurationRequest) (ListWorkOrdersByDurationReply, error)

	// ListWorkOrdersByIDs lists work orders according to given ids and station.
	//
	// ListWorkOrdersByIDs needs the following required input:
	//  - IDs
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	ListWorkOrdersByIDs(context.Context, ListWorkOrdersByIDsRequest) (ListWorkOrdersByIDsReply, error)

	// GetWorkOrder returns work order data by specified ID.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_WORKORDER_NOT_FOUND
	GetWorkOrder(context.Context, GetWorkOrderRequest) (GetWorkOrderReply, error)

	// CreateRecipes needs the following required input:
	//  - all field in CreateRecipesRequest
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_RECIPE_ALREADY_EXISTS
	CreateRecipes(context.Context, CreateRecipesRequest) error

	// DeleteRecipe needs the following required input:
	//  - the length of the DeleteRecipeRequest field is at least 1
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	DeleteRecipe(context.Context, DeleteRecipeRequest) error

	// ListRecipesByProduct returns recipes with specified productID.
	//
	// ListRecipesByProduct needs the following required input:
	//  - ProductID
	// other are optional.
	//
	// this function is orderable with following fields:
	//  - "id"
	//  - "product_id"
	//  - "released_at"
	//  - "stage"
	//  - "major"
	//  - "minor"
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	ListRecipesByProduct(context.Context, ListRecipesByProductRequest) (ListRecipesByProductReply, error)

	// GetRecipe returns recipe with all processes.
	//
	// use WithoutProcesses() if you don't need processes.
	//
	// GetRecipe needs the following required input:
	//  - ID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_RECIPE_NOT_FOUND
	GetRecipe(context.Context, GetRecipeRequest) (GetRecipeReply, error)

	// GetProcessDefinition returns a recipe process definition according to the specified conditions.
	//
	// All the fields in the request are required.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_PROCESS_NOT_FOUND
	GetProcessDefinition(context.Context, GetProcessDefinitionRequest) (GetProcessDefinitionReply, error)

	// ListSubstitutions lists substitutions by given product ID and grade.
	// the replied substitutions are ordered by name then by grade.
	// required arguments:
	//  - ID
	//  - Grade
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	ListSubstitutions(ctx context.Context, req ListSubstitutionsRequest) (ListSubstitutionsReply, error)

	// ListMultipleSubstitutions lists substitutions by given product set.
	// Each set in the reply will be ordered by name then by grade.
	//
	// A material without any substitutions will return an empty slice.
	ListMultipleSubstitutions(ctx context.Context, req ListMultipleSubstitutionsRequest) (ListMultipleSubstitutionsReply, error)

	// DeleteSubstitutions Removes specified substitutions.
	// required arguments:
	//  - ID
	//  - Grade
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	DeleteSubstitutions(ctx context.Context, req DeleteSubstitutionsRequest) error

	// UpdateSubstitutions set the substitutions with the specified contents.
	// If there were no substitutions related to the target product, the method will create new data.
	//
	// required arguments:
	//  - ID
	//  - Grade
	//  - Contents
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	UpdateSubstitutions(ctx context.Context, req BasicSubstitutionRequest) error

	// AddSubstitutions adds new substitutions.
	// required arguments:
	//  - ID
	//  - Grade
	//  - Contents
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_SUBSTITUTION_ALREADY_EXISTS
	AddSubstitutions(ctx context.Context, req BasicSubstitutionRequest) error

	// ListProductTypes returns all of the product types that the department is in charge of.
	// If DepartmentOID == "", it returns all of product types exist in recipe table.
	// The replied types will be ordered by type name.
	ListProductTypes(context.Context, ListProductTypesRequest) (ListProductTypesReply, error)

	// ListProductIDs needs the following required input:
	//  - Type
	// If IsLastProcess is true, returns all last stage productIDs.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	// The replied product ids will be ordered by product id.
	ListProductIDs(context.Context, ListProductIDsRequest) (ListProductIDsReply, error)

	// ListProductGroups needs the following required input:
	//  - DepartmentOID
	//  - Type
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	// the replied product groups will be ordered by id.
	ListProductGroups(context.Context, ListProductGroupsRequest) (ListProductGroupsReply, error)

	// GetToolResource returns a tool according to specified resource id.
	//
	// GetToolResource needs the following required input:
	//  - resource id
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_RESOURCE_NOT_FOUND
	GetToolResource(context.Context, GetToolResourceRequest) (GetToolResourceReply, error)

	// ListToolResources lists tool Resources.
	//
	// ListToolResources needs the following required input:
	//  - ResourcesID
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	ListToolResources(context.Context, ListToolResourcesRequest) (ListToolResourcesReply, error)

	// ToolResourceBind.
	//
	// ToolResourceBind needs the following required input:
	// - Details:
	//    - Type: should be slot bind or slot clear
	//    - Site
	//    - ResourceID: is required if bind type is slot bind.
	//
	// About ResourceID field:
	//  - Do limit checks before calling this function.
	//  - Make sure that the resource is not in any site before calling this function.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_NOT_FOUND
	//  - Code_STATION_SITE_NOT_FOUND
	//  - Code_RESOURCE_NOT_FOUND
	//
	// Deprecated: use V2 instead
	ToolResourceBind(context.Context, ToolResourceBindRequest) error

	// ToolResourceBindV2
	//
	// ToolResourceBindV2 needs the following required input:
	// - Details:
	//    - Type: should be slot bind or slot clear
	//    - Site
	//    - ResourceID: is required if bind type is slot bind.
	//
	// About ResourceID field:
	//  - Do limit checks before calling this function.
	//  - Make sure that the resource is not in any site before calling this function.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_NOT_FOUND
	//  - Code_STATION_SITE_NOT_FOUND
	//  - Code_RESOURCE_NOT_FOUND
	ToolResourceBindV2(context.Context, ToolResourceBindRequestV2) error

	// MaterialResourceBind:
	// MaterialResourceBindRequest needs the following required input:
	//  - Station : use "" to specify a shared site
	// MaterialResourceBindRequest.Details needs the following required input:
	//  - Type
	//  - Site
	//  - Resources field is required in some of the bind types such as bind, add, and push.
	//  - Option field is required in Queue / ColQueue sites.
	// MaterialResourceBindRequest.Details.Resources needs the following required input:
	//  - Quantity: must be greater than zero, except for UnspecifiedQuantity == true.
	//  - ResourceID
	//
	// About UnspecifiedQuantity:
	//  - this field can only be used in following bind types:
	//     - slot bind
	//     - queue bind
	//     - queue push
	//     - queue pushpop
	//  - The quantity of the resource will not be mutated while binding if you set this property is true.
	//
	// Notice that the method can only work on material sites.
	//
	// Notice that all the information in the request will be updated to the database directly.
	// An incomplete request or incorrect request may cause unexpected situations.
	// The returned error code would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_BAD_REQUEST
	//
	// Deprecated: use V2 instead
	MaterialResourceBind(context.Context, MaterialResourceBindRequest) error

	// MaterialResourceBindV2:
	// MaterialResourceBindRequestV2 needs the following required input:
	//  - Details
	// MaterialResourceBindRequestV2.Details needs the following required input:
	//  - Type
	//  - Site
	//  - Resources field is required in some of the bind types such as bind, add, and push.
	//  - Option field is required in Queue / ColQueue sites.
	// MaterialResourceBindRequestV2.Details.Resources needs the following required input:
	//  - Quantity: must be greater than zero, except for UnspecifiedQuantity == true.
	//  - ResourceID
	//
	// About UnspecifiedQuantity:
	//  - this field can only be used in following bind types:
	//     - slot bind
	//     - queue bind
	//     - queue push
	//     - queue pushpop
	//  - The quantity of the resource will not be mutated while binding if you set this property is true.
	//
	// Notice that the method can only work on material sites.
	//
	// Notice that all the information in the request will be updated to the database directly.
	// An incomplete request or incorrect request may cause unexpected situations.
	// The returned error code would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_BAD_REQUEST
	//  - Code_INVALID_NUMBER
	//  - Code_STATION_SITE_NOT_FOUND
	MaterialResourceBindV2(context.Context, MaterialResourceBindRequestV2) error

	// BindRecordsCheck checks if the specified materials were ever bound in the specified site.
	//
	// All fields in the request are required.
	//
	// Notice that this method only supports material sites.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_SITE_BIND_RECORD_NOT_FOUND
	BindRecordsCheck(context.Context, BindRecordsCheckRequest) error

	// ListSiteMaterials needs the following required input:
	//  - Station
	//  - Site.Name
	//  - Site.Index
	// The returned error code would be as below:
	//  - Code_RESOURCE_NOT_FOUND
	//  - Code_STATION_NOT_FOUND
	//  - Code_INSUFFICIENT_REQUEST
	// The replied materials will be ordered by id.
	ListSiteMaterials(context.Context, ListSiteMaterialsRequest) (ListSiteMaterialsReply, error)

	// CreateMaterialResources create resources in which you can create multiple material resources.
	// Requirement:
	//  - a slice of Material : the length of it should be at least 1
	//     - if a ResourceID is not specified, it will be generated automatically.
	//     - the quantity and the planned quantity should be >= 0, and not simultaneously equal to 0.
	// It also offers stock in resource into warehouse as OPTIONAL request.
	//
	// It will return a list of ResourceID which have been created.
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_RESOURCE_EXISTED
	//  - Code_INVALID_NUMBER
	CreateMaterialResources(ctx context.Context, req CreateMaterialResourcesRequest, opts ...CreateMaterialResourcesOption) (CreateMaterialResourcesReply, error)

	// WarehousingStock updates warehouse stock transactions and records into a corresponding warehouse & location.
	// Notice that this method does not check whether the warehouse exists or not.
	//
	// Requirement:
	//  - Warehouse (all fields are required)
	//  - ResourceIDs (at least a data is provided)
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_RESOURCE_NOT_FOUND
	WarehousingStock(ctx context.Context, req WarehousingStockRequest) error

	// GetMaterialResource queries warehouse data of the corresponding resource.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_RESOURCE_NOT_FOUND
	GetMaterialResource(context.Context, GetMaterialResourceRequest) (GetMaterialResourceReply, error)

	// GetMaterialResourceIdentity queries the material by the specified resource id and product type.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_RESOURCE_NOT_FOUND
	GetMaterialResourceIdentity(context.Context, GetMaterialResourceIdentityRequest) (GetMaterialResourceIdentityReply, error)

	// ListMaterialResourceIdentities queries the materials by the specified resources id and product type.
	// where the order of reply will be the same as the request. (resources not found will be shown as nil)
	ListMaterialResourceIdentities(context.Context, ListMaterialResourceIdentitiesRequest) (ListMaterialResourceIdentitiesReply, error)

	// ListMaterialResources queries the materials by the specified conditions.
	//
	// this function is able to be paginated. 🗐
	//
	// this function is orderable with following fields:
	//  - "created_at"
	ListMaterialResources(context.Context, ListMaterialResourcesRequest) (ListMaterialResourcesReply, error)

	// ListMaterialResourcesById lists resources by specified resources ID.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	ListMaterialResourcesById(context.Context, ListMaterialResourcesByIdRequest) (ListMaterialResourcesByIdReply, error)

	// SplitMaterialResource needs the following required input:
	//  - ResourceID
	//  - ProductType
	//  - Quantity
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_INVALID_NUMBER
	//  - Code_RESOURCE_NOT_FOUND
	SplitMaterialResource(context.Context, SplitMaterialResourceRequest) (SplitMaterialResourceReply, error)

	// GetResourceWarehouse queries warehouse data of the corresponding resource.
	// The following input arguments are required:
	//  - ResourceID
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_WAREHOUSE_RESOURCE_NOT_FOUND
	GetResourceWarehouse(context.Context, GetResourceWarehouseRequest) (GetResourceWarehouseReply, error)

	// CreateBatch needs the following required input:
	//  - WorkOrder
	//  - Number
	//
	// the default status of batch is preparing.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_BATCH_ALREADY_EXISTS
	CreateBatch(context.Context, CreateBatchRequest) error

	// ListBatches needs the following required input:
	//  - WorkOrder
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//
	// the reply of this method will be ordered by "number".
	ListBatches(context.Context, ListBatchesRequest) (ListBatchesReply, error)

	// GetBatch needs the following required input:
	//  - WorkOrder
	//  - Number
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_BATCH_NOT_FOUND
	GetBatch(context.Context, GetBatchRequest) (GetBatchReply, error)

	// UpdateBatch needs the following required input:
	//  - WorkOrder
	//  - Number
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_BATCH_NOT_FOUND
	UpdateBatch(context.Context, UpdateBatchRequest) error

	// Feed feeds materials on the specified sites and save the records.
	// Feed needs the following required input:
	//  - FeedContent: at least one.
	//
	// More about FeedContent:
	//  - There are three ways to feed materials.
	//  - FeedPerSiteType1:
	//		- feeds materials in the specified site.
	//		- sets the specified site to the feed record.
	//  - FeedPerSiteType2:
	//		- feeds the materials directly.
	//		- sets the specified site to the feed record.
	//		- does NOT update the quantity of the materials which could not be found in the mes system.
	//  - FeedPerSiteType3:
	//		- feeds the materials directly.
	//		- creates a feed record with an empty site.
	//		- does NOT update the quantity of the materials which could not be found in the mes system.
	//
	// FeedPerSiteType1 needs the following required input:
	//  - Site
	// One of following properties in FeedPerSiteType1 is required:
	//  - FeedAll: default false. for collection or colqueue sites, the property must be true.
	//  - Quantity: must be greater than zero if not FeedAll.
	//
	// FeedPerSiteType2 needs the following required input:
	//  - Site
	//  - Quantity
	//  - ResourceID
	//  - ProductType
	//
	// FeedPerSiteType3 needs the following required input:
	//  - Quantity
	//  - ResourceID
	//  - ProductType
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_STATION_SITE_NOT_FOUND
	Feed(context.Context, FeedRequest) (FeedReply, error)

	// ListUnauthorizedUsers needs the following required input:
	//  - departmentOID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_DEPARTMENT_ID_NOT_FOUND
	// Add options to exclude specified users:
	//  - ExcludeUsers(users []string)
	// The reply of the method will be ordered.
	ListUnauthorizedUsers(ctx context.Context, departmentOID ListUnauthorizedUsersRequest, opts ...ListUnauthorizedUsersOption) (ListUnauthorizedUsersReply, error)

	// ListCarriers lists the carriers in the specified department.
	//
	// ListCarriers needs the following required input:
	//  - DepartmentOID
	//
	// this function is able to be paginated. 🗐
	//
	// this function is orderable with following fields:
	//  - "id_prefix"
	//  - "serial_number"
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	ListCarriers(context.Context, ListCarriersRequest) (ListCarriersReply, error)

	// GetCarrier returns specified carrier information.
	//
	// GetCarrier needs the following required input:
	//  - ID
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_CARRIER_NOT_FOUND
	GetCarrier(context.Context, GetCarrierRequest) (GetCarrierReply, error)

	// CreateCarrier needs the following required input:
	//  - DepartmentOID
	//  - IdPrefix : 2 characters
	//  - Quantity
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_CARRIER_QUANTITY_LIMIT
	CreateCarrier(ctx context.Context, req CreateCarrierRequest) error

	// UpdateCarrier needs the following required input:
	//  - ID
	//  - Action : DO NOT pass a pointer to this field
	//
	// where UpdateCarrierAction contains following structs.
	//  - UpdateProperties : update allowed materials or department oid
	// (DepartmentOID == "" means "do not update", AllowMaterials == "" means "remove all limitations")
	//  - BindResources : append resources to contents
	//  - RemoveResources : remove specified resources from contents
	//  - ClearResources : remove all resources from contents
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_CARRIER_NOT_FOUND
	UpdateCarrier(ctx context.Context, req UpdateCarrierRequest) error

	// DeleteCarrier needs the following required input:
	//  - ID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_CARRIER_NOT_FOUND
	DeleteCarrier(context.Context, DeleteCarrierRequest) error

	// ListUserRoles needs the following required input:
	//  - DepartmentOID
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_DEPARTMENT_NOT_FOUND
	// The reply of this method will be ordered by account id.
	ListUserRoles(context.Context, ListUserRolesRequest) (ListUserRolesReply, error)

	// ListRoles returns Roles are relative to gitlab.kenda.com.tw/kenda/mcom/utils/roles enumerations.
	// in ascending order of value.
	ListRoles(context.Context) (ListRolesReply, error)

	// CreateAccounts creates new accounts with specified roles and default passwords.
	//
	// The default password is equivalent to the account or it can be specified.
	// by the optional parameter WithDefaultPasswords (more details in the comments of the method).
	//
	// The request is a CreateAccountRequest slice where the following fields are required.
	//  - ID
	//  - Roles
	//  - password => must be specified by following methods
	//     - WithDefaultPassword()
	//     - WithSpecifiedPassword(pwd string)
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_ACCOUNT_ROLES_NOT_PERMIT
	//  - Code_ACCOUNT_ALREADY_EXISTS
	CreateAccounts(context.Context, CreateAccountsRequest) error

	// UpdateAccount updates the information of account according to your request.
	//
	// UpdateAccount needs the following required arguments if you want to change the password of the account:
	//  - UserID
	//  - NewPassword
	//  - OldPassword
	// UpdateAccount needs the following required arguments if you want to reset the password of the account:
	//  - UserID
	//  - Optional Parameter ResetPassword()
	// UpdateAccount needs the following required arguments if you want to update the roles of the account:
	//  - UserID
	//  - Roles (not empty and not nil)
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_ACCOUNT_ROLES_NOT_PERMIT
	//  - Code_ACCOUNT_BAD_OLD_PASSWORD
	//  - Code_ACCOUNT_SAME_AS_OLD_PASSWORD
	//  - Code_ACCOUNT_NOT_FOUND
	UpdateAccount(context.Context, UpdateAccountRequest, ...UpdateAccountOption) error

	// DeleteAccount deletes the specified account.
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - CODE_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD
	//  - Code_ACCOUNT_NOT_FOUND
	DeleteAccount(context.Context, DeleteAccountRequest) error

	// CreatePackRecords creates the specified pack records.
	CreatePackRecords(context.Context, CreatePackRecordsRequest) error

	// ListPackRecords lists all pack records.
	// the reply will be ordered by serial numbers.
	ListPackRecords(context.Context) (ListPackRecordsReply, error)

	// ListSiteType return site types are relative to gitlab.kenda.com.tw/kenda/mcom/utils/sites enumerations.
	// in ascending order of value.
	ListSiteType(context.Context) (ListSiteTypeReply, error)

	// ListSiteSubType return site sub types are relative to gitlab.kenda.com.tw/kenda/mcom/utils/sites enumerations.
	// in ascending order of value.
	ListSiteSubType(context.Context) (ListSiteSubTypeReply, error)

	// ListStationState returns station states are relative to gitlab.kenda.com.tw/kenda/mcom/utils/sites enumerations.
	// in ascending order of value.
	ListStationState(context.Context) (ListStationStateReply, error)

	// ListMaterialResourceStatus returns resource material status are relative to gitlab.kenda.com.tw/kenda/mcom/utils/resources enumerations.
	// in ascending order of value.
	ListMaterialResourceStatus(context.Context) (ListMaterialResourceStatusReply, error)

	// Create LimitaryHour.
	// CreateLimitaryHour needs the following required input:
	//  - all field in CreateLimitaryHoursRequest
	//
	// For more details about LimitaryHour, please refer to the comments of models.LimitaryHour.
	//
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_LIMITARY_HOUR_NOT_FOUND
	//  - Code_LIMITARY_HOUR_ALREADY_EXISTS
	CreateLimitaryHour(context.Context, CreateLimitaryHourRequest) error

	// GetLimitaryHour needs the following required input:
	//  - ProductType
	// The returned USER_ERROR would be as below:
	//  - Code_INSUFFICIENT_REQUEST
	//  - Code_LIMITARY_HOUR_NOT_FOUND
	GetLimitaryHour(context.Context, GetLimitaryHourRequest) (GetLimitaryHourReply, error)
}
