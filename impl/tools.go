package impl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	"gitlab.kenda.com.tw/kenda/mcom"
	mcomErr "gitlab.kenda.com.tw/kenda/mcom/errors"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/bindtype"
	"gitlab.kenda.com.tw/kenda/mcom/utils/sites"
)

// GetToolResource implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) GetToolResource(ctx context.Context, req mcom.GetToolResourceRequest) (mcom.GetToolResourceReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.GetToolResourceReply{}, err
	}

	var res models.ToolResource
	session := dm.newSession(ctx)
	if err := session.db.Where(models.ToolResource{ID: req.ResourceID}).Take(&res).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mcom.GetToolResourceReply{}, mcomErr.Error{Code: mcomErr.Code_RESOURCE_NOT_FOUND}
		}
		return mcom.GetToolResourceReply{}, err
	}
	return mcom.GetToolResourceReply{
		ToolID:      res.ToolID,
		BindingSite: res.BindingSite,
		CreatedBy:   res.CreatedBy,
		CreatedAt:   res.CreatedAt,
	}, nil
}

// ToolResourceBind implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
// Deprecated: use V2 instead
func (dm *DataManager) ToolResourceBind(ctx context.Context, req mcom.ToolResourceBindRequest) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	updatedBy := commonsCtx.UserID(ctx)

	session := dm.newSession(ctx)
	tx := session.beginTx()
	defer tx.Rollback() // nolint: errcheck

	// #region get site information
	isSharedSite := req.Station == ""
	var stationSites []mcom.ListStationSite

	if !isSharedSite {
		station, err := session.getStation(req.Station)
		if err != nil {
			return err
		}
		stationSites, err = tx.listStationSites(station.Sites, selectForUpdate())
		if err != nil {
			return err
		}
	} else {
		temp, err := dm.GetSite(ctx, mcom.GetSiteRequest{
			StationID: "",
			SiteName:  req.Details[0].Site.Name,
			SiteIndex: req.Details[0].Site.Index,
		})

		if err != nil {
			return err
		}

		stationSites = append(stationSites, mcom.ListStationSite{
			Information: mcom.ListStationSitesInformation{
				UniqueSite: models.UniqueSite{
					SiteID: models.SiteID{
						Name:  temp.Name,
						Index: temp.Index,
					},
					Station: req.Station,
				},
				Type:    temp.Attributes.Type,
				SubType: temp.Attributes.SubType,
			},
			Content: temp.Content,
		})
	}

	sitesInfo := make(map[models.SiteID]mcom.ListStationSite)
	for _, site := range stationSites {
		sitesInfo[site.Information.SiteID] = site
	}

	getSitesInfo := func(id models.SiteID) (mcom.ListStationSite, error) {
		res, ok := sitesInfo[id]
		if !ok {
			return mcom.ListStationSite{}, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND}
		}
		if res.Information.Type == sites.Type_SLOT && res.Information.SubType == sites.SubType_TOOL {
			return res, nil
		}
		return mcom.ListStationSite{},
			mcomErr.Error{
				Code:    mcomErr.Code_STATION_SITE_SUB_TYPE_MISMATCH,
				Details: fmt.Sprintf("illegal site type or sub type, site name: %s, site index: %d", id.Name, id.Index),
			}
	}
	// #endregion get site information

	// #region bind
	toUpdate := struct {
		siteContents    map[models.UniqueSite]models.SiteContent
		bindResources   []models.ToolResource
		unbindResources []string
	}{
		siteContents: make(map[models.UniqueSite]models.SiteContent),
	}

	for _, detail := range req.Details {
		site, err := getSitesInfo(detail.Site)
		if err != nil {
			return err
		}

		if detail.Type == bindtype.BindType_RESOURCE_BINDING_SLOT_BIND {
			if err := detail.Resource.Require(); err != nil {
				return err
			}

			resource := detail.Resource.ToModelResource()

			unbindResource := site.Content.Slot.Bind(models.BoundResource{
				Tool: &models.ToolSite{
					ResourceID:    resource.ID,
					ToolID:        resource.ToolID,
					InstalledTime: time.Now(),
				},
			})
			toUpdate.siteContents[models.UniqueSite{
				SiteID:  site.Information.SiteID,
				Station: req.Station,
			}] = site.Content

			resource.BindTo(models.UniqueSite{
				SiteID:  site.Information.SiteID,
				Station: req.Station,
			})
			resource.UpdatedBy = updatedBy
			toUpdate.bindResources = append(toUpdate.bindResources, resource)

			if unbindResource.Tool != nil {
				toUpdate.unbindResources = append(toUpdate.unbindResources, unbindResource.Tool.ResourceID)
			}
		} else { // clear.
			if err := detail.Resource.Empty(); err != nil {
				return err
			}
			unbindResource := site.Content.Slot.Clear()
			toUpdate.siteContents[models.UniqueSite{
				SiteID:  site.Information.SiteID,
				Station: req.Station,
			}] = site.Content

			if unbindResource.Tool != nil {
				toUpdate.unbindResources = append(toUpdate.unbindResources, unbindResource.Tool.ResourceID)
			}
		}
	}
	// #endregion bind

	// #region get unbind resources
	var unbindResources []models.ToolResource
	if err := tx.db.Model(&models.ToolResource{}).
		Where(`id IN ?`, toUpdate.unbindResources).
		Find(&unbindResources).Error; err != nil {
		return err
	}
	for i := range unbindResources {
		unbindResources[i].Unbind()
		unbindResources[i].UpdatedBy = updatedBy
	}
	// #endregion get unbind resources

	// #region update to db
	if err := tx.updateSiteContents(toUpdate.siteContents, updatedBy); err != nil {
		return err
	}

	if err := tx.maybeUpdateToolResources(toUpdate.bindResources); err != nil {
		return err
	}

	if err := tx.maybeUpdateToolResources(unbindResources); err != nil {
		return err
	}
	// #endregion update to db
	return tx.Commit()
}

