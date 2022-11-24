package impl

import (
	"context"
	"strings"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models/cloud"
)

func (dm *DataManager) CreateBlobResourceRecord(ctx context.Context, req mcom.CreateBlobResourceRecordRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}
	session := dm.newSession(ctx)
	return session.createBlobResourceRecord(req)
}

func (session *session) createBlobResourceRecord(req mcom.CreateBlobResourceRecordRequest) error {
	toCreate := make([]cloud.Blob, len(req.Details))
	for i := range req.Details {
		toCreate[i] = cloud.Blob{
			BlobURI:       req.Details[i].BlobURI,
			Resources:     req.Details[i].Resources,
			Station:       req.Details[i].Station,
			DateTime:      req.Details[i].DateTime,
			ContainerName: req.Details[i].ContainerName,
		}
	}
	err := session.db.Model(&cloud.Blob{}).Create(toCreate).Error

	if IsPqError(err, UniqueViolation) {
		uris := make([]string, len(req.Details))
		for i := range req.Details {
			uris[i] = req.Details[i].BlobURI
		}
		names := strings.Join(uris, "\t")
		return mcomErr.Error{Code: mcomErr.Code_BLOB_ALREADY_EXIST, Details: "uri: " + names}
	}
	return err
}

func (dm *DataManager) ListBlobURIs(ctx context.Context, req mcom.ListBlobURIsRequest) (mcom.ListBlobURIsReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListBlobURIsReply{}, err
	}
	session := dm.newSession(ctx)
	return session.ListBlobURIs(req)
}

func (session *session) ListBlobURIs(req mcom.ListBlobURIsRequest) (mcom.ListBlobURIsReply, error) {
	var res []cloud.Blob
	db := session.db.Where(&cloud.Blob{
		Station:       req.Station,
		DateTime:      req.DateTime,
		ContainerName: req.ContainerName,
	})
	if req.Resource != "" {
		db = db.Where(`? = ANY(resources)`, req.Resource)
	}
	if err := db.Find(&res).Error; err != nil {
		return mcom.ListBlobURIsReply{}, err
	}

	uris := make([]mcom.BlobURI, len(res))
	for i := range res {
		uris[i].URI = res[i].BlobURI
		uris[i].ContainerName = res[i].ContainerName
	}

	return mcom.ListBlobURIsReply{
		BlobURIs: uris,
	}, nil
}
