package patches

import (
	"context"
	"io"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl"
)

type DBConnection struct {
	Name     string `yaml:"name"`
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	Schema   string `yaml:"schema"`
}

func DecodeDBConnectionYaml(file io.Reader) (DBConnection, error) {
	var connectionConfigs DBConnection
	if err := yaml.NewDecoder(file).Decode(&connectionConfigs); err != nil {
		return DBConnection{}, err
	}
	return connectionConfigs, nil
}

func (dbc DBConnection) ToDataManager(ctx context.Context) (mcom.DataManager, error) {
	cfg := impl.PGConfig{
		Address:  dbc.Address,
		Port:     dbc.Port,
		UserName: dbc.UserName,
		Password: dbc.Password,
		Database: dbc.Name,
	}

	return impl.New(ctx, cfg, impl.WithPostgreSQLSchema(dbc.Schema), impl.AutoMigrateTables())
}

func (dbc DBConnection) ToGormDB() (*gorm.DB, error) {
	cfg := impl.PGConfig{
		Address:  dbc.Address,
		Port:     dbc.Port,
		UserName: dbc.UserName,
		Password: dbc.Password,
		Database: dbc.Name,
	}

	return impl.NewDB(cfg, impl.DBOptions{
		Schema: dbc.Schema,
		Logger: (&impl.Logger{}).LogMode(logger.Error),
	})
}
