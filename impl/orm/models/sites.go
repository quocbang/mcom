package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// SiteID definition.
type SiteID struct {
	// Name is relative to Site.Name.
	Name string `json:"name"`
	// Index is relative to Site.Index.
	Index int16 `json:"index"`
}

// Site is the place where things put in or operators work at. e.g. operator
// site, material site, tool site, machine site etc..
type Site struct {
	Name  string `gorm:"type:varchar(16);primaryKey"`
	Index int16  `gorm:"default:0;primaryKey"`

	// The value is relative to Station.ID.
	Station string `gorm:"type:varchar(32);primaryKey"`

	// AdminDepartmentID is the department ID to be charge of this site.
	// The value is relative to Department.ID.
	AdminDepartmentID string `gorm:"column:admin_department_id;type:text;not null;index:idx_site_department"`

	Attributes SiteAttributes `gorm:"type:jsonb;default:'{}';not null"`

	// UpdatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	UpdatedAt types.TimeNano `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy string         `gorm:"type:text;not null"`
	// CreatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	CreatedAt types.TimeNano `gorm:"autoCreateTime:nano;not null"`
	CreatedBy string         `gorm:"type:text;not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (Site) TableName() string {
	return "site"
}

// SiteAttributes definition.
type SiteAttributes struct {
	Type    sites.Type    `json:"type"`
	SubType sites.SubType `json:"sub_type"`
	// 限制 ProductID 。
	// mcom 不會在掛載行為檢查這個項目。
	// 沒有限制的情況下其值將為 empty Slice 。
	Limitation []string `json:"limitation"`
}

// Scan implements database/sql Scanner interface.
func (c *SiteAttributes) Scan(src interface{}) error {
	return ScanJSON(src, c)
}

// Value implements database/sql/driver Valuer interface.
func (c SiteAttributes) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// SiteContents Definition.
//
// Check the type of the site from Site.Attributes.Type.
//  - TypeUnspecified
//  - TypeContainer
//  - TypeSlot
//  - TypeCollection
//  - TypeQueue
//  - TypeColqueue
type SiteContents struct {
	// Name is relative to Site.Name.
	Name string `gorm:"type:varchar(16);primaryKey"`
	// Index is relative to Site.Index.
	Index int16 `gorm:"default:0;primaryKey"`
	// Station is relative to Site.Station.
	Station string      `gorm:"type:varchar(32);primaryKey"`
	Content SiteContent `gorm:"type:jsonb;default:'{}';not null"`

	// UpdatedAt the number of nanoseconds elapsed since January 1, 1970 UTC.
	UpdatedAt types.TimeNano `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy string         `gorm:"type:text;not null"`
}

