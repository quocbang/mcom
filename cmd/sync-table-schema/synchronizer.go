package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom/impl"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

type options struct {
	logger *zap.Logger
}

type option func(*options)

func withLogger(logger *zap.Logger) option {
	return func(o *options) { o.logger = logger }
}

type dbSynchronizer struct {
	db             *gorm.DB
	logger         *zap.Logger
	migrateHandler func(*zap.Logger) error

	source, dest string // schema.
}

func newDBSynchronizer(config dbConfig, opts ...option) (*dbSynchronizer, error) {
	db, err := impl.NewDB(impl.PGConfig{
		Address:  config.Address,
		Port:     config.Port,
		UserName: config.UserName,
		Password: config.Password,
		Database: config.DatabaseName,
	}, impl.DBOptions{
		Schema: config.Schema,
	})
	if err != nil {
		return nil, err
	}

	s := &dbSynchronizer{
		db:     db,
		logger: zap.NewNop(),
		migrateHandler: func(logger *zap.Logger) error {
			dm, err := impl.New(context.Background(), impl.PGConfig{
				Address:  config.Address,
				Port:     config.Port,
				UserName: config.UserName,
				Password: config.Password,
				Database: config.DatabaseName,
			}, impl.WithPostgreSQLSchema(config.Schema), impl.AutoMigrateTables())
			if err != nil {
				logger.Error("failed to migrate tables", zap.String("schema", config.Schema), zap.Error(err))
				return err
			}
			logger.Info("migrate tables successful", zap.String("schema", config.Schema))
			dm.Close()
			return nil
		},
		source: config.SourceData,
		dest:   config.Schema,
	}

	var o options
	for _, opt := range opts {
		opt(&o)
	}

	if o.logger != nil {
		s.logger = o.logger
	}
	return s, nil
}

func (s *dbSynchronizer) Close() error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (s *dbSynchronizer) MaybeDropDestinationSchema() error {
	if s.source == "" {
		return nil
	}

	if err := s.db.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", s.dest)).Error; err != nil {
		s.logger.Error("failed to drop schema", zap.String("schema", s.dest), zap.Error(err))
		return err
	}
	s.logger.Info("drop schema successful", zap.String("schema", s.dest))
	return nil
}

func (s *dbSynchronizer) MaybeCreateDestinationSchema() error {
	if s.source == "" {
		return nil
	}

	if err := s.db.Exec(fmt.Sprintf("CREATE SCHEMA %s", s.dest)).Error; err != nil {
		s.logger.Error("failed to create schema", zap.String("schema", s.dest), zap.Error(err))
		return err
	}
	s.logger.Info("create schema successful", zap.String("schema", s.dest))
	return nil
}

func (s *dbSynchronizer) MigrateTables() error {
	return s.migrateHandler(s.logger)
}

// SyncTablesData evan if there is an error, it tries to synchronize.
// the data we expect.
func (s *dbSynchronizer) MaybeSyncTablesData() []error {
	if s.source == "" {
		return nil
	}

	ms := models.GetModelList()

	errs := []error{}
	for _, m := range ms {
		if err := s.syncTableData(m.TableName()); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (s *dbSynchronizer) syncTableData(table string) error {
	if err := s.db.Exec(fmt.Sprintf(`INSERT INTO "%s"."%s" SELECT * FROM "%s"."%s"`, s.dest, table, s.source, table)).Error; err != nil {
		s.logger.Error("failed to synchronize the data", zap.String("table", table), zap.Error(err))
		return err
	}
	s.logger.Info("synchronize the data successful", zap.String("table", table))
	return nil
}
