package models

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/utils/resources"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func TestSiteAttributes_implementation(t *testing.T) {
	assert := assert.New(t)

	// good.
	{
		expecteds := []SiteAttributes{
			{ // slot operator.
				Type:       sites.Type_SLOT,
				SubType:    sites.SubType_OPERATOR,
				Limitation: []string{},
			}, { // slot tool.
				Type:       sites.Type_SLOT,
				SubType:    sites.SubType_TOOL,
				Limitation: []string{"spec"},
			}, { // container material.
				Type:       sites.Type_CONTAINER,
				SubType:    sites.SubType_MATERIAL,
				Limitation: []string{"CC", "UF"},
			},
		}
		for i, expected := range expecteds {
			v, err := expected.Value()
			assert.NoErrorf(err, "index %d", i)

			var s SiteAttributes
			assert.NoErrorf(s.Scan(v), "index %d", i)
			assert.Equalf(expected, s, "index %d", i)
		}
	}
}

func TestSiteContent_implementation(t *testing.T) {
	assert := assert.New(t)

	// good.
	{
		expecteds := []SiteContent{
			{ // Slot operator.
				Slot: &Slot{
					Operator: &OperatorSite{
						EmployeeID: "12345",
						Group:      9,
						WorkDate:   time.Date(2021, time.April, 1, 4, 5, 22, 214, time.Local),
					},
				},
			}, { // Slot tool.
				Slot: &Slot{
					Tool: &ToolSite{
						ResourceID:    "R5",
						InstalledTime: time.Date(2021, time.April, 1, 1, 5, 22, 214, time.Local),
					},
				},
			}, { // Colqueue material.
				Colqueue: &Colqueue{
					{
						BoundResource{
							Material: &MaterialSite{
								Material: Material{
									ID:    "M0",
									Grade: "B",
								},
								Quantity:   types.Decimal.NewFromFloat32(88.999),
								ResourceID: "resource_0",
							},
						},
						BoundResource{
							Material: &MaterialSite{
								Material: Material{
									ID:    "M1",
									Grade: "B",
								},
								Quantity:   types.Decimal.NewFromInt32(123),
								ResourceID: "resource_1",
							},
						},
					}, {
						BoundResource{
							Material: &MaterialSite{
								Material: Material{
									ID:    "M0",
									Grade: "B",
								},
								Quantity:   types.Decimal.NewFromInt32(321),
								ResourceID: "resource_0",
							},
						},
					},
				},
			}, { // Container material.
				Container: &Container{
					{
						Material: &MaterialSite{
							Material: Material{
								ID:    "M0",
								Grade: "B",
							},
							Quantity:   types.Decimal.NewFromFloat32(88.999),
							ResourceID: "resource_0",
						},
					},
				},
			}, { // Collection material.
				Collection: &Collection{
					{
						Material: &MaterialSite{
							Material: Material{
								ID:    "M0",
								Grade: "B",
							},
							Quantity:   types.Decimal.NewFromFloat32(88.999),
							ResourceID: "resource_0",
						},
					}, {
						Material: &MaterialSite{
							Material: Material{
								ID:    "M1",
								Grade: "B",
							},
							Quantity:   types.Decimal.NewFromInt32(123),
							ResourceID: "resource_1",
						},
					},
				},
			}, { // Queue material.
				Queue: &Queue{
					{
						Material: &MaterialSite{
							Material: Material{
								ID:    "M0",
								Grade: "B",
							},
							Quantity:   types.Decimal.NewFromFloat32(88.999),
							ResourceID: "resource_0",
						},
					}, {
						Material: &MaterialSite{
							Material: Material{
								ID:    "M1",
								Grade: "B",
							},
							Quantity:   types.Decimal.NewFromInt32(123),
							ResourceID: "resource_1",
						},
					},
				},
			},
		}

		for i, expected := range expecteds {
			expectedValue, err := expected.Value()
			assert.NoErrorf(err, "index %d", i)

			var s SiteContent
			assert.NoErrorf(s.Scan(expectedValue), "index %d", i)

			actual, err := s.Value()
			assert.NoError(err)
			assert.Equalf(expectedValue, actual, "index %d", i)
		}
	}
}

// #region setup
type testBoundResources struct{}

func (testBoundResources) resourceA(quantity int) BoundResource {
	return BoundResource{
		Material: &MaterialSite{
			Material: Material{
				ID:    "A",
				Grade: "S",
			},
			Quantity:    types.Decimal.NewFromInt32(int32(quantity)),
			ResourceID:  "A",
			ProductType: "A",
			Status:      resources.MaterialStatus_AVAILABLE,
			ExpiryTime:  0,
		},
	}
}

func (testBoundResources) resourceB(quantity int) BoundResource {
	return BoundResource{
		Material: &MaterialSite{
			Material: Material{
				ID:    "B",
				Grade: "S",
			},
			Quantity:    types.Decimal.NewFromInt32(int32(quantity)),
			ResourceID:  "B",
			ProductType: "B",
			Status:      resources.MaterialStatus_HOLD,
			ExpiryTime:  0,
		},
	}
}

