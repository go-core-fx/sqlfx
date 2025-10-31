package sqlfx

import "database/sql"

type DB interface {
	Dialect() string
	SQL() *sql.DB
}
