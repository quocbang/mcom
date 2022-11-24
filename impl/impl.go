package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // register postgresql driver.
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	commonsAccount "gitlab.kenda.com.tw/kenda/commons/v2/util/account"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
	"gitlab.kenda.com.tw/kenda/mcom/impl/pda"
)

// PGConfig is connection configuration for postgreSQL.
type PGConfig struct {
	Address  string
	Port     int
	UserName string
	Password string
	Database string

	// default 1 sec.
	LockTimeout time.Duration
}

// ADConfig is a config for active-directory.
type ADConfig struct {
	Host          string
	Port          int
	DN            string
	QueryUser     string
	QueryPassword string
	WithTLS       bool
}

type options struct {
	schema             string
	pdaServiceEndpoint string
	autoMigrateTables  bool
	migrateCloudTables bool
	adAuth             bool
	adConfig           ADConfig
}

func parseOptions(opts []Option) options {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// Option definition.
type Option func(*options)

// WithPostgreSQLSchema with given schema name for postgreSQL.
func WithPostgreSQLSchema(schema string) Option {
	return func(o *options) {
		o.schema = schema
	}
}

// WithPDAWebServiceEndpoint with given endpoint for PDA web service.
func WithPDAWebServiceEndpoint(endpoint string) Option {
	return func(o *options) {
		o.pdaServiceEndpoint = endpoint
	}
}

// AutoMigrateTables make the system migrate tables automatically.
func AutoMigrateTables() Option {
	return func(o *options) {
		o.autoMigrateTables = true
	}
}

func MigrateCloudTables() Option {
	return func(o *options) {
		o.migrateCloudTables = true
	}
}

// ADAuth for Active Directory Authorization.
func ADAuth(adConfig ADConfig) Option {
	return func(o *options) {
		o.adAuth = true
		o.adConfig = adConfig
	}
}

// DataManager definition.
type DataManager struct {
	db *gorm.DB

	pdaService pda.WebService

	agent *commonsAccount.ADAgent

	lockTimeout time.Duration
}

func newDataManager(cfg PGConfig, o options) (*DataManager, error) {
	db, err := NewDB(cfg, DBOptions{
		Schema: o.schema,
		Logger: newLogger(),
	})
	if err != nil {
		return nil, err
	}
	return &DataManager{db: db, lockTimeout: cfg.LockTimeout}, nil
}

// New creates a new data manager instance to access data.
func New(
	ctx context.Context,
	pgConfig PGConfig,
	opts ...Option,
) (mcom.DataManager, error) {
	o := parseOptions(opts)

	// #region set default timeout
	if pgConfig.LockTimeout == 0 {
		pgConfig.LockTimeout = time.Second
	}
	// #endregion set default timeout

	dm, err := newDataManager(pgConfig, o)
	if err != nil {
		return nil, err
	}

	if o.autoMigrateTables {
		if err := dm.maybeMigrate(o.migrateCloudTables); err != nil {
			return nil, err
		}
	}

	dm.pdaService = pda.NewWebService(o.pdaServiceEndpoint)

	if o.adAuth {
		agent, err := commonsAccount.NewADAgent(commonsAccount.ADConfig{
			Host:          o.adConfig.Host,
			Port:          o.adConfig.Port,
			DN:            o.adConfig.DN,
			QueryUser:     o.adConfig.QueryUser,
			QueryPassword: o.adConfig.QueryPassword,
			WithTLS:       o.adConfig.WithTLS,
		})
		if err != nil {
			return nil, err
		}
		dm.agent = agent
	}

	return dm, nil
}

func (dm *DataManager) maybeMigrate(migrateCloudTable bool) error {
	ms := models.GetModelList()
	if migrateCloudTable {
		cms := models.GetCloudModelList()
		ms = append(ms, cms...)
	}

	if err := maybeMigrateTables(dm.db, ms...); err != nil {
		return err
	}
	if err := maybeMigrateFunctions(dm.db, ms...); err != nil {
		return err
	}
	return maybeMigrateTriggers(dm.db, ms...)
}

// maybeMigrateTables attempts to create tables automatically if implement
// models.Model interface.
func maybeMigrateTables(db *gorm.DB, ms ...models.Model) error {
	dst := []interface{}{}
	for _, m := range ms {
		dst = append(dst, m)
	}
	return db.AutoMigrate(dst...)
}

// maybeMigrateFunctions attempts to create functions automatically if implement
// models.Function interface.
func maybeMigrateFunctions(db *gorm.DB, ms ...models.Model) error {
	for _, m := range ms {
		f, ok := m.(models.Function)
		if !ok {
			continue
		}
		if err := f.MigrateFunction(db); err != nil {
			return err
		}
	}
	return nil
}

// maybeMigrateTriggers attempts to create triggers automatically if implement
// models.Trigger interface.
func maybeMigrateTriggers(db *gorm.DB, ms ...models.Model) error {
	for _, m := range ms {
		t, ok := m.(models.Trigger)
		if !ok {
			continue
		}
		if err := t.MigrateTrigger(db); err != nil {
			return err
		}
	}
	return nil
}

// Close implements gitlab.kenda.com.tw/kenda/mcom DataManager interface.
func (dm *DataManager) Close() error {
	db, err := dm.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

// session is a DataManager struct with context.
type session struct {
	DataManager
	ctx context.Context
}

// newSession returns a new session.
func (dm *DataManager) newSession(ctx context.Context) session {
	var res session
	res.agent = dm.agent
	res.db = dm.db.WithContext(ctx)
	res.pdaService = dm.pdaService
	res.lockTimeout = dm.lockTimeout
	res.ctx = ctx
	return res
}

// beginTx begins a transaction.
func (session *session) beginTx(opts ...*sql.TxOptions) *txDataManager {
	return &txDataManager{
		db:          session.db.Begin(opts...),
		lockTimeout: session.lockTimeout,
		ctx:         session.ctx,
	}
}

// txDataManager is a transaction DataManager. It is NOT supported goroutine.
type txDataManager struct {
	db          *gorm.DB
	lockTimeout time.Duration
	ctx         context.Context
}

// beginTx begins a transaction.
func (dm *DataManager) beginTx(ctx context.Context, opts ...*sql.TxOptions) *txDataManager {
	return &txDataManager{
		db:          dm.db.WithContext(ctx).Begin(opts...),
		lockTimeout: dm.lockTimeout,
		ctx:         ctx,
	}
}

func (tx txDataManager) withTimeout() (txWithTimeout txDataManager, cancel context.CancelFunc) {
	txWithTimeout = tx
	txWithTimeout.ctx, cancel = context.WithTimeout(tx.ctx, tx.lockTimeout)
	txWithTimeout.db = txWithTimeout.db.WithContext(txWithTimeout.ctx)
	return
}

// Commit commits a transaction.
func (tx *txDataManager) Commit() error {
	return tx.db.Commit().Error
}

// Rollback rollbacks a transaction.
func (tx *txDataManager) Rollback() error {
	return tx.db.Rollback().Error
}

// DBOptions are option settings for gorm database.
type DBOptions struct {
	Schema string
	Logger logger.Interface
}

// NewDB creates a new gorm DB.
func NewDB(cfg PGConfig, opts DBOptions) (*gorm.DB, error) {
	connectionString := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		cfg.Address,
		cfg.Port,
		cfg.UserName,
		cfg.Database,
		cfg.Password,
	)
	if opts.Schema != "" {
		connectionString += fmt.Sprintf(" search_path=%s", opts.Schema)
	}
	db, err := gorm.Open(postgres.New(
		postgres.Config{
			DriverName: "postgres",
			DSN:        connectionString,
		},
	), &gorm.Config{
		Logger: opts.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgreSQL database: %v", err)
	}
	return db, nil
}

type whereConditionDelegate func(*gorm.DB) *gorm.DB

// listHandler SQL order clause if req implements mcom.Orderable interface
// and handles SQL limit, offset clause if req implements mcom.Sliceable
// interface.
//
// CAUTION: before setting db as request parameter, please call db.Model()
// and db.Where() if any to make sure the total of the data is correct
// for pagination needs.
//
func listHandler[m models.Model](session *session, req mcom.Request, condition whereConditionDelegate) (dataCount int64, outModels []m, err error) {
	if err := req.CheckInsufficiency(); err != nil {
		return 0, []m{}, err
	}
	db := session.db
	db = condition(db)
	if o, ok := req.(mcom.Orderable); ok {
		if err := o.ValidateOrder(); err != nil {
			return 0, []m{}, err
		}

		for _, oClause := range o.GetOrder() {
			db = db.Order(clause.OrderByColumn{
				Column: clause.Column{Name: oClause.Name},
				Desc:   oClause.Descending,
			})
		}
	}

	if p, ok := req.(mcom.Sliceable); ok && p.NeedPagination() {
		if err := p.ValidatePagination(); err != nil {
			return 0, []m{}, err
		}

		if err := db.Count(&dataCount).Error; err != nil {
			return 0, []m{}, fmt.Errorf("failed to get the total amount of the data: %v", err)
		}

		db = db.Limit(p.GetLimit()).Offset(p.GetOffset())
	}

	if err := db.Find(&outModels).Error; err != nil {
		return 0, []m{}, err
	}

	return
}