func (testBoundResources) resourceC(quantity int) BoundResource {
	return BoundResource{
		Material: &MaterialSite{
			Material: Material{
				ID:    "C",
				Grade: "S",
			},
			Quantity:    types.Decimal.NewFromInt32(int32(quantity)),
			ResourceID:  "C",
			ProductType: "C",
			Status:      resources.MaterialStatus_INSPECTION,
			ExpiryTime:  0,
		},
	}
}

func (testBoundResources) Empty() BoundResource {
	return BoundResource{}
}

// #endregion setup

func TestContainer_CleanDeviation(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	container := Container{testBoundResources.resourceA(15)}
	// #endregion arrange

	// #region act
	container.CleanDeviation()
	// #endregion act

	// #region assert
	assert.Equal(Container{}, container)
	// #endrefion assert
}

func TestContainer_Clear(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	container := Container{testBoundResources.resourceA(15), testBoundResources.resourceB(20)}
	// #endregion arrange

	// #region act
	actual := container.Clear()
	// #endregion act

	// #region assert
	expected := []BoundResource{testBoundResources.resourceA(15), testBoundResources.resourceB(20)}
	assert.Equal(expected, actual)
	assert.Equal(Container{}, container)
	// #endrefion assert
}

func TestContainer_Add(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	container := Container{testBoundResources.resourceA(15), testBoundResources.resourceB(20)}
	// #endregion arrange

	// #region act
	container.Add([]BoundResource{testBoundResources.resourceC(20)})
	// #endregion act

	// #region assert
	assert.Equal(Container{
		testBoundResources.resourceA(15),
		testBoundResources.resourceB(20),
		testBoundResources.resourceC(20),
	}, container)
	// #endrefion assert
}

func TestContainer_Bind(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	container := Container{testBoundResources.resourceA(15), testBoundResources.resourceB(20)}
	// #endregion arrange

	// #region act
	actual := container.Bind([]BoundResource{testBoundResources.resourceC(20)})
	// #endregion act

	// #region assert
	expected := []BoundResource{testBoundResources.resourceA(15), testBoundResources.resourceB(20)}
	assert.Equal(expected, actual)
	assert.Equal(Container{testBoundResources.resourceC(20)}, container)
	// #endrefion assert
}

func TestSlot_Clear(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	slot := Slot(testBoundResources.resourceA(30))
	// #endregion arrange

	// #region act
	actual := slot.Clear()
	// #endregion act

	// #region assert
	expected := testBoundResources.resourceA(30)
	assert.Equal(expected, actual)

	assert.Equal(Slot{}, slot)
	// #endregion assert
}

func TestSlot_Bind(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	slot := Slot(testBoundResources.resourceA(30))
	// #endregion arrange

	// #region act
	actual := slot.Bind(testBoundResources.resourceC(15))
	// #endregion act

	// #region assert
	expected := testBoundResources.resourceA(30)
	assert.Equal(expected, actual)

	assert.Equal(Slot(testBoundResources.resourceC(15)), slot)
	// #endregion assert
}

func TestCollection_Clear(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	collection := Collection{testBoundResources.resourceA(5), testBoundResources.resourceB(10)}
	// #endregion arrange

	// #region act
	actual := collection.Clear()
	// #endregion act

	// #region assert
	expected := []BoundResource{testBoundResources.resourceA(5), testBoundResources.resourceB(10)}
	assert.Equal(expected, actual)

	assert.Equal(Collection{}, collection)
	// #endregion assert
}

func TestCollection_Add(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	collection := Collection{testBoundResources.resourceA(5), testBoundResources.resourceB(10)}
	// #endregion arrange

	// #region act
	collection.Add([]BoundResource{testBoundResources.resourceC(10)})
	// #endregion act

	// #region assert
	assert.Equal(Collection{
		testBoundResources.resourceA(5),
		testBoundResources.resourceB(10),
		testBoundResources.resourceC(10),
	}, collection)
	// #endregion assert
}

func TestCollection_Bind(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	collection := Collection{testBoundResources.resourceA(5), testBoundResources.resourceB(10)}
	// #endregion arrange

	// #region act
	actual := collection.Bind([]BoundResource{testBoundResources.resourceC(10)})
	// #endregion act

	// #region assert
	expected := []BoundResource{testBoundResources.resourceA(5), testBoundResources.resourceB(10)}
	assert.Equal(expected, actual)

	assert.Equal(Collection{testBoundResources.resourceC(10)}, collection)
	// #endregion assert
}

func TestColQueue_Bind(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	colqueue := Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}
	// #endregion arrange

	// #region act
	actual := colqueue.Bind(1, []BoundResource{testBoundResources.resourceB(20)})
	// #endregion act

	// #region assert
	expected := []BoundResource{testBoundResources.resourceB(15), testBoundResources.resourceC(30)}
	assert.Equal(expected, actual)

	assert.Equal(Colqueue{{testBoundResources.resourceA(10)}, {testBoundResources.resourceB(20)}}, colqueue)
	// #endregion assert
}

func TestColQueue_Add(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	colqueue := Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}
	// #endregion arrange

	// #region act
	colqueue.Add(0, []BoundResource{testBoundResources.resourceB(20)})
	// #endregion act

	// #region assert
	assert.Equal(Colqueue{
		{testBoundResources.resourceA(10), testBoundResources.resourceB(20)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}, colqueue)
	// #endregion assert
}

