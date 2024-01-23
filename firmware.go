package main

import (
	"crypto/md5"
	"fmt"
)

type FirmwareService struct {
	db   *DB
	bins *BinariesService
}

type Md5DiffersError struct {
	given    string
	computed string
}

func (e *Md5DiffersError) Error() string {
	return fmt.Sprintf("MD5 %s (given) != %s (computed)", e.given, e.computed)
}

func (svc *FirmwareService) AddFirmware(info *FirmwareInfo, bytes []byte) (*FirmwareInfo, error) {
	// TODO: AES encryption.
	h := md5.New()
	h.Write([]byte(bytes))
	hash := fmt.Sprintf("%x", h.Sum(nil))
	if info.Md5 == "" {
		info.Md5 = hash
	} else if hash != info.Md5 {
		return nil, &Md5DiffersError{given: info.Md5, computed: hash}
	}

	info.Size = len(bytes)

	addedInfo, err := svc.db.AddFirmwareInfo(info)
	if err != nil {
		return nil, err
	}

	if err := svc.bins.AddFirmwareBinary(addedInfo.Id, bytes); err != nil {
		// TODO: remove record from db.
		return nil, err
	}

	return addedInfo, nil
}

func (serv *FirmwareService) GetLatestFirmware(repo string, board string) (*FirmwareInfo, error) {
	return serv.db.GetLatestFirmwareInfo(repo, board)
}

func (serv *FirmwareService) GetFirmwareBinaryPath(firmwareId int64) (string, error) {
	fi, err := serv.db.GetFirmareInfoById(firmwareId)
	if err != nil {
		return "", err
	}

	if fi == nil {
		return "", nil
	}

	return serv.bins.GetFirmwareBinaryPath(firmwareId), nil
}

func (serv *FirmwareService) GetAllFirmwaresInfo() ([]FirmwareInfo, error) {
	return serv.db.GetAllFirmwaresInfo()
}
