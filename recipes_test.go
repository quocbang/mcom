package mcom

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func Test_ToSubstitution(t *testing.T) {
	require := require.New(t)
	actual := SubstitutionDeletionContent{
		ProductID: models.ProductID{
			ID:    "A",
			Grade: "B",
		},
	}.ToSubstitution()
	expected := models.Substitution{
		ID:    "A",
		Grade: "B",
	}
	require.Equal(expected, actual)
}

func Test_ToSubstitutions(t *testing.T) {
	require := require.New(t)
	actual := SubstitutionDeletionContents{{
		ProductID: models.ProductID{
			ID:    "A",
			Grade: "B",
		},
	}}.ToSubstitutions()
	expected := models.Substitutions{{
		ID:    "A",
		Grade: "B",
	}}
	require.Equal(expected, actual)
}