func TestColQueue_Clear(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	colqueue := Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}
	// #endregion arrange

	// #region act
	actual := colqueue.Clear()
	// #endregion act

	// #region assert
	expected := []BoundResource{testBoundResources.resourceA(10), testBoundResources.resourceB(15), testBoundResources.resourceC(30)}
	assert.ElementsMatch(expected, actual)

	assert.Equal(Colqueue{}, colqueue)
	// #endregion assert
}

func TestColQueue_Pop(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	colqueue := Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}
	// #endregion arrange

	// #region act
	actual := colqueue.Pop()
	// #endregion act

	// #region assert
	expected := []BoundResource{testBoundResources.resourceA(10)}
	assert.Equal(expected, actual)

	assert.Equal(Colqueue{{testBoundResources.resourceB(15), testBoundResources.resourceC(30)}}, colqueue)
	// #endregion assert
}

func TestColQueue_Push(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	colqueue := Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}
	// #endregion arrange

	// #region act
	colqueue.Push([]BoundResource{testBoundResources.resourceA(10)})
	// #endregion act

	// #region assert
	assert.Equal(Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
		{testBoundResources.resourceA(10)},
	}, colqueue)
	// #endregion assert
}

func TestColQueue_PushPop(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	colqueue := Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}
	// #endregion arrange

	// #region act
	actual := colqueue.PushPop([]BoundResource{testBoundResources.resourceA(10)})
	// #endregion act

	// #region assert
	expected := []BoundResource{testBoundResources.resourceA(10)}
	assert.Equal(expected, actual)
	assert.Equal(Colqueue{
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
		{testBoundResources.resourceA(10)},
	}, colqueue)
	// #endregion assert
}

func TestColQueue_Remove(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	colqueue := Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}
	// #endregion arrange

	// #region act
	actual := colqueue.Remove(1)
	// #endregion act

	// #region assert
	expected := []BoundResource{testBoundResources.resourceB(15), testBoundResources.resourceC(30)}
	assert.ElementsMatch(expected, actual)

	assert.Equal(Colqueue{{testBoundResources.resourceA(10)}, {}}, colqueue)
	// #endregion assert
}

func TestColqueue_maybeExtend(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	colqueue := Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}
	// #endregion arrange

	// #region act
	colqueue.maybeExtend(5)
	// #endregion act

	// #region assert
	assert.Equal(
		Colqueue{
			{testBoundResources.resourceA(10)},
			{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
			{},
			{},
			{},
			{},
		},
		colqueue)
	// #endregion assert
}

func TestColQueue_Bind_OutOfRange(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	colqueue := Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}
	// #endregion arrange

	// #region act
	actual := colqueue.Bind(3, []BoundResource{testBoundResources.resourceB(20)})
	// #endregion act

	// #region assert
	expected := []BoundResource{}
	assert.Equal(expected, actual)

	assert.Equal(Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
		{},
		{testBoundResources.resourceB(20)},
	}, colqueue)
	// #endregion assert
}

func TestColQueue_Add_OutOfRange(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	colqueue := Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
	}
	// #endregion arrange

	// #region act
	colqueue.Add(3, []BoundResource{testBoundResources.resourceB(20)})
	// #endregion act

	// #region assert
	assert.Equal(Colqueue{
		{testBoundResources.resourceA(10)},
		{testBoundResources.resourceB(15), testBoundResources.resourceC(30)},
		{},
		{testBoundResources.resourceB(20)},
	}, colqueue)
	// #endregion assert
}

func TestQueue_Bind(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	queue := Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
	}
	// #endregion arrange

	// #region act
	actual := queue.Bind(1, testBoundResources.resourceB(5))
	// #endregion act

	// #region assert
	expected := testBoundResources.resourceC(20)

	assert.Equal(expected, actual)
	assert.Equal(Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceB(5)),
	}, queue)
	// #endregion assert
}

func TestQueue_Push(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	queue := Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
	}
	// #endregion arrange

	// #region act
	queue.Push(testBoundResources.resourceB(5))
	// #endregion act

	// #region assert
	assert.Equal(Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
		Slot(testBoundResources.resourceB(5)),
	}, queue)
	// #endregion assert
}

func TestQueue_Pop(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	queue := Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
	}
	// #endregion arrange

	// #region act
	actual := queue.Pop()
	// #endregion act

	// #region assert
	expected := testBoundResources.resourceA(10)

	assert.Equal(expected, actual)
	assert.Equal(Queue{
		Slot(testBoundResources.resourceC(20)),
	}, queue)
	// #endregion assert
}

func TestQueue_PushPop(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	queue := Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
	}
	// #endregion arrange

	// #region act
	actual := queue.PushPop(testBoundResources.resourceB(5))
	// #endregion act

	// #region assert
	expected := testBoundResources.resourceA(10)

	assert.Equal(expected, actual)
	assert.Equal(Queue{
		Slot(testBoundResources.resourceC(20)),
		Slot(testBoundResources.resourceB(5)),
	}, queue)
	// #endregion assert
}

