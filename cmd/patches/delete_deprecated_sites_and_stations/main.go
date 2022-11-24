package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	patches "gitlab.kenda.com.tw/kenda/mcom/cmd/patches/common"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/utils/stations"
	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

func main() {
	dbConnectionPath := flag.String("dbConnection", "", "path of db connection parameters yaml file")
	flag.Parse()

	file, err := os.Open(*dbConnectionPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	dbConfigs, err := patches.DecodeDBConnectionYaml(file)
	if err != nil {
		log.Fatal(err)
	}

	dropStationColumn := fmt.Sprintf("ALTER TABLE %s.station DROP COLUMN deprecated;", dbConfigs.Schema)
	dropSiteColumn := fmt.Sprintf("ALTER TABLE %s.site DROP COLUMN deprecated;", dbConfigs.Schema)

	db, err := dbConfigs.ToGormDB()
	if err != nil {
		log.Fatal(err)
	}

	// #region log context
	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	ctx = commonsCtx.WithLogger(ctx, logger)
	// #endregion log context

	if err := db.WithContext(ctx).Model(&OldStation{}).Where(OldStation{Deprecated: true}).Delete(&OldStation{}).Error; err != nil {
		log.Fatal(err)
	}
	if err := db.WithContext(ctx).Model(&OldSite{}).Where(OldSite{Deprecated: true}).Delete(&OldSite{}).Error; err != nil {
		log.Fatal(err)
	}

	if err := db.WithContext(ctx).Exec(dropStationColumn).Error; err != nil {
		log.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(dropSiteColumn).Error; err != nil {
		log.Fatal(err)
	}

	patches.PrintDragon()
}

type OldSite struct {
	Name              string                `gorm:"type:varchar(16);primaryKey"`
	Index             int16                 `gorm:"default:0;primaryKey"`
	Station           string                `gorm:"type:varchar(20);primaryKey"`
	AdminDepartmentID string                `gorm:"column:admin_department_id;type:text;not null;index:idx_site_department"`
	Attributes        models.SiteAttributes `gorm:"type:jsonb;default:'{}';not null"`
	Deprecated        bool                  `gorm:"default:false;not null"`
	UpdatedAt         types.TimeNano        `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy         string                `gorm:"type:text;not null"`
	CreatedAt         types.TimeNano        `gorm:"autoCreateTime:nano;not null"`
	CreatedBy         string                `gorm:"type:text;not null"`
}

func (OldSite) TableName() string {
	return "site"
}

type OldStation struct {
	ID                string                    `gorm:"type:varchar(20);primaryKey"`
	AdminDepartmentID string                    `gorm:"column:admin_department_id;type:text;not null;index:idx_station_department"`
	Sites             models.StationSites       `gorm:"type:jsonb;default:'[]';not null"`
	State             stations.State            `gorm:"default:1;not null"`
	Information       models.StationInformation `gorm:"type:json;default:'{}';not null"`
	UpdatedAt         types.TimeNano            `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy         string                    `gorm:"type:text;not null"`
	CreatedAt         types.TimeNano            `gorm:"autoCreateTime:nano;not null"`
	CreatedBy         string                    `gorm:"type:text;not null"`
	Deprecated        bool                      `gorm:"default:false;not null"`
}

func (OldStation) TableName() string {
	return "station"
}