// ToolResourceBindV2 implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ToolResourceBindV2(ctx context.Context, req mcom.ToolResourceBindRequestV2) error {
	if err := req.CheckInsufficiency(); err != nil {
		return err
	}

	updatedBy := commonsCtx.UserID(ctx)

	session := dm.newSession(ctx)
	tx, cancel := session.beginTx().withTimeout()
	defer cancel()
	defer tx.Rollback() // nolint: errcheck

	// #region get site information

	stationSites, err := tx.getToBindSites(req)
	if err != nil {
		return err
	}

	sitesInfo := make(map[models.UniqueSite]mcom.ListStationSite)
	for _, site := range stationSites {
		sitesInfo[site.Information.UniqueSite] = site
	}

	getSitesInfo := func(id models.UniqueSite) (mcom.ListStationSite, error) {
		res, ok := sitesInfo[id]
		if !ok {
			return mcom.ListStationSite{}, mcomErr.Error{Code: mcomErr.Code_STATION_SITE_NOT_FOUND}
		}
		if res.Information.Type == sites.Type_SLOT && res.Information.SubType == sites.SubType_TOOL {
			return res, nil
		}
		return mcom.ListStationSite{},
			mcomErr.Error{
				Code:    mcomErr.Code_STATION_SITE_SUB_TYPE_MISMATCH,
				Details: fmt.Sprintf("illegal site type or sub type, station: %s, site name: %s, site index: %d", id.Station, id.SiteID.Name, id.SiteID.Index),
			}
	}
	// #endregion get site information

	// #region bind
	toUpdate := struct {
		siteContents    map[models.UniqueSite]models.SiteContent
		bindResources   []models.ToolResource
		unbindResources []string
	}{
		siteContents: make(map[models.UniqueSite]models.SiteContent),
	}

	for _, detail := range req.Details {
		site, err := getSitesInfo(detail.Site)
		if err != nil {
			return err
		}

		if detail.Type == bindtype.BindType_RESOURCE_BINDING_SLOT_BIND {
			if err := detail.Resource.Require(); err != nil {
				return err
			}

			resource := detail.Resource.ToModelResource()

			unbindResource := site.Content.Slot.Bind(models.BoundResource{
				Tool: &models.ToolSite{
					ResourceID:    resource.ID,
					ToolID:        resource.ToolID,
					InstalledTime: time.Now(),
				},
			})
			toUpdate.siteContents[models.UniqueSite{
				SiteID:  site.Information.SiteID,
				Station: site.Information.Station,
			}] = site.Content

			resource.BindTo(models.UniqueSite{
				SiteID:  site.Information.SiteID,
				Station: site.Information.Station,
			})
			resource.UpdatedBy = updatedBy
			toUpdate.bindResources = append(toUpdate.bindResources, resource)

			if unbindResource.Tool != nil {
				toUpdate.unbindResources = append(toUpdate.unbindResources, unbindResource.Tool.ResourceID)
			}
		} else { // clear
			if err := detail.Resource.Empty(); err != nil {
				return err
			}
			unbindResource := site.Content.Slot.Clear()
			toUpdate.siteContents[models.UniqueSite{
				SiteID:  site.Information.SiteID,
				Station: site.Information.Station,
			}] = site.Content

			if unbindResource.Tool != nil {
				toUpdate.unbindResources = append(toUpdate.unbindResources, unbindResource.Tool.ResourceID)
			}
		}
	}
	// #endregion bind

	// #region get unbind resources
	var unbindResources []models.ToolResource
	if err := tx.db.Model(&models.ToolResource{}).
		Where(`id IN ?`, toUpdate.unbindResources).
		Find(&unbindResources).Error; err != nil {
		return err
	}
	for i := range unbindResources {
		unbindResources[i].Unbind()
		unbindResources[i].UpdatedBy = updatedBy
	}
	// #endregion get unbind resources

	// #region update to db
	if err := tx.updateSiteContents(toUpdate.siteContents, updatedBy); err != nil {
		return err
	}

	if err := tx.maybeUpdateToolResources(toUpdate.bindResources); err != nil {
		return err
	}

	if err := tx.maybeUpdateToolResources(unbindResources); err != nil {
		return err
	}
	// #endregion update to db
	return tx.Commit()
}

