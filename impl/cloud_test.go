package impl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models/cloud"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func TestDataManager_CreateBlobResourceRecord_ListBlobURIs(t *testing.T) {
	assert := assert.New(t)
	ctx, dm, db := initializeDB(t)
	defer dm.Close()
	cm := newClearMaster(db, &cloud.Blob{})
	assert.NoError(cm.Clear())

	now := types.ToTimeNano(time.Now())
	uri := "https://testuri/"

	{ // CreateBlobResourceRecord good case
		assert.NoError(dm.CreateBlobResourceRecord(ctx, mcom.CreateBlobResourceRecordRequest{
			Details: []mcom.CreateBlobResourceRecordDetail{
				{
					BlobURI:       uri,
					Resources:     []string{"b1", "b2", "b3"},
					Station:       "station",
					DateTime:      now,
					ContainerName: "container",
				},
			},
		}))

		actual, err := dm.ListBlobURIs(ctx, mcom.ListBlobURIsRequest{
			DateTime: now,
		})
		assert.NoError(err)

		assert.Equal(mcom.ListBlobURIsReply{BlobURIs: []mcom.BlobURI{{
			ContainerName: "container",
			URI:           uri,
		}}}, actual)
	}
	{ // CreateBlobResourceRecord already exist
		assert.ErrorIs(dm.CreateBlobResourceRecord(ctx, mcom.CreateBlobResourceRecordRequest{
			Details: []mcom.CreateBlobResourceRecordDetail{
				{
					BlobURI:       uri,
					Resources:     []string{"b1", "b2", "b3"},
					Station:       "station",
					DateTime:      now,
					ContainerName: "container",
				},
			},
		}), mcomErr.Error{Code: mcomErr.Code_BLOB_ALREADY_EXIST, Details: "uri: https://testuri/"})
	}
	{ // ListBlobURIs by barcode
		actual, err := dm.ListBlobURIs(ctx, mcom.ListBlobURIsRequest{
			Resource: "b1",
		})
		assert.NoError(err)

		assert.Equal(mcom.ListBlobURIsReply{BlobURIs: []mcom.BlobURI{{
			ContainerName: "container",
			URI:           uri,
		}}}, actual)
	}
}
