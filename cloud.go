package mcom

import (
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

type CreateBlobResourceRecordRequest struct {
	Details []CreateBlobResourceRecordDetail `validate:"gt=0,dive"`
}

type CreateBlobResourceRecordDetail struct {
	BlobURI       string `validate:"required"`
	Resources     []string
	Station       string         `validate:"required"`
	DateTime      types.TimeNano `validate:"required"`
	ContainerName string         `validate:"required"`
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req CreateBlobResourceRecordRequest) CheckInsufficiency() error {
	if err := validate.Struct(req); err != nil {
		return mcomErr.Error{Code: mcomErr.Code_INSUFFICIENT_REQUEST, Details: err.Error()}
	}
	return nil
}

type ListBlobURIsRequest struct {
	Resource      string
	Station       string
	ContainerName string
	DateTime      types.TimeNano
}

// CheckInsufficiency implements gitlab.kenda.com.tw/kenda/mcom Request interface.
func (req ListBlobURIsRequest) CheckInsufficiency() error {
	return nil
}

// ListBlobURIsReply definition
type ListBlobURIsReply struct {
	BlobURIs []BlobURI
}

type BlobURI struct {
	ContainerName string
	URI           string
}