func TestQueue_Remove(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	queue := Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
	}
	// #endregion arrange

	// #region act
	actual := queue.Remove(0)
	// #endregion act

	// #region assert
	expected := testBoundResources.resourceA(10)

	assert.Equal(expected, actual)
	assert.Equal(Queue{
		Slot{},
		Slot(testBoundResources.resourceC(20)),
	}, queue)
	// #endregion assert
}

func TestQueue_Clear(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	queue := Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
	}
	// #endregion arrange

	// #region act
	actual := queue.Clear()
	// #endregion act

	// #region assert
	expected := []BoundResource{
		testBoundResources.resourceA(10),
		testBoundResources.resourceC(20),
	}

	assert.Equal(expected, actual)
	assert.Equal(Queue{}, queue)
	// #endregion assert
}

func TestQueue_maybeExtend(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	queue := Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
	}
	// #endregion arrange

	{ // should not extend.
		// #region act
		queue.maybeExtend(1)
		// #endregion act

		// #region assert
		assert.Equal(Queue{
			Slot(testBoundResources.resourceA(10)),
			Slot(testBoundResources.resourceC(20)),
		}, queue)
		// #endregion assert
	}
	{ // should extend.
		// #region act
		queue.maybeExtend(4)
		// #endregion act

		// #region assert
		assert.Equal(Queue{
			Slot(testBoundResources.resourceA(10)),
			Slot(testBoundResources.resourceC(20)),
			Slot{},
			Slot{},
			Slot{},
		}, queue)
		// #endregion assert
	}
}

func TestQueue_Bind_OutOfRange(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	queue := Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
	}
	// #endregion arrange

	// #region act
	actual := queue.Bind(4, testBoundResources.resourceB(5))
	// #endregion act

	// #region assert
	expected := BoundResource{}

	assert.Equal(expected, actual)
	assert.Equal(Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
		Slot{},
		Slot{},
		Slot(testBoundResources.resourceB(5)),
	}, queue)
	// #endregion assert
}

func TestQueue_Remove_OutOfRange(t *testing.T) {
	// #region arrange
	assert := assert.New(t)
	var testBoundResources testBoundResources
	queue := Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
	}
	// #endregion arrange

	// #region act
	actual := queue.Remove(4)
	// #endregion act

	// #region assert
	expected := BoundResource{}

	assert.Equal(expected, actual)
	assert.Equal(Queue{
		Slot(testBoundResources.resourceA(10)),
		Slot(testBoundResources.resourceC(20)),
	}, queue)
	// #endregion assert
}

// TestOperatorSite_operation only tests slot type operator site.
func TestOperatorSite_operation(t *testing.T) {
	require := require.New(t)

	{ // no operator is in the new site.
		s := NewSlotSiteContent()

		require.Empty(s.Slot.Operator.Current())
		require.True(s.Slot.Operator.AllowedLogin("anyone"))
	}
	{ // an operator has signed in.
		s := NewSlotSiteContent()

		now := time.Now()

		s.Slot.Bind(BoundResource{
			Operator: &OperatorSite{
				EmployeeID: "ya",
				Group:      9,
				WorkDate:   now,
			},
		})

		require.Equal(OperatorSite{
			EmployeeID: "ya",
			Group:      9,
			WorkDate:   now,
		}, s.Slot.Operator.Current())
		require.True(s.Slot.Operator.AllowedLogin("ya"))
		require.False(s.Slot.Operator.AllowedLogin("mismatch_one"))
	}
	{ // an operator has signed out and:
		s := NewSlotSiteContent()

		now := time.Now()

		s.Slot.Bind(BoundResource{
			Operator: &OperatorSite{
				EmployeeID: "ya",
				Group:      9,
				WorkDate:   now,
			},
		})
		s.Slot.Clear()

		require.Empty(s.Slot.Operator.Current())
		require.True(s.Slot.Operator.AllowedLogin("ya"))
		require.True(s.Slot.Operator.AllowedLogin("mismatch_one"))

		{ // another one tries to sign in.
			s.Slot.Bind(BoundResource{
				Operator: &OperatorSite{
					EmployeeID: "another",
					Group:      1,
					WorkDate:   now,
				},
			})

			require.Equal(OperatorSite{
				EmployeeID: "another",
				Group:      1,
				WorkDate:   now,
			}, s.Slot.Operator.Current())
		}
		{ // sign out again.
			require.NotPanics(func() {
				s.Slot.Clear()
			})

			require.Empty(s.Slot.Operator.Current())
			require.True(s.Slot.Operator.AllowedLogin("ya"))
			require.True(s.Slot.Operator.AllowedLogin("mismatch_one"))
		}
	}
}

