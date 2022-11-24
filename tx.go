package mcom

// TxDataManager is a transaction DataManager with context.
//
// Notice that you should always call dm.beginTx() method to get the TxDataManager entity.
type TxDataManager interface {
	// Commit commits a transaction.
	Commit() error

	// Rollback rollbacks a transaction.
	Rollback() error
}
