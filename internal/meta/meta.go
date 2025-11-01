package meta

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func Open(dsn string) (*sql.DB, error) {
	return sql.Open("sqlite", dsn)
}

func Migrate(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS objects(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  k TEXT UNIQUE,               -- 对象key
  size INTEGER,
  etag TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
	_, err := db.Exec(schema)
	return err
}