func TestContainer_Feed(t *testing.T) {
	assert := assert.New(t)
	{ // case none.
		container := NewContainerSiteContent().Container

		records, shortage := container.Feed(decimal.NewFromInt(30))
		assert.Equal([]FedMaterial{}, records)
		assert.True(shortage)
		assert.Equal(NewContainerSiteContent().Container, container)

		records = container.FeedAll()
		assert.Equal([]FedMaterial{}, records)
		assert.Equal(NewContainerSiteContent().Container, container)
	}
	{ // case less.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		container := NewContainerSiteContent().Container
		container.Bind([]BoundResource{
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(5),
					ResourceID:  "A",
					ProductType: "A",
					Status:      resources.MaterialStatus_HOLD,
					ExpiryTime:  expiryTimeNano,
				},
			},
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    types.Decimal.NewFromInt32(5),
					ResourceID:  "B",
					ProductType: "B",
					Status:      resources.MaterialStatus_AVAILABLE,
					ExpiryTime:  expiryTimeNano,
				},
			},
		})

		records, shortage := container.Feed(decimal.NewFromInt(7))
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			Status:      resources.MaterialStatus_HOLD,
			ExpiryTime:  expiryTime,
			Quantity:    decimal.NewFromInt(5),
		}, {
			Material: Material{
				ID:    "B",
				Grade: "B",
			},
			ResourceID:  "B",
			ProductType: "B",
			Status:      resources.MaterialStatus_AVAILABLE,
			ExpiryTime:  expiryTime,
			Quantity:    decimal.NewFromInt(2),
		}}, records)
		assert.False(shortage)

		assert.Equal(&Container{{
			Material: &MaterialSite{
				Material: Material{
					ID:    "B",
					Grade: "B",
				},
				Quantity:    types.Decimal.NewFromInt32(3),
				ResourceID:  "B",
				ProductType: "B",
				Status:      resources.MaterialStatus_AVAILABLE,
				ExpiryTime:  expiryTimeNano,
			},
		}}, container)
	}

	{ // case equal.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		container := NewContainerSiteContent().Container
		container.Bind([]BoundResource{
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(5),
					ResourceID:  "A",
					ProductType: "A",
					ExpiryTime:  expiryTimeNano,
				},
			},
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    types.Decimal.NewFromInt32(5),
					ResourceID:  "B",
					ProductType: "B",
					ExpiryTime:  expiryTimeNano,
				},
			},
		})

		records, shortage := container.Feed(decimal.NewFromInt(10))
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			ExpiryTime:  expiryTime,
			Quantity:    decimal.NewFromInt(5),
		}, {
			Material: Material{
				ID:    "B",
				Grade: "B",
			},
			ResourceID:  "B",
			ProductType: "B",
			ExpiryTime:  expiryTime,
			Quantity:    decimal.NewFromInt(5),
		}}, records)
		assert.False(shortage)

		assert.Equal(&Container{}, container)
	}

	{ // case over.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		container := NewContainerSiteContent().Container
		container.Bind([]BoundResource{
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(5),
					ResourceID:  "A",
					ProductType: "A",
					ExpiryTime:  expiryTimeNano,
				},
			},
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    types.Decimal.NewFromInt32(5),
					ResourceID:  "B",
					ProductType: "B",
					ExpiryTime:  expiryTimeNano,
				},
			},
		})

		records, shortage := container.Feed(decimal.NewFromInt(11))
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			ExpiryTime:  expiryTime,
			Quantity:    decimal.NewFromInt(5),
		}, {
			Material: Material{
				ID:    "B",
				Grade: "B",
			},
			ResourceID:  "B",
			ProductType: "B",
			ExpiryTime:  expiryTime,
			Quantity:    decimal.NewFromInt(6),
		}}, records)
		assert.True(shortage)

		assert.Equal(&Container{}, container)
	}

	{ // case feed all.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		container := NewContainerSiteContent().Container
		container.Bind([]BoundResource{
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(5),
					ResourceID:  "A",
					ProductType: "A",
					ExpiryTime:  expiryTimeNano,
				},
			},
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    types.Decimal.NewFromInt32(5),
					ResourceID:  "B",
					ProductType: "B",
					ExpiryTime:  expiryTimeNano,
				},
			},
		})

		records := container.FeedAll()
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			ExpiryTime:  expiryTime,
			Quantity:    decimal.NewFromInt(5),
		}, {
			Material: Material{
				ID:    "B",
				Grade: "B",
			},
			ResourceID:  "B",
			ProductType: "B",
			ExpiryTime:  expiryTime,
			Quantity:    decimal.NewFromInt(5),
		}}, records)

		assert.Equal(&Container{}, container)
	}

}

