package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"path/filepath"
	"time"
)

type FirmwareInfo struct {
	id        int64
	repo_name string
	commit_id string
	tag       string
	built_at  time.Time
	loaded_at time.Time
	loaded_by string
	sha256    string
}

type DB struct {
	*sql.DB
}

const SQLITE_DB_FILENAME = "storage/firmware.db"

func (db *DB) createFirmwareTable() error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS firmwares (
	    id          BIGINT PRIMARY KEY,
	    repo_name   TEXT NOT NULL,
	    commit_id   TEXT NOT NULL,
	    tag         TEXT NOT NULL,
	    built_at    DATETIME NOT NULL,
	    loaded_at   DATETIME NOT NULL,
	    loaded_by   TEXT NOT NULL,
        sha256      TEXT NOT NULL
	);`)

	return err
}

func NewDB() (*DB, error) {
	path := filepath.Join(STORAGE_PATH, SQLITE_DB_FILENAME)
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
        repo_name,
        commit_id,
        tag,
        built_at,
        loaded_at,
        loaded_by,
        sha256
    ) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}

	result, err := stmt.Exec(
		info.repo_name,
		info.commit_id,
		info.tag,
		info.built_at,
		info.loaded_at,
		info.loaded_by,
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
		&fi.repo_name,
		&fi.commit_id,
		&fi.tag,
		&fi.built_at,
		&fi.loaded_at,
		&fi.loaded_by,
		&fi.sha256,
	); err != nil {
		return nil, err
	}

	return &fi, nil
}

func (db *DB) GetNewestLoadedFirmware(repo string, tags []string) (*FirmwareInfo, error) {
	query := "SELECT * FROM firmwares WHERE repo_name = ?"
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
	query += " ORDER BY loaded_at DESC LIMIT 1;"
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