// avoid nil pointer panic.
func (sc *SiteContents) AfterFind(tx *gorm.DB) error {
	if sc.Content.Collection != nil {
		for i, res := range *sc.Content.Collection {
			if res.Material != nil && res.Material.Quantity == nil {
				(*sc.Content.Collection)[i].Material.Quantity = types.Decimal.NewFromInt32(0)
			}

		}
	}
	if sc.Content.Colqueue != nil {
		for i := range *sc.Content.Colqueue {
			for j, res := range (*sc.Content.Colqueue)[i] {
				if res.Material != nil && res.Material.Quantity == nil {
					(*sc.Content.Colqueue)[i][j].Material.Quantity = types.Decimal.NewFromInt32(0)
				}
			}
		}
	}
	if sc.Content.Container != nil {
		for i, res := range *sc.Content.Container {
			if res.Material != nil && res.Material.Quantity == nil {
				(*sc.Content.Container)[i].Material.Quantity = types.Decimal.NewFromInt32(0)
			}
		}
	}
	return nil
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (SiteContents) TableName() string {
	return "site_contents"
}

// SiteContent defines all supported sites.
type SiteContent struct {
	Slot       *Slot       `json:"slot,omitempty"`
	Container  *Container  `json:"container,omitempty"`
	Collection *Collection `json:"collection,omitempty"`
	Queue      *Queue      `json:"queue,omitempty"`
	Colqueue   *Colqueue   `json:"colqueue,omitempty"`
}

func NewContainerSiteContent() SiteContent {
	return SiteContent{Container: &Container{}}
}
func NewSlotSiteContent() SiteContent {
	return SiteContent{Slot: &Slot{}}
}
func NewCollectionSiteContent() SiteContent {
	return SiteContent{Collection: &Collection{}}
}
func NewQueueSiteContent() SiteContent {
	return SiteContent{Queue: &Queue{}}
}
func NewColqueueSiteContent() SiteContent {
	return SiteContent{Colqueue: &Colqueue{}}
}

// Scan implements database/sql Scanner interface.
func (c *SiteContent) Scan(src interface{}) error {
	return ScanJSON(src, c)
}

// Value implements database/sql/driver Valuer interface.
func (c SiteContent) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// OperatorSite definition.
type OperatorSite struct {
	EmployeeID string    `json:"employee_id"`
	Group      int8      `json:"group"`
	WorkDate   time.Time `json:"work_date"`
}

// Current returns the current operator in the site.
// If the EmployeeID == "", it means there is no one in the site.
// It is still considered someone in the sit even if after ExpiryTime.
func (os *OperatorSite) Current() OperatorSite {
	if os == nil {
		return OperatorSite{}
	}
	return *os
}

func (os *OperatorSite) AllowedLogin(userID string) bool {
	current := os.Current()
	return current.EmployeeID == "" || current.EmployeeID == userID
}

// ToolSite definition.
type ToolSite struct {
	ResourceID    string    `json:"resource_id"`
	ToolID        string    `json:"tool_id"`
	InstalledTime time.Time `json:"installed_time"`
}

// MaterialSite definition.
type MaterialSite struct {
	Material    Material                 `json:"material"`
	Quantity    *decimal.Decimal         `json:"quantity"`
	ResourceID  string                   `json:"resource_id"`
	ProductType string                   `json:"product_type"`
	Status      resources.MaterialStatus `json:"status"`
	// ExpiryTime is expiry time of the MaterialSite. The number of nanoseconds.
	// elapsed since January 1, 1970 UTC.
	ExpiryTime types.TimeNano `json:"expiry_time"`
}

// Material definition.
type Material struct {
	ID    string `json:"id"`
	Grade string `json:"grade"`
}

// BoundResource is the material bound in the site.
type BoundResource struct {
	Material *MaterialSite `json:"material,omitempty"`
	Tool     *ToolSite     `json:"tool,omitempty"`
	Operator *OperatorSite `json:"operator,omitempty"`
}

// Container is a non-discrete material site.
type Container []BoundResource

func (container *Container) CleanDeviation() {
	*container = Container{}
}

func (container *Container) Clear() []BoundResource {
	res := *container
	*container = Container{}
	return res
}

func (container *Container) Add(resources []BoundResource) {
	*container = append(*container, resources...)
}

func (container *Container) Bind(resources []BoundResource) []BoundResource {
	res := container.Clear()
	container.Add(resources)
	return res
}

// Feed feeds the specified quantity from the site and returns fed materials.
// If the material is NOT enough to feed, the returned value shortage is true.
func (container *Container) Feed(quantity decimal.Decimal) (materials []FedMaterial, shortage bool) {
	if len(*container) == 0 {
		return []FedMaterial{}, true
	}

	var counter int
	for i, resource := range *container {
		if resource.Material.Quantity.GreaterThan(quantity) {
			(*container)[i].Material.Quantity = types.Decimal.NewFromDecimal(resource.Material.Quantity.Sub(quantity))
			materials = append(materials, FedMaterial{
				Material:    resource.Material.Material,
				ResourceID:  resource.Material.ResourceID,
				ProductType: resource.Material.ProductType,
				Status:      resource.Material.Status,
				ExpiryTime:  resource.Material.ExpiryTime.Time(),
				Quantity:    quantity,
			})
			quantity = decimal.Zero
			break
		}
		quantity = quantity.Sub(*resource.Material.Quantity)
		materials = append(materials, FedMaterial{
			Material:    resource.Material.Material,
			ResourceID:  resource.Material.ResourceID,
			ProductType: resource.Material.ProductType,
			Status:      resource.Material.Status,
			ExpiryTime:  resource.Material.ExpiryTime.Time(),
			Quantity:    *resource.Material.Quantity,
		})
		counter++
	}
	*container = (*container)[counter:]

	if quantity.GreaterThan(decimal.Zero) {
		// ! 投入數量多於剩餘量
		materials[len(materials)-1].Quantity = materials[len(materials)-1].Quantity.Add(quantity)
		shortage = true
	}

	return materials, shortage
}

func (container *Container) FeedAll() []FedMaterial {
	records := make([]FedMaterial, len(*container))
	for i, res := range *container {
		records[i] = FedMaterial{
			Material:    res.Material.Material,
			ResourceID:  res.Material.ResourceID,
			ProductType: res.Material.ProductType,
			Status:      res.Material.Status,
			ExpiryTime:  res.Material.ExpiryTime.Time(),
			Quantity:    *res.Material.Quantity,
		}
	}
	container.Clear()
	return records
}

type FedMaterial struct {
	Material    Material
	ResourceID  string
	ProductType string
	Status      resources.MaterialStatus
	ExpiryTime  time.Time
	Quantity    decimal.Decimal
}

// Slot is a 1-1 resource site.
type Slot BoundResource

func (slot *Slot) Clear() BoundResource {
	res := *slot
	*slot = Slot{}
	return BoundResource(res)
}

func (slot *Slot) Bind(resource BoundResource) BoundResource {
	res := slot.Clear()
	*slot = Slot(resource)
	return res
}

// Feed feeds the specified quantity from the site and returns fed materials.
// If the material is NOT enough to feed, the returned value shortage is true.
// If reduceQuantityFromWarehouse is true, than the shortage field will be false.
func (slot *Slot) Feed(quantity decimal.Decimal) (materials []FedMaterial, shortage bool, reduceQuantityFromWarehouse bool) {
	if slot.Material == nil {
		return []FedMaterial{}, true, reduceQuantityFromWarehouse
	}

	materials = []FedMaterial{{
		Material:    slot.Material.Material,
		ResourceID:  slot.Material.ResourceID,
		ProductType: slot.Material.ProductType,
		Status:      slot.Material.Status,
		ExpiryTime:  slot.Material.ExpiryTime.Time(),
		Quantity:    quantity,
	}}

	if slot.Material.Quantity == nil {
		reduceQuantityFromWarehouse = true
		return
	}

	if slot.Material.Quantity.GreaterThan(quantity) {
		slot.Material.Quantity = types.Decimal.NewFromDecimal(slot.Material.Quantity.Sub(quantity))
	} else {
		// ! 投入數量多於剩餘量
		if slot.Material.Quantity.LessThan(quantity) {
			shortage = true
		}

		slot.Clear()
	}
	return materials, shortage, reduceQuantityFromWarehouse
}

// for feed method.
type MaterialsWithoutQuantity struct {
	ResourceID  string
	ProductType string
	FedQuantity decimal.Decimal
}

func (mwq MaterialsWithoutQuantity) IsEmpty() bool {
	return mwq == MaterialsWithoutQuantity{}
}

func (slot *Slot) FeedAll() []FedMaterial {
	if slot.Material == nil {
		return []FedMaterial{}
	}

	materials, _, _ := slot.Feed(*slot.Material.Quantity)
	return materials
}

// Collection is the site for collection.
type Collection []BoundResource

func (collection *Collection) Clear() []BoundResource {
	res := *collection
	*collection = Collection{}
	return res
}

func (collection *Collection) Add(resources []BoundResource) {
	*collection = append(*collection, resources...)
}

func (collection *Collection) Bind(resources []BoundResource) []BoundResource {
	res := collection.Clear()
	collection.Add(resources)
	return res
}

func (collection *Collection) FeedAll() []FedMaterial {
	res := make([]FedMaterial, len(*collection))
	for i, elem := range *collection {
		res[i] = FedMaterial{
			Material:    elem.Material.Material,
			ResourceID:  elem.Material.ResourceID,
			ProductType: elem.Material.ProductType,
			Status:      elem.Material.Status,
			ExpiryTime:  elem.Material.ExpiryTime.Time(),
			Quantity:    *elem.Material.Quantity,
		}
	}
	collection.Clear()
	return res
}

// Queue is the site for queue.
type Queue []Slot

func (queue *Queue) Bind(index uint16, resource BoundResource) BoundResource {
	queue.maybeExtend(index)
	return (*queue)[index].Bind(resource)
}

func (queue *Queue) Push(resource BoundResource) {
	*queue = append(*queue, Slot(resource))
}

func (queue *Queue) Pop() BoundResource {
	res := (*queue)[0]
	*queue = (*queue)[1:]
	return BoundResource(res)
}

func (queue *Queue) PushPop(resource BoundResource) BoundResource {
	queue.Push(resource)
	return queue.Pop()
}

func (queue *Queue) Remove(index uint16) BoundResource {
	// out of index : do nothing.
	if len(*queue)-1 < int(index) {
		return BoundResource{}
	}
	return (*queue)[index].Clear()
}

func (queue *Queue) Clear() []BoundResource {
	res := []BoundResource{}
	for _, elem := range *queue {
		if elem.Material != nil {
			res = append(res, BoundResource(elem))
		}
	}
	*queue = Queue{}
	return res
}

func (queue *Queue) ParseIndex(opt BindOption) uint16 {
	switch {
	case opt.Head:
		return 0
	case opt.Tail:
		return uint16(len(*queue) - 1)
	default:
		return *opt.QueueIndex
	}
}

func (queue *Queue) maybeExtend(maxIndex uint16) {
	if len(*queue)-1 < int(maxIndex) {
		newQueue := make(Queue, maxIndex+1)
		copy(newQueue, *queue)

		for i := len(*queue); i < int(maxIndex+1); i++ {
			newQueue[i] = Slot{}
		}
		*queue = newQueue
	}
}

// Feed feeds materials at the first-none-empty index in the site.
// It ONLY returns Code_RESOURCE_MATERIAL_SHORTAGE if there is no material
// in the site.
// If reduceQuantityFromWarehouse is true, than the shortage field will be false.
func (queue *Queue) Feed(quantity decimal.Decimal) (records []FedMaterial, shortage bool, reduceQuantityFromWarehouse bool, err error) {
	for {
		if len(*queue) == 0 {
			return nil, false, false, mcomErr.Error{Code: mcomErr.Code_RESOURCE_MATERIAL_SHORTAGE}
		}
		if (*queue)[0].Material != nil {
			break
		}
		queue.Pop()
	}
	rec, shortage, reduceQuantityFromWarehouse := (*queue)[0].Feed(quantity)
	if shortage {
		queue.Pop()
	}
	return rec, shortage, reduceQuantityFromWarehouse, nil
}

// FeedAll feeds materials at the first-none-empty index in the site.
// It ONLY returns Code_RESOURCE_MATERIAL_SHORTAGE if there is no material
// in the site.
func (queue *Queue) FeedAll() ([]FedMaterial, error) {
	var slot Slot
	for {
		if len(*queue) == 0 {
			return nil, mcomErr.Error{Code: mcomErr.Code_RESOURCE_MATERIAL_SHORTAGE}
		}
		slot = Slot(queue.Pop())
		if slot.Material != nil {
			break
		}
	}

	return slot.FeedAll(), nil
}

// Colqueue is the site for colqueue.
type Colqueue []Collection

func (colqueue *Colqueue) Bind(index uint16, resources []BoundResource) []BoundResource {
	newColqueue := *colqueue
	newColqueue.maybeExtend(index)
	res := newColqueue[index].Bind(resources)
	*colqueue = newColqueue
	return res
}

func (colqueue *Colqueue) Add(index uint16, resources []BoundResource) {
	newColqueue := *colqueue
	newColqueue.maybeExtend(index)
	newColqueue[index].Add(resources)
	*colqueue = newColqueue
}

func (colqueue *Colqueue) Clear() []BoundResource {
	returnValue := []BoundResource{}
	for _, resource := range *colqueue {
		returnValue = append(returnValue, resource.Clear()...)
	}
	*colqueue = Colqueue{}
	return returnValue
}

func (colqueue *Colqueue) Pop() []BoundResource {
	newColqueue := *colqueue
	res := newColqueue[0].Clear()
	newColqueue = newColqueue[1:]
	*colqueue = newColqueue
	return res
}

func (colqueue *Colqueue) Push(resources []BoundResource) {
	newColqueue := *colqueue
	newColqueue = append(newColqueue, resources)
	*colqueue = newColqueue
}

func (colqueue *Colqueue) PushPop(resources []BoundResource) []BoundResource {
	colqueue.Push(resources)
	return colqueue.Pop()
}

func (colqueue *Colqueue) Remove(index uint16) []BoundResource {
	// out of index : do nothing.
	if len(*colqueue)-1 < int(index) {
		return []BoundResource{}
	}

	return (*colqueue)[index].Clear()
}

// FeedAll feeds materials at the first-none-empty index in the site.
// It ONLY returns Code_RESOURCE_MATERIAL_SHORTAGE if there is no material
// in the site.
func (colqueue *Colqueue) FeedAll() ([]FedMaterial, error) {
	var collection Collection
	for {
		if len(*colqueue) == 0 {
			return nil, mcomErr.Error{Code: mcomErr.Code_RESOURCE_MATERIAL_SHORTAGE}
		}
		collection = colqueue.Pop()
		if len(collection) != 0 {
			break
		}
	}

	return collection.FeedAll(), nil
}

func (colqueue *Colqueue) maybeExtend(maxIndex uint16) {
	if len(*colqueue)-1 < int(maxIndex) {
		newColqueue := make(Colqueue, maxIndex+1)
		copy(newColqueue, *colqueue)

		for i := len(*colqueue); i < int(maxIndex+1); i++ {
			newColqueue[i] = []BoundResource{}
		}
		*colqueue = newColqueue
	}
}

// BindOption definition.
type BindOption struct {
	Head       bool
	Tail       bool
	QueueIndex *uint16
}

func (colqueue *Colqueue) ParseIndex(opt BindOption) uint16 {
	switch {
	case opt.Head:
		return 0
	case opt.Tail:
		return uint16(len(*colqueue) - 1)
	default:
		return *opt.QueueIndex
	}
}

// BindRecords shows the binding history of the site.
type BindRecords struct {
	StationID string         `gorm:"type:varchar(32);not null;primaryKey"`
	SiteName  string         `gorm:"type:varchar(16);not null;primaryKey"`
	SiteIndex int16          `gorm:"not null;primaryKey"`
	Records   BindRecordsSet `gorm:"type:jsonb;not null"`
}

// Model implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (BindRecords) Model() {}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (BindRecords) TableName() string {
	return "bind_records"
}

type BindRecordsSet struct {
	// key: "%s %s",ResourceID ProductType.
	Set map[string]struct{} `json:"set"`
}

// Scan implements database/sql Scanner interface.
func (brs *BindRecordsSet) Scan(src interface{}) error {
	return ScanJSON(src, brs)
}

// Value implements database/sql/driver Valuer interface.
func (brs BindRecordsSet) Value() (driver.Value, error) {
	return json.Marshal(brs)
}

type UniqueMaterialResource struct {
	ResourceID  string
	ProductType string
}

func (brs *BindRecordsSet) Add(umr UniqueMaterialResource) {
	if brs.Set == nil {
		brs.Set = make(map[string]struct{})
	}
	brs.Set[brs.parseSetKey(umr)] = struct{}{}
}

func (brs BindRecordsSet) Contains(umr UniqueMaterialResource) bool {
	_, ok := brs.Set[brs.parseSetKey(umr)]
	return ok
}

func (BindRecordsSet) parseSetKey(umr UniqueMaterialResource) string {
	return umr.ResourceID + " " + umr.ProductType
}