func TestSlot_Feed(t *testing.T) {
	assert := assert.New(t)
	{ // case none.
		slot := NewSlotSiteContent().Slot

		records, shortage, noQtyMaterials := slot.Feed(decimal.NewFromInt(10))
		assert.Equal([]FedMaterial{}, records)
		assert.True(shortage)
		assert.Equal(NewSlotSiteContent().Slot, slot)
		assert.False(noQtyMaterials)

		records = slot.FeedAll()
		assert.Equal([]FedMaterial{}, records)
		assert.Equal(NewSlotSiteContent().Slot, slot)
	}
	{ // case less.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		slot := NewSlotSiteContent().Slot
		slot.Bind(BoundResource{
			Material: &MaterialSite{
				Material: Material{
					ID:    "A",
					Grade: "A",
				},
				Quantity:    types.Decimal.NewFromInt32(10),
				ResourceID:  "A",
				ProductType: "A",
				ExpiryTime:  expiryTimeNano,
			}})

		records, shortage, noQtyMaterials := slot.Feed(decimal.NewFromInt(8))
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			Quantity:    decimal.NewFromInt(8),
			ExpiryTime:  expiryTime,
		}}, records)
		assert.False(shortage)
		assert.False(noQtyMaterials)
		assert.Equal(&Slot{
			Material: &MaterialSite{
				Material: Material{
					ID:    "A",
					Grade: "A",
				},
				Quantity:    types.Decimal.NewFromInt32(2),
				ResourceID:  "A",
				ProductType: "A",
				ExpiryTime:  expiryTimeNano,
			},
		}, slot)
	}

	{ // case equal.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		slot := NewSlotSiteContent().Slot
		slot.Bind(BoundResource{
			Material: &MaterialSite{
				Material: Material{
					ID:    "A",
					Grade: "A",
				},
				Quantity:    types.Decimal.NewFromInt32(10),
				ResourceID:  "A",
				ProductType: "A",
				ExpiryTime:  expiryTimeNano,
			}})

		records, shortage, noQtyMaterials := slot.Feed(decimal.NewFromInt(10))
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			Quantity:    decimal.NewFromInt(10),
			ExpiryTime:  expiryTime,
		}}, records)
		assert.False(shortage)
		assert.False(noQtyMaterials)
		assert.Equal(&Slot{}, slot)
	}

	{ // case over.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		slot := NewSlotSiteContent().Slot
		slot.Bind(BoundResource{
			Material: &MaterialSite{
				Material: Material{
					ID:    "A",
					Grade: "A",
				},
				Quantity:    types.Decimal.NewFromInt32(10),
				ResourceID:  "A",
				ProductType: "A",
				ExpiryTime:  expiryTimeNano,
			}})

		records, shortage, noQtyMaterials := slot.Feed(decimal.NewFromInt(15))
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			Quantity:    decimal.NewFromInt(15),
			ExpiryTime:  expiryTime,
		}}, records)
		assert.True(shortage)
		assert.False(noQtyMaterials)
		assert.Equal(&Slot{}, slot)
	}

	{ // case feed all.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		slot := NewSlotSiteContent().Slot
		slot.Bind(BoundResource{
			Material: &MaterialSite{
				Material: Material{
					ID:    "A",
					Grade: "A",
				},
				Quantity:    types.Decimal.NewFromInt32(10),
				ResourceID:  "A",
				ProductType: "A",
				ExpiryTime:  expiryTimeNano,
			}})

		records := slot.FeedAll()
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			Quantity:    decimal.NewFromInt(10),
			ExpiryTime:  expiryTime,
		}}, records)
		assert.Equal(&Slot{}, slot)
	}
	{ // test unspecified quantity.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		slot := NewSlotSiteContent().Slot
		slot.Bind(BoundResource{
			Material: &MaterialSite{
				Material: Material{
					ID:    "A",
					Grade: "A",
				},
				Quantity:    nil,
				ResourceID:  "A",
				ProductType: "A",
				ExpiryTime:  expiryTimeNano,
			},
		})
		materials, shortage, maybeBeFedMaterials := slot.Feed(decimal.NewFromInt(20))
		assert.False(shortage)
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			Quantity:    decimal.NewFromInt(20),
			ExpiryTime:  expiryTime,
		}}, materials)
		assert.True(maybeBeFedMaterials)
	}
}

func TestCollection_FeedAll(t *testing.T) {
	assert := assert.New(t)
	{ // empty collection.
		collection := NewCollectionSiteContent().Collection
		records := collection.FeedAll()
		assert.Equal([]FedMaterial{}, records)
		assert.Equal(NewCollectionSiteContent().Collection, collection)
	}
	{ // ordinary case.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		collection := NewCollectionSiteContent().Collection
		collection.Bind([]BoundResource{
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
					ExpiryTime:  expiryTimeNano,
				},
			},
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "B",
					ProductType: "B",
					ExpiryTime:  expiryTimeNano,
				},
			},
		})

		records := collection.FeedAll()
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			Quantity:    decimal.NewFromInt(10),
			ExpiryTime:  expiryTime,
		}, {
			Material: Material{
				ID:    "B",
				Grade: "B",
			},
			ResourceID:  "B",
			ProductType: "B",
			Quantity:    decimal.NewFromInt(10),
			ExpiryTime:  expiryTime,
		}}, records)

		assert.Equal(&Collection{}, collection)
	}
}

