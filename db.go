package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"path/filepath"
	"time"
)

type FirmwareInfo struct {
	id       int64
	repoName string
	commitId string
	tag      string
	builtAt  time.Time
	loadedAt time.Time
	loadedBy string
	sha256   string
}

type DB struct {
	*sql.DB
}

const SQLITE_DB_FILENAME = "storage/firmware.db"

func (db *DB) createFirmwareTable() error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS firmwares (
	    id         BIGINT PRIMARY KEY,
	    repoName   TEXT NOT NULL,
	    commitId   TEXT NOT NULL,
	    tag        TEXT NOT NULL,
	    builtAt    DATETIME NOT NULL,
	    loadedAt   DATETIME NOT NULL,
	    loadedBy   TEXT NOT NULL,
        sha256     TEXT NOT NULL
	);`)

	return err
}

func NewDB(cfg *Config) (*DB, error) {
	path := filepath.Join(cfg.storagePath, SQLITE_DB_FILENAME)
	_db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	db := &DB{_db}
	err = db.createFirmwareTable()

	return db, err
}

func (db *DB) AddFirmwareInfo(info *FirmwareInfo) (int64, error) {
	stmt, err := db.Prepare(`
    INSERT INTO firmwares (
        repoName,
        commitId,
        tag,
        builtAt,
        loadedAt,
        loadedBy,
        sha256
    ) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}

	result, err := stmt.Exec(
		info.repoName,
		info.commitId,
		info.tag,
		info.builtAt,
		info.loadedAt,
		info.loadedBy,
		info.sha256,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func firmwareInfoFromSqlRows(rows *sql.Rows) (*FirmwareInfo, error) {
	var fi FirmwareInfo
	if err := rows.Scan(
		&fi.id,
		&fi.repoName,
		&fi.commitId,
		&fi.tag,
		&fi.builtAt,
		&fi.loadedAt,
		&fi.loadedBy,
		&fi.sha256,
	); err != nil {
		return nil, err
	}

	return &fi, nil
}

func (db *DB) GetNewestFirmwareInfo(repo string, tags []string) (*FirmwareInfo, error) {
	query := "SELECT * FROM firmwares WHERE repoName = ?"
	values := [](any){repo}
	if len(tags) > 0 {
		query += " AND tag IN ("
		for i := 0; i < len(tags)-1; i++ {
			query += "?, "
			values = append(values, tags[i])
		}
		query += "?)"
		values = append(values, tags[len(tags)-1])
	}
	query += " ORDER BY loadedAt DESC LIMIT 1;"
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(values...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}

	return firmwareInfoFromSqlRows(rows)
}

func (db *DB) GetAllFirmwares() ([]FirmwareInfo, error) {
	stmt, err := db.Prepare("SELECT * FROM firmwares;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	var fis []FirmwareInfo
	for rows.Next() {
		fi, err := firmwareInfoFromSqlRows(rows)
		if err != nil {
			return nil, err
		}
		fis = append(fis, *fi)
	}

	return fis, nil
}
