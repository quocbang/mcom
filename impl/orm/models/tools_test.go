package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ToolResource(t *testing.T) {
	assert := assert.New(t)
	var tr ToolResource
	site := UniqueSite{
		SiteID: SiteID{
			Name:  "site",
			Index: 2,
		},
		Station: "station",
	}

	tr.BindTo(site)
	assert.Equal(tr, ToolResource{BindingSite: site})

	tr.Unbind()
	assert.Equal(tr, ToolResource{})
}