func TestColQueue_FeedAll(t *testing.T) {
	assert := assert.New(t)
	{ // good case.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		colqueue := NewColqueueSiteContent().Colqueue
		colqueue.Push([]BoundResource{
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
					ExpiryTime:  expiryTimeNano,
				},
			},
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "B",
					ProductType: "B",
					ExpiryTime:  expiryTimeNano,
				},
			},
		})
		colqueue.Push([]BoundResource{
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "C",
						Grade: "C",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "C",
					ProductType: "C",
					ExpiryTime:  expiryTimeNano,
				},
			},
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "D",
						Grade: "D",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "D",
					ProductType: "D",
					ExpiryTime:  expiryTimeNano,
				},
			},
		})
		records, err := colqueue.FeedAll()
		assert.NoError(err)
		assert.Equal([]FedMaterial{{
			Material: Material{
				ID:    "A",
				Grade: "A",
			},
			ResourceID:  "A",
			ProductType: "A",
			Quantity:    decimal.NewFromInt(10),
			ExpiryTime:  expiryTime,
		}, {
			Material: Material{
				ID:    "B",
				Grade: "B",
			},
			ResourceID:  "B",
			ProductType: "B",
			Quantity:    decimal.NewFromInt(10),
			ExpiryTime:  expiryTime,
		}}, records)

		assert.Equal(&Colqueue{
			{
				{
					Material: &MaterialSite{
						Material: Material{
							ID:    "C",
							Grade: "C",
						},
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "C",
						ProductType: "C",
						ExpiryTime:  expiryTimeNano,
					},
				},
				{
					Material: &MaterialSite{
						Material: Material{
							ID:    "D",
							Grade: "D",
						},
						Quantity:    types.Decimal.NewFromInt32(10),
						ResourceID:  "D",
						ProductType: "D",
						ExpiryTime:  expiryTimeNano,
					},
				},
			},
		}, colqueue)
	}
	{ // special case.
		expiryTimeNano := types.TimeNano(12345)
		expiryTime := expiryTimeNano.Time()
		colqueue := NewColqueueSiteContent().Colqueue
		colqueue.Push([]BoundResource{})
		colqueue.Push([]BoundResource{
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "A",
						Grade: "A",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "A",
					ProductType: "A",
					ExpiryTime:  expiryTimeNano,
				},
			},
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "B",
						Grade: "B",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "B",
					ProductType: "B",
					ExpiryTime:  expiryTimeNano,
				},
			},
		})
		colqueue.Push([]BoundResource{
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "C",
						Grade: "C",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "C",
					ProductType: "C",
					ExpiryTime:  expiryTimeNano,
				},
			},
		})

		records, err := colqueue.FeedAll()
		assert.NoError(err)
		assert.Equal([]FedMaterial{
			{
				Material: Material{
					ID:    "A",
					Grade: "A",
				},
				ResourceID:  "A",
				ProductType: "A",
				Quantity:    decimal.NewFromInt(10),
				ExpiryTime:  expiryTime,
			},
			{
				Material: Material{
					ID:    "B",
					Grade: "B",
				},
				ResourceID:  "B",
				ProductType: "B",
				Quantity:    decimal.NewFromInt(10),
				ExpiryTime:  expiryTime,
			},
		}, records)
		assert.Equal(&Colqueue{{
			{
				Material: &MaterialSite{
					Material: Material{
						ID:    "C",
						Grade: "C",
					},
					Quantity:    types.Decimal.NewFromInt32(10),
					ResourceID:  "C",
					ProductType: "C",
					ExpiryTime:  expiryTimeNano,
				},
			},
		}}, colqueue)
	}
	{ // shortage.
		colqueue := NewColqueueSiteContent().Colqueue
		_, err := colqueue.FeedAll()
		assert.ErrorIs(mcomErr.Error{Code: mcomErr.Code_RESOURCE_MATERIAL_SHORTAGE}, err)

		colqueue.Push([]BoundResource{})
		_, err = colqueue.FeedAll()
		assert.ErrorIs(mcomErr.Error{Code: mcomErr.Code_RESOURCE_MATERIAL_SHORTAGE}, err)
	}
}

func TestQueue_FeedAll(t *testing.T) {
	assert := assert.New(t)
	testBoundResources := testBoundResources{}
	expiryTimeNano := types.TimeNano(0)
	expiryTime := expiryTimeNano.Time()
	{ // good case.
		queue := NewQueueSiteContent().Queue
		queue.Push(testBoundResources.resourceA(10))
		queue.Push(testBoundResources.resourceB(20))

		rec, err := queue.FeedAll()
		assert.NoError(err)

		expectedQueue := &Queue{
			Slot(testBoundResources.resourceB(20)),
		}
		material := testBoundResources.resourceA(10).Material
		expectedRec := []FedMaterial{{
			Material:    material.Material,
			ResourceID:  material.ResourceID,
			ProductType: material.ProductType,
			Quantity:    *material.Quantity,
			Status:      material.Status,
			ExpiryTime:  expiryTime,
		}}

		assert.Equal(expectedQueue, queue)
		assert.Equal(expectedRec, rec)
	}
	{ // shortage.
		queue := NewQueueSiteContent().Queue

		_, err := queue.FeedAll()
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_RESOURCE_MATERIAL_SHORTAGE})
	}
}