func (tx *txDataManager) maybeUpdateToolResources(resources []models.ToolResource) error {
	for _, resource := range resources {
		if err := tx.db.Model(&models.ToolResource{}).
			Where(&models.ToolResource{ID: resource.ID}).
			Select("BindingSite", "UpdatedBy").
			Updates(models.ToolResource{
				BindingSite: resource.BindingSite,
				UpdatedBy:   resource.UpdatedBy,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

// ListToolResources implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) ListToolResources(ctx context.Context, req mcom.ListToolResourcesRequest) (mcom.ListToolResourcesReply, error) {
	if err := req.CheckInsufficiency(); err != nil {
		return mcom.ListToolResourcesReply{}, err
	}

	session := dm.newSession(ctx)
	var resources []models.ToolResource
	if err := session.db.Where(`id IN ?`, req.ResourcesID).Find(&resources).Error; err != nil {
		return mcom.ListToolResourcesReply{}, err
	}

	res := make(map[string]mcom.GetToolResourceReply, len(resources))
	for _, resource := range resources {
		res[resource.ID] = mcom.GetToolResourceReply{
			ToolID:      resource.ToolID,
			BindingSite: resource.BindingSite,
			CreatedBy:   resource.CreatedBy,
			CreatedAt:   resource.CreatedAt,
		}
	}

	return mcom.ListToolResourcesReply{
		Resources: res,
	}, nil
}
