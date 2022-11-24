package impl

import (
	"testing"

	"gitlab.kenda.com.tw/kenda/mcom"
)

func TestDataManager(t *testing.T) {
	// implement DataManager interface.
	var _ mcom.DataManager = &DataManager{}
}
