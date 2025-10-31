package mysql

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-core-fx/sqlfx"
	_ "github.com/go-sql-driver/mysql" // MySQL Driver
)

type wrapper struct {
	sql *sql.DB
}

func (d *wrapper) Dialect() string {
	return "mysql"
}

func (d *wrapper) SQL() *sql.DB {
	return d.sql
}

var _ sqlfx.DB = (*wrapper)(nil)

func New(cfg sqlfx.Config) (sqlfx.DB, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	if u.Scheme != "mysql" && u.Scheme != "mariadb" {
		return nil, fmt.Errorf("%w: unsupported scheme: %s", sqlfx.ErrInvalidConfig, u.Scheme)
	}

	password, ok := u.User.Password()
	if !ok {
		return nil, fmt.Errorf("%w: missing password", sqlfx.ErrInvalidConfig)
	}

	db, err := sql.Open(
		"mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
			u.User.Username(),
			password,
			u.Host,
			strings.TrimPrefix(u.Path, "/"),
			u.RawQuery,
		),
	)

	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}

	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return &wrapper{sql: db}, nil
}
