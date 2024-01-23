package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"path/filepath"
	"time"
)

type FirmwareInfo struct {
	Id          int64
	RepoName    string
	CommitId    string
	Tag         string
	BuiltAt     time.Time
	LoadedAt    time.Time
	LoadedBy    string
	Sha256      string
	Description string
	Size        int
}

type DB struct {
	*sql.DB
}

const SQLITE_DB_FILENAME = "firmware.db"

func (db *DB) createFirmwareTable() error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS firmwares (
	    id          INTEGER PRIMARY KEY AUTOINCREMENT,
	    repoName    TEXT NOT NULL,
	    commitId    TEXT NOT NULL,
	    tag         TEXT NOT NULL,
	    builtAt     DATETIME NOT NULL,
	    loadedAt    DATETIME NOT NULL,
	    loadedBy    TEXT NOT NULL,
        sha256      TEXT NOT NULL,
        description TEXT NOT NULL,
        size        INTEGER NOT NULL
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

func (db *DB) AddFirmwareInfo(info *FirmwareInfo) (*FirmwareInfo, error) {
	stmt, err := db.Prepare(`
    INSERT INTO firmwares (
        repoName,
        commitId,
        tag,
        builtAt,
        loadedAt,
        loadedBy,
        sha256,
        description,
        size
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}

	result, err := stmt.Exec(
		info.RepoName,
		info.CommitId,
		info.Tag,
		info.BuiltAt,
		info.LoadedAt,
		info.LoadedBy,
		info.Sha256,
		info.Description,
        info.Size,
	)
	if err != nil {
		return nil, err
	}

	ret := *info
	ret.Id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func firmwareInfoFromSqlRows(rows *sql.Rows) (*FirmwareInfo, error) {
	var fi FirmwareInfo
	if err := rows.Scan(
		&fi.Id,
		&fi.RepoName,
		&fi.CommitId,
		&fi.Tag,
		&fi.BuiltAt,
		&fi.LoadedAt,
		&fi.LoadedBy,
		&fi.Sha256,
		&fi.Description,
        &fi.Size,
	); err != nil {
		return nil, err
	}

	return &fi, nil
}

func (db *DB) GetLatestFirmwareInfo(repo string, tags []string) (*FirmwareInfo, error) {
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
	query += " ORDER BY builtAt DESC LIMIT 1;"
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

func (db *DB) GetFirmareInfoById(id int64) (*FirmwareInfo, error) {
	stmt, err := db.Prepare("SELECT * FROM firmwares WHERE id=?")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}

	var fi *FirmwareInfo = nil
	if rows.Next() {
		fi, err = firmwareInfoFromSqlRows(rows)
		if err != nil {
			return nil, err
		}
	}

	return fi, nil
}

func (db *DB) GetAllFirmwaresInfo() ([]FirmwareInfo, error) {
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
