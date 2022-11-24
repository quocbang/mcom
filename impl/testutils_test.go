package impl

import (
	"context"
	"os"
	"testing"

	"bou.ke/monkey"
	flags "github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	commonsAccount "gitlab.kenda.com.tw/kenda/commons/v2/util/account"

	"gitlab.kenda.com.tw/kenda/mcom"
	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

var testDBOptions struct {
	DBSchema   string `long:"db_schema" description:"the test DB schema" env:"DB_SCHEMA"`
	DBAddress  string `long:"db_address" description:"the test DB address" env:"DB_ADDRESS"`
	DBPort     int    `long:"db_port" description:"the test DB port" env:"DB_PORT"`
	DBUsername string `long:"db_username" description:"the test DB username" env:"DB_USERNAME"`
	DBPassword string `long:"db_password" description:"the test DB password" env:"DB_PASSWORD"`
	DBDatabase string `long:"db_database" description:"the test DB database" env:"DB_DATABASE"`
}

var testADConfigs struct {
	Host          string `long:"ad_host" description:"the test ad agent host" env:"AD_HOST"`
	Port          int    `long:"ad_port" description:"the test ad agent port" env:"AD_PORT"`
	DN            string `long:"ad_dn" description:"the test ad agent dn" env:"AD_DN"`
	QueryUser     string `long:"ad_user" description:"the test ad agent query user" env:"AD_USER"`
	QueryPassword string `long:"ad_password" description:"the test ad agent query password" env:"AD_PASSWORD"`
	WithTLS       bool   `long:"ad_with_tls" description:"the test ad agent with tls flag" env:"AD_WITH_TLS"`
}

func initializeDB(t *testing.T) (ctx context.Context, dm mcom.DataManager, db *gorm.DB) {
	ctx = context.Background()
	dm, db, err := newTestDataManager()
	assert.NoError(t, err, "func error : newTestDataManager")
	assert.NotNil(t, dm, "nil datamanager")
	assert.NotNil(t, db, "nil gorm.DB")
	return
}

func TestMain(m *testing.M) {
	parser := flags.NewParser(&testDBOptions, flags.IgnoreUnknown)
	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			code = 0
		}
		os.Exit(code)
	}

	parser = flags.NewParser(&testADConfigs, flags.IgnoreUnknown)
	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			code = 0
		}
		os.Exit(code)
	}
	os.Exit(m.Run())
}

// NewADAgent will be replace to get empty agent.
func newTestDataManager() (mcom.DataManager, *gorm.DB, error) {
	pgConfig := PGConfig{
		Address:  testDBOptions.DBAddress,
		Port:     testDBOptions.DBPort,
		UserName: testDBOptions.DBUsername,
		Password: testDBOptions.DBPassword,
		Database: testDBOptions.DBDatabase,
	}

	adConfig := ADConfig{
		Host:          testADConfigs.Host,
		Port:          testADConfigs.Port,
		DN:            testADConfigs.DN,
		QueryUser:     testADConfigs.QueryUser,
		QueryPassword: testADConfigs.QueryPassword,
		WithTLS:       testADConfigs.WithTLS,
	}

	monkey.Patch(commonsAccount.NewADAgent, func(cfg commonsAccount.ADConfig) (*commonsAccount.ADAgent, error) {
		return &commonsAccount.ADAgent{}, nil
	})

	dm, err := New(
		context.Background(),
		pgConfig,
		MigrateCloudTables(),
		WithPostgreSQLSchema(testDBOptions.DBSchema),
		AutoMigrateTables(),
		ADAuth(adConfig),
	)
	if err != nil {
		return nil, nil, err
	}
	monkey.Unpatch(commonsAccount.NewADAgent)
	return dm, dm.(*DataManager).db, nil
}

type modelDataClearMaster struct {
	db *gorm.DB
	ms []models.Model
}

func newClearMaster(db *gorm.DB, ms ...models.Model) *modelDataClearMaster {
	return &modelDataClearMaster{
		db: db,
		ms: ms,
	}
}

func (cm *modelDataClearMaster) Clear() error {
	for _, m := range cm.ms {
		if err := cm.db.Where("1=1").Delete(m).Error; err != nil {
			return err
		}
	}
	return nil
}