func TestQueue_Feed(t *testing.T) {
	assert := assert.New(t)
	testBoundResources := testBoundResources{}
	expiryTimeNano := types.TimeNano(0)
	expiryTime := expiryTimeNano.Time()
	{ // good case.
		queue := NewQueueSiteContent().Queue
		queue.Push(testBoundResources.resourceA(10))
		queue.Push(testBoundResources.resourceB(20))

		rec, isShortage, noQtyMaterials, err := queue.Feed(decimal.NewFromInt(2))
		assert.NoError(err)
		assert.False(isShortage)
		assert.False(noQtyMaterials)

		expectedQueue := &Queue{
			Slot(testBoundResources.resourceA(8)),
			Slot(testBoundResources.resourceB(20)),
		}
		material := testBoundResources.resourceA(2).Material
		expectedRec := []FedMaterial{{
			Material:    material.Material,
			ResourceID:  material.ResourceID,
			ProductType: material.ProductType,
			Quantity:    *material.Quantity,
			Status:      material.Status,
			ExpiryTime:  expiryTime,
		}}

		assert.Equal(expectedQueue, queue)
		assert.Equal(expectedRec, rec)
	}
	{ // special case : skip empty slots.
		queue := NewQueueSiteContent().Queue
		queue.Push(BoundResource{})
		queue.Push(BoundResource{})
		queue.Push(testBoundResources.resourceA(10))
		queue.Push(testBoundResources.resourceB(20))

		rec, isShortage, noQtyMaterials, err := queue.Feed(decimal.NewFromInt(2))
		assert.NoError(err)
		assert.False(isShortage)
		assert.False(noQtyMaterials)

		expectedQueue := &Queue{
			Slot(testBoundResources.resourceA(8)),
			Slot(testBoundResources.resourceB(20)),
		}
		material := testBoundResources.resourceA(2).Material
		expectedRec := []FedMaterial{{
			Material:    material.Material,
			ResourceID:  material.ResourceID,
			ProductType: material.ProductType,
			Quantity:    *material.Quantity,
			Status:      material.Status,
			ExpiryTime:  expiryTime,
		}}

		assert.Equal(expectedQueue, queue)
		assert.Equal(expectedRec, rec)
	}
	{ // special case : first resource shortage.
		queue := NewQueueSiteContent().Queue
		queue.Push(testBoundResources.resourceA(10))
		queue.Push(testBoundResources.resourceB(20))

		rec, isShortage, noQtyMaterials, err := queue.Feed(decimal.NewFromInt(12))
		assert.NoError(err)
		assert.True(isShortage)
		assert.False(noQtyMaterials)

		expectedQueue := &Queue{
			Slot(testBoundResources.resourceB(20)),
		}
		material := testBoundResources.resourceA(12).Material
		expectedRec := []FedMaterial{{
			Material:    material.Material,
			ResourceID:  material.ResourceID,
			ProductType: material.ProductType,
			Quantity:    *material.Quantity,
			Status:      material.Status,
			ExpiryTime:  expiryTime,
		}}

		assert.Equal(expectedQueue, queue)
		assert.Equal(expectedRec, rec)
	}
	{ // shortage.
		queue := NewQueueSiteContent().Queue

		_, _, _, err := queue.Feed(decimal.NewFromInt(20))
		assert.ErrorIs(err, mcomErr.Error{Code: mcomErr.Code_RESOURCE_MATERIAL_SHORTAGE})
	}
}

func Test_BindRecordsSet(t *testing.T) {
	assert := assert.New(t)
	var brs BindRecordsSet
	{ // not found.
		assert.False(brs.Contains(UniqueMaterialResource{
			ResourceID:  "A",
			ProductType: "A",
		}))
	}
	{ // add.
		brs.Add(UniqueMaterialResource{
			ResourceID:  "A",
			ProductType: "A",
		})

		assert.True(brs.Contains(UniqueMaterialResource{
			ResourceID:  "A",
			ProductType: "A",
		}))
	}
}

func TestSiteContents_AfterFind(t *testing.T) {
	assert := assert.New(t)

	{ // replace the contents in container, collection and colqueue sites.
		contents := SiteContents{
			Name:    "test",
			Index:   0,
			Station: "test",
			Content: SiteContent{
				Slot: &Slot{
					Material: &MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				},
				Container: &Container{{
					Material: &MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				}},
				Collection: &Collection{{
					Material: &MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				}},
				Queue: &Queue{{
					Material: &MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				}},
				Colqueue: &Colqueue{{{
					Material: &MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				}}},
			},
		}

		assert.NoError(contents.AfterFind(nil))

		expected := SiteContents{
			Name:    "test",
			Index:   0,
			Station: "test",
			Content: SiteContent{
				Slot: &Slot{
					Material: &MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				},
				Container: &Container{{
					Material: &MaterialSite{
						Quantity:   types.Decimal.NewFromInt32(0),
						ResourceID: "id",
					},
				}},
				Collection: &Collection{{
					Material: &MaterialSite{
						Quantity:   types.Decimal.NewFromInt32(0),
						ResourceID: "id",
					},
				}},
				Queue: &Queue{{
					Material: &MaterialSite{
						Quantity:   nil,
						ResourceID: "id",
					},
				}},
				Colqueue: &Colqueue{{{
					Material: &MaterialSite{
						Quantity:   types.Decimal.NewFromInt32(0),
						ResourceID: "id",
					},
				}}},
			},
		}

		assert.Equal(expected, contents)
	}
	{ // do not replace the contents in operator and tool sites.
		contents := SiteContents{
			Name: "test2",
			Content: SiteContent{
				Container: &Container{{
					Tool: &ToolSite{
						ResourceID: "t",
						ToolID:     "t",
					},
				}},
			},
		}

		assert.NoError(contents.AfterFind(nil))

		expected := SiteContents{
			Name: "test2",
			Content: SiteContent{
				Container: &Container{{
					Tool: &ToolSite{
						ResourceID: "t",
						ToolID:     "t",
					},
				}},
			},
		}

		assert.Equal(expected, contents)
	}
}
