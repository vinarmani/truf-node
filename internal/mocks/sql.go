package mocks

import "github.com/kwilteam/kwil-db/common/sql"

type DB interface {
	sql.Executor
	sql.TxMaker
	// AccessMode gets the access mode of the database.
	// It can be either read-write or read-only.
	AccessMode() sql.AccessMode
}
