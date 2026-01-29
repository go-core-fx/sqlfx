package sqlfx

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
)

func New(cfg Config) (*sql.DB, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	var driverName, dsn string
	switch u.Scheme {
	case "postgres", "postgresql":
		driverName = "postgres"
		dsn, err = makePostgresDSN(u)
		if err != nil {
			return nil, err
		}
	case "mysql", "mariadb":
		driverName = "mysql"
		dsn, err = makeMySQLDSN(u)
		if err != nil {
			return nil, err
		}
	case "sqlite3", "sqlite":
		driverName = u.Scheme
		// Per RFC 8089: sqlite3:///abs/path uses triple-slash for absolute paths
		// Preserve the leading slash for absolute paths; only strip for relative host-based paths
		if u.Host != "" && u.Host != "localhost" {
			// Relative path: host indicates relative (e.g., "." in sqlite3://./rel/path)
			dsn = u.Host + u.Path
		} else {
			// Absolute path (u.Path already has leading /) or opaque path (no leading /)
			dsn = u.Path
		}
		if raw := u.RawQuery; raw != "" {
			dsn = fmt.Sprintf("%s?%s", dsn, raw)
		}
	default:
		return nil, fmt.Errorf("%w: unsupported scheme: %s", ErrInvalidConfig, u.Scheme)
	}

	db, err := sql.Open(
		driverName,
		dsn,
	)

	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}

	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}

func escapePostgresValue(value string) string {
	if strings.ContainsAny(value, " \t\n\r\f\v'\\") {
		// Escape single quotes and backslashes
		value = strings.ReplaceAll(value, "\\", "\\\\")
		value = strings.ReplaceAll(value, "'", "''")
		return "'" + value + "'"
	}
	return value
}
func parseUsernamePassword(user *url.Userinfo) (string, string, error) {
	if user == nil {
		return "", "", fmt.Errorf("%w: missing username and password", ErrInvalidConfig)
	}

	password, _ := user.Password()
	if password == "" {
		return "", "", fmt.Errorf("%w: missing password", ErrInvalidConfig)
	}

	return user.Username(), password, nil
}

func makePostgresDSN(u *url.URL) (string, error) {
	username, password, err := parseUsernamePassword(u.User)
	if err != nil {
		return "", err
	}

	port := u.Port()
	if port == "" {
		port = "5432"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		u.Hostname(),
		port,
		escapePostgresValue(username),
		escapePostgresValue(password),
		escapePostgresValue(strings.TrimPrefix(u.Path, "/")),
	)
	queryBuilder := strings.Builder{}
	for key, values := range u.Query() {
		for _, value := range values {
			queryBuilder.WriteString(" ")
			queryBuilder.WriteString(key)
			queryBuilder.WriteString("=")
			queryBuilder.WriteString(escapePostgresValue(value))
		}
	}
	dsn += queryBuilder.String()

	return dsn, nil
}

func makeMySQLDSN(u *url.URL) (string, error) {
	username, password, err := parseUsernamePassword(u.User)
	if err != nil {
		return "", err
	}

	port := u.Port()
	if port == "" {
		port = "3306"
	}

	encodedUser := url.PathEscape(username)
	encodedPass := url.PathEscape(password)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		encodedUser,
		encodedPass,
		u.Hostname(),
		port,
		strings.TrimPrefix(u.Path, "/"),
	)

	params := u.Query()
	// Apply defaults only if not already set
	if params.Get("charset") == "" {
		params.Set("charset", "utf8mb4")
	}
	if params.Get("parseTime") == "" {
		params.Set("parseTime", "True")
	}
	if params.Get("loc") == "" {
		params.Set("loc", "Local")
	}
	return fmt.Sprintf("%s?%s", dsn, params.Encode()), nil
}
