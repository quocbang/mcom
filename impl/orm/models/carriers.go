package models

import (
	"strings"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"gitlab.kenda.com.tw/kenda/mcom/utils/types"
)

// Carrier is carrier definition. The ID is the IDPrefix+SerialNumber.
type Carrier struct {
	IDPrefix     string `gorm:"type:varchar(2) REFERENCES carrier_serial(carrier_id_prefix);primaryKey"`
	SerialNumber int32  `gorm:"type:integer;primaryKey"`
	// DepartmentOID relative to Department.OID shows which department the carrier belongs to.
	DepartmentOID   string `gorm:"column:department_oid;type:text;not null"`
	AllowedMaterial string `gorm:"type:varchar(20);not null"`
	Deprecated      bool   `gorm:"default:false;not null"`
	// Contents relative to MaterialResource.ID.
	Contents  pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	UpdatedAt types.TimeNano `gorm:"autoUpdateTime:nano;not null"`
	UpdatedBy string         `gorm:"type:text;not null"`
	CreatedAt types.TimeNano `gorm:"autoCreateTime:nano;not null"`
	CreatedBy string         `gorm:"type:text;not null"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (Carrier) TableName() string {
	return "carrier"
}

// MigrateFunction implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
// Function interface.
func (*Carrier) MigrateFunction(db *gorm.DB) error {
	sql := strings.Join([]string{
		"CREATE OR REPLACE FUNCTION fill_in_carrier_serial_number() RETURNS trigger",
		"    LANGUAGE plpgsql",
		"    AS $$",
		"begin",
		"  NEW.serial_number := nextval('carrier_seq_' || FORMAT('%s_%s',ASCII(SUBSTRING(NEW.id_prefix,1,1)),ASCII(SUBSTRING(NEW.id_prefix,2,2))));",
		"  RETURN NEW;",
		"end",
		"$$;",
	}, "\n")
	return db.Exec(sql).Error
}

// MigrateTrigger implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
// Trigger interface.
func (*Carrier) MigrateTrigger(db *gorm.DB) error {
	if err := db.Exec("DROP TRIGGER IF EXISTS fill_in_carrier_serial_number on carrier;").Error; err != nil {
		return err
	}
	return db.Exec("CREATE TRIGGER fill_in_carrier_serial_number BEFORE INSERT ON carrier FOR EACH ROW EXECUTE PROCEDURE fill_in_carrier_serial_number();").Error
}

// CarrierSerial definition.
type CarrierSerial struct {
	CarrierIDPrefix string `gorm:"type:varchar(2);primaryKey"`
}

// TableName implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models" Model interface.
func (CarrierSerial) TableName() string {
	return "carrier_serial"
}

// MigrateFunction implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
// Function interface.
func (*CarrierSerial) MigrateFunction(db *gorm.DB) error {
	sql := strings.Join([]string{
		"CREATE OR REPLACE FUNCTION build_carrier_serial() RETURNS trigger",
		"    LANGUAGE plpgsql",
		"    AS $$",
		"begin",
		"  execute format('CREATE SEQUENCE carrier_seq_%s_%s MAXVALUE 9999',ASCII(SUBSTRING(NEW.carrier_id_prefix,1,1)),ASCII(SUBSTRING(NEW.carrier_id_prefix,2,2)));",
		"  return NEW;",
		"end",
		"$$;",
	}, "\n")
	return db.Exec(sql).Error
}

// MigrateTrigger implements "gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
// Trigger interface.
func (*CarrierSerial) MigrateTrigger(db *gorm.DB) error {
	if err := db.Exec("DROP TRIGGER IF EXISTS build_carrier_serial on carrier_serial;").Error; err != nil {
		return err
	}
	return db.Exec("CREATE TRIGGER build_carrier_serial AFTER INSERT ON carrier_serial FOR EACH ROW EXECUTE PROCEDURE build_carrier_serial();").Error
}
