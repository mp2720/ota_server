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
	Boards      []string // not presented in firmwares table
	CreatedAt   time.Time
	LoadedBy    string
	Md5         string
	Description string
	Size        int // 0 if no binary file uploaded, empty files are not allowed
}

func (fi *FirmwareInfo) hasBin() bool {
	return fi.Size != 0
}

type FirmwareForBoardRecord struct {
	BoardName  string
	FirmwareId int64
}

type DB struct {
	*sql.DB
}

const SQLITE_DB_FILENAME = "firmware.db"

func (db *DB) createTables() error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS firmwares (
	    id          INTEGER PRIMARY KEY AUTOINCREMENT,
	    repoName    TEXT NOT NULL,
	    commitId    TEXT NOT NULL,
	    builtAt     DATETIME NOT NULL,
	    loadedBy    TEXT NOT NULL,
        md5         TEXT NOT NULL,
        description TEXT NOT NULL,
        size        INTEGER NOT NULL
	);
    CREATE TABLE IF NOT EXISTS boards (
        boardName   TEXT NOT NULL,
        firmwareId  INTEGER NOT NULL
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
	err = db.createTables()

	return db, err
}

func (db *DB) AddFirmwareInfo(info *FirmwareInfo) (*FirmwareInfo, error) {
	stmt, err := db.Prepare(`
    INSERT INTO firmwares (
        repoName,
        commitId,
        builtAt,
        loadedBy,
        md5,
        description,
        size
    ) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		info.RepoName,
		info.CommitId,
		info.CreatedAt,
		info.LoadedBy,
		info.Md5,
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

	// TODO: if firmware was added, but adding boards failed?
	stmt2, err := db.Prepare(`
    INSERT INTO boards (
        boardName,
        firmwareId
    ) VALUES (?, ?)
    `)
	if err != nil {
		return nil, err
	}
	defer stmt2.Close()

	for _, board := range info.Boards {
		_, err := stmt2.Exec(
			board,
			ret.Id,
		)
		if err != nil {
			return nil, err
		}
	}

	return &ret, nil
}

func (db *DB) firmwareInfoFromSqlRows(firmwareRows *sql.Rows) (*FirmwareInfo, error) {
	var fi FirmwareInfo
	if err := firmwareRows.Scan(
		&fi.Id,
		&fi.RepoName,
		&fi.CommitId,
		&fi.CreatedAt,
		&fi.LoadedBy,
		&fi.Md5,
		&fi.Description,
		&fi.Size,
	); err != nil {
		return nil, err
	}

	stmt, err := db.Prepare("SELECT boardName FROM boards where firmwareId = ?;")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	boardRows, err := stmt.Query(fi.Id)
	if err != nil {
		return nil, err
	}
	defer boardRows.Close()

	for boardRows.Next() {
		var board string
		if err := boardRows.Scan(&board); err != nil {
			return nil, err
		}
		fi.Boards = append(fi.Boards, board)
	}

	return &fi, nil
}

func (db *DB) GetLatestFirmwareInfo(repo string, board string) (*FirmwareInfo, error) {
	stmt, err := db.Prepare(`
        SELECT
	    	firmwares.id,
	    	firmwares.repoName,
	    	firmwares.commitId,
	    	firmwares.builtAt,
	    	firmwares.loadedBy,
	    	firmwares.md5,
	    	firmwares.description,
	    	firmwares.size
        FROM boards JOIN firmwares ON firmwares.id = boards.firmwareId
        WHERE
            firmwares.repoName = ?
            AND boards.boardName = ?
            AND firmwares.size != 0
        ORDER BY firmwares.builtAt DESC LIMIT 1;`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(repo, board)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	return db.firmwareInfoFromSqlRows(rows)
}

func (db *DB) GetFirmareInfoById(id int64) (*FirmwareInfo, error) {
	stmt, err := db.Prepare("SELECT * FROM firmwares WHERE id=?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	return db.firmwareInfoFromSqlRows(rows)
}

func (db *DB) GetAllFirmwaresInfo() ([]FirmwareInfo, error) {
	stmt, err := db.Prepare("SELECT * FROM firmwares;")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fis []FirmwareInfo
	for rows.Next() {
		fi, err := db.firmwareInfoFromSqlRows(rows)
		if err != nil {
			return nil, err
		}
		fis = append(fis, *fi)
	}

	return fis, nil
}

func (db *DB) UpdateFirmwareFileInfo(fi *FirmwareInfo) error {
	stmt, err := db.Prepare(`
    UPDATE firmwares
    SET
        md5 = ?,
        size = ?
    WHERE firmwares.id = ?
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		fi.Md5,
		fi.Size,
		fi.Id,
	)
	return err
}
