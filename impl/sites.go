package impl

import (
	"context"
	"errors"
	"fmt"

	"github.com/ahmetb/go-linq"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
)

// ListSiteType implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListSiteType(context.Context) (mcom.ListSiteTypeReply, error) {
	types := make([]mcom.SiteType, len(sites.Type_value)-1)
	for i := 1; ; i++ {
		k, ok := sites.Type_name[int32(i)]
		if !ok {
			break
		}
		types[i-1] = mcom.SiteType{
			Name:  k,
			Value: sites.Type(i),
		}
	}
	return types, nil
}

// ListSiteSubType implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListSiteSubType(context.Context) (mcom.ListSiteSubTypeReply, error) {
	subTypes := make([]mcom.SiteSubType, len(sites.SubType_value)-1)
	for i := 1; ; i++ {
		k, ok := sites.SubType_name[int32(i)]
		if !ok {
			break
		}
		subTypes[i-1] = mcom.SiteSubType{
			Name:  k,
			Value: sites.SubType(i),
		}
	}
	return subTypes, nil
}

// GetSite implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetSite(ctx context.Context, req mcom.GetSiteRequest) (mcom.GetSiteReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetSiteReply{}, err
	}

	session := dm.newSession(ctx)

	var site models.Site
	var siteContents models.SiteContents

	if err := session.db.Where(`name = ? AND index = ? AND station = ?`, req.SiteName, req.SiteIndex, req.StationID).Take(&site).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetSiteReply{}, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND}
		}
		return mcom.GetSiteReply{}, err
	}

	if err := session.db.Where(`name = ? AND index = ? AND station = ?`, req.SiteName, req.SiteIndex, site.Station).Take(&siteContents).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetSiteReply{}, fmt.Errorf("site table and site_contents table do not match")
		}
		return mcom.GetSiteReply{}, err
	}

	return mcom.GetSiteReply{
		Name:               req.SiteName,
		Index:              req.SiteIndex,
		AdminDepartmentOID: site.AdminDepartmentID,
		Attributes:         mcom.NewSiteAttributes(site.Attributes),
		Content:            siteContents.Content,
	}, nil
}

func (session *session) checkStationSiteRelation(stationName, siteName string, siteIndex int16) error {
	// station contains specified site.
	var station models.Station
	if err := session.db.Where(&models.Station{ID: stationName}).First(&station).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcomErr.Error{
				Code: mcomErr.Code_STATION_NOT_FOUND,
			}
		}
		return err
	}
	if !linq.From(station.Sites).AnyWith(func(site interface{}) bool {
		return site.(models.UniqueSite).SiteID.Name == siteName &&
			site.(models.UniqueSite).SiteID.Index == siteIndex
	}) {
		return mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND}
	}
	return nil
}

func (dm *DataManager) ListAssociatedStations(ctx context.Context, req mcom.ListAssociatedStationsRequest) (mcom.ListAssociatedStationsReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListAssociatedStationsReply{}, err
	}

	tx := dm.beginTx(ctx)
	defer tx.Rollback() // nolint: errcheck

	res, err := tx.listAssociatedStations(req)
	if err != nil {
		return mcom.ListAssociatedStationsReply{}, err
	}
	return res, tx.Commit()
}

func (tx *txDataManager) listAssociatedStations(req mcom.ListAssociatedStationsRequest) (mcom.ListAssociatedStationsReply, error) {
	var stationIDs []string

	if err := tx.db.Model(&models.Station{}).Select("id").
		Where(`sites @> ('[{"site":{"name":' || TO_JSON(?::TEXT)::TEXT || ',"index":' || ? || '},"station":' || TO_JSON(?::TEXT)::TEXT || '}]')::jsonb`, req.Site.SiteID.Name, req.Site.SiteID.Index, req.Site.Station).
		Order("id").Find(&stationIDs).Error; err != nil {
		return mcom.ListAssociatedStationsReply{}, err
	}

	if len(stationIDs) == 0 {
		return mcom.ListAssociatedStationsReply{}, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND}
	}

	res := mcom.ListAssociatedStationsReply{
		StationIDs: stationIDs,
	}
	return res, nil
}
