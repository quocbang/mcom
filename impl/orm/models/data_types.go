package models

import (
	"context"
	"crypto/sha256"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EncryptedData is an SHA256 hash data.
type EncryptedData []byte

// GormValue implements gorm.Valuer interface.
func (data EncryptedData) GormValue(context.Context, *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL:  "?",
		Vars: []interface{}{Encrypt(data)},
	}
}

// Encrypt encrypts the given data.
func Encrypt(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}
