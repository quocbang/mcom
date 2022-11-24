package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStationInformation_implementation(t *testing.T) {
	assert := assert.New(t)

	// good.
	{
		expected := StationInformation{
			Code:        "7",
			Description: "(╯‵□′)╯︵┴─┴",
		}
		v, err := expected.Value()
		assert.NoError(err)

		var s StationInformation
		assert.NoError(s.Scan(v))
		assert.Equal(expected, s)
	}
}

func TestStationSites_implementation(t *testing.T) {
	assert := assert.New(t)

	// good.
	{
		expected := SitesID{
			{
				Name:  "site_0",
				Index: 0,
			},
			{
				Name:  "site_1",
				Index: 0,
			},
			{
				Name:  "site_1",
				Index: 1,
			},
		}

		v, err := expected.Value()
		assert.NoError(err)

		var s SitesID
		assert.NoError(s.Scan(v))
		assert.Equal(expected, s)
	}
}

func TestStation_RemoveSharedSite(t *testing.T) {
	assert := assert.New(t)

	station := Station{
		ID: "test",
		Sites: []UniqueSite{
			{
				SiteID: SiteID{
					Name:  "test0",
					Index: 0,
				},
				Station: "test",
			},
			{
				SiteID: SiteID{
					Name:  "test1",
					Index: 1,
				},
				Station: "",
			},
			{
				SiteID: SiteID{
					Name:  "test2",
					Index: 2,
				},
				Station: "test",
			},
		},
	}

	assert.False(station.RemoveSharedSite(SiteID{
		Name:  "test0",
		Index: 0,
	}))

	assert.False(station.RemoveSharedSite(SiteID{
		Name:  "test2",
		Index: 2,
	}))

	assert.Equal(Station{
		ID: "test",
		Sites: []UniqueSite{
			{
				SiteID: SiteID{
					Name:  "test0",
					Index: 0,
				},
				Station: "test",
			},
			{
				SiteID: SiteID{
					Name:  "test1",
					Index: 1,
				},
				Station: "",
			},
			{
				SiteID: SiteID{
					Name:  "test2",
					Index: 2,
				},
				Station: "test",
			},
		},
	}, station)

	assert.True(station.RemoveSharedSite(SiteID{
		Name:  "test1",
		Index: 1,
	}))

	assert.Equal(Station{
		ID: "test",
		Sites: []UniqueSite{
			{
				SiteID: SiteID{
					Name:  "test0",
					Index: 0,
				},
				Station: "test",
			},
			{
				SiteID: SiteID{
					Name:  "test2",
					Index: 2,
				},
				Station: "test",
			},
		},
	}, station)
}
